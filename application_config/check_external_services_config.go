package application_config

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_tests/helper"
	"integration_tests/testing"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type CheckExternalServicesConfig struct {
	checkMap     map[string]interface{}
	checkConfigs map[string]ICheckExternalServiceConfig
}

func (c CheckExternalServicesConfig) Check(externalServices map[string]*ExternalServiceDefinitionSection) error {
	variables := make(map[string]interface{})
	saveResult := func(key string, value interface{}) {
		variables[key] = value
	}

	for name, checkConfig := range c.checkConfigs {
		externalService := externalServices[name]
		if externalService == nil {
			return fmt.Errorf("unable to find external service %q", name)
		}
		err := externalService.executor.Check(checkConfig, saveResult, variables)
		if err != nil {
			return fmt.Errorf("unable to check external service %q: %v", name, err)
		}
	}
	return nil
}

type FnResultSaver func(key string, value interface{})

type IPostgresCheck interface {
	Check(conn *sqlx.DB, saveResult FnResultSaver, variables map[string]interface{}) error
}

func NewPostgresChecker(query string, expectedResult interface{}, saveResultTo string) *PostgresChecker {
	return &PostgresChecker{
		query:          query,
		expectedResult: expectedResult,
		saveResultTo:   saveResultTo,
	}
}

type PostgresChecker struct {
	query          string
	expectedResult interface{}
	saveResultTo   string
}

func (pc PostgresChecker) makeRequest(conn *sqlx.DB) ([]interface{}, error) {
	actualResult := make([]interface{}, 0)
	rows, err := conn.Query(pc.query)
	if err != nil {
		return nil, fmt.Errorf("unable to make sql request: %v", err)
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("unable to get result columns: %v", err)
	}
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("unable to scan result: %v", err)
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		actualResult = append(actualResult, m)
	}
	return actualResult, nil
}

var interpolationRegexp = regexp.MustCompile(`{{\s*([\w.]+)\s*}}`)

func interpolate(variableName string, rootValue interface{}) (string, error) {
	// fmt.Printf("interpolate %q %#v\n", variableName, rootValue)
	variableNameParts := strings.SplitN(variableName, ".", 2)

	var newRootValue interface{}
	switch val := rootValue.(type) {
	case map[string]interface{}:
		newRootValue = val[variableNameParts[0]]
	case []interface{}:
		index, err := strconv.Atoi(variableNameParts[0])
		if err != nil {
			return "", fmt.Errorf("cannot parse index %v: %v", variableNameParts[0], err)
		}
		newRootValue = val[index]
	default:
		return "", fmt.Errorf("cannot read value by key %q of %T (%#v): not supported type", variableName, rootValue, rootValue)
	}

	if len(variableNameParts) == 1 {
		return fmt.Sprintf("%v", newRootValue), nil
	}
	return interpolate(variableNameParts[1], newRootValue)
}

func applyInterpolation(input interface{}, variables map[string]interface{}) (interface{}, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal input: %v", err)
	}

	matches := interpolationRegexp.FindAllSubmatchIndex(encoded, -1)
	lenDiff := 0
	for i, m := range matches {
		replaceFromIndex := m[0] + lenDiff
		replaceToIndex := m[1] + lenDiff
		interpolateFromIndex := m[2] + lenDiff
		interpolateToIndex := m[3] + lenDiff
		replWith, err := interpolate(string(encoded[interpolateFromIndex:interpolateToIndex]), variables)
		if err != nil {
			return nil, fmt.Errorf("unable to interpolate match %d: %v", i, err)
		}
		encoded = append(encoded[:replaceFromIndex], append([]byte(replWith), encoded[replaceToIndex:]...)...)
		lenDiff += len(replWith) - (replaceToIndex - replaceFromIndex)
	}

	var output interface{}
	err = json.Unmarshal(encoded, &output)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal result: %v", err)
	}
	return output, nil
}

func (pc PostgresChecker) Check(conn *sqlx.DB, saveResult FnResultSaver, variables map[string]interface{}) error {
	// fmt.Println(">> postgres checker", pc.query)

	actualResult, err := pc.makeRequest(conn)
	if err != nil {
		return fmt.Errorf("unable to make requset to db: %v", err)
	}
	// fmt.Printf("\n!! query: %s\n", pc.query)
	// fmt.Printf("!! actual  : %#v\n", actualResult)

	if pc.expectedResult != nil {
		// fmt.Printf("!! expected: %#v\n", pc.expectedResult)
		// separate variable to be able to convert to slice
		expectedResult := pc.expectedResult
		_, expectingSlice := expectedResult.([]interface{})
		if !expectingSlice {
			expectedResult = []interface{}{expectedResult}
		}
		expectedResult, err = applyInterpolation(expectedResult, variables)
		if err != nil {
			return fmt.Errorf("unable to apply interpolation: %v", err)
		}
		err := testing.IsEqual(actualResult, expectedResult)
		// fmt.Println("!! is equal", err)
		// fmt.Printf("\n")
		if err != nil {
			return fmt.Errorf("db state not valid: %v", err)
		}
	}
	if pc.saveResultTo != "" {
		saveResult(pc.saveResultTo, actualResult)
	}

	return nil
}

func NewPostgresCheckConfig(checkConfig interface{}) (*PostgresCheckConfig, error) {
	checksList, ok := checkConfig.([]interface{})
	if !ok {
		return nil, fmt.Errorf("service check for postgres must be list, but it is %T (%#v)", checkConfig, checkConfig)
	}
	var checkers []IPostgresCheck
	for _, checkInterface := range checksList {
		checkYaml, ok := checkInterface.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("service check for postgres must be map, but it is %T (%#v)", checkInterface, checkInterface)
		}
		check := helper.YamlMapToJsonMap(checkYaml)
		query, ok := check["query"].(string)
		if !ok {
			return nil, fmt.Errorf("service check query for postgres must be string, but it is %T %#v", check["query"], check["query"])
		}
		var expectedResult interface{}
		expectedResultStruct, ok := check["expected_result"].(map[string]interface{})
		if ok {
			expectedResult = expectedResultStruct
		}
		expectedResultSlice, ok := check["expected_result"].([]interface{})
		if ok {
			expectedResult = expectedResultSlice
		}
		saveResultTo, _ := check["save_result_to"].(string)

		checkers = append(checkers, NewPostgresChecker(query, expectedResult, saveResultTo))
	}

	return &PostgresCheckConfig{
		checkers: checkers,
	}, nil
}

type PostgresCheckConfig struct {
	checkers []IPostgresCheck
}

func (pcc PostgresCheckConfig) Check(conn *sqlx.DB, saveResult FnResultSaver, variables map[string]interface{}) error {
	for i, check := range pcc.checkers {
		err := check.Check(conn, saveResult, variables)
		if err != nil {
			return fmt.Errorf("unable to check postgres %d: %v", i, err)
		}
	}
	return nil
}

type IHttpCheck interface {
	Check(serviceUrl string) error
}

func NewCallsHttpCheck(calls []CheckCall) *CallsHttpCheck {
	return &CallsHttpCheck{
		calls: calls,
	}
}

type CallsHttpCheck struct {
	calls []CheckCall
}

func (c CallsHttpCheck) Check(serviceUrl string) error {
	resp, err := http.Get(serviceUrl + "__calls")
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %v", err)
	}
	var actualCalls []ActualCall
	err = json.Unmarshal(respBody, &actualCalls)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response body: %v", err)
	}
	if len(actualCalls) != len(c.calls) {
		return fmt.Errorf("different calls count: actual is %d, expected is %d", len(actualCalls), len(c.calls))
	}
	for i, check := range c.calls {
		actualCall := actualCalls[i]
		var parsedActualBody interface{}
		err := check.unmarshal([]byte(actualCall.Body), &parsedActualBody)
		if err != nil {
			return fmt.Errorf("check #%d failed: unable to parse actual call body: %v", i, err)
		}
		if actualCall.Route != check.Route {
			return fmt.Errorf("check #%d failed: route not matched. expcted %q, actual %q", i, check.Route, actualCall.Route)
		}
		err = testing.IsEqual(parsedActualBody, check.Body)
		if err != nil {
			return fmt.Errorf("check #%d failed: body not matched: %v", i, err)
		}
	}
	return nil
}

type FnUnmarshal = func(data []byte, dst interface{}) error

type CheckCall struct {
	// unmarshal is method based on body content type to unmarshal string body to be comparable with body defined in config
	unmarshal FnUnmarshal

	Route string
	Body  interface{}
}

type ActualCall struct {
	Route string `json:"route"`
	Body  string `json:"body"`
}

func NewHttpCheckConfig(checkConfig interface{}) (*HttpCheckConfig, error) {
	checkYaml, ok := checkConfig.(helper.YamlMap)
	if !ok {
		return nil, fmt.Errorf("http check config should be map")
	}
	var checks []IHttpCheck
	check := helper.YamlMapToJsonMap(checkYaml)
	for action, value := range check {
		switch action {
		case "calls":
			callsArr, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("calls should be array")
			}
			var calls []CheckCall
			for _, callsDefinitionInterface := range callsArr {
				callDefinitionMap, ok := callsDefinitionInterface.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("calls definition should be map")
				}
				for route, callDefinitionInterface := range callDefinitionMap {
					callDefinition, ok := callDefinitionInterface.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("call definition should be map")
					}
					var unmarshal FnUnmarshal
					switch callDefinition["content_type"] {
					case "application/json":
						unmarshal = json.Unmarshal
					default:
						return nil, fmt.Errorf("content_type for body should be defined")
					}
					calls = append(calls, CheckCall{
						unmarshal: unmarshal,
						Route:     route,
						Body:      callDefinition["body"],
					})
				}
			}
			checks = append(checks, NewCallsHttpCheck(calls))
		default:
			return nil, fmt.Errorf("invalid http check action %q", action)
		}
	}
	return &HttpCheckConfig{
		checks: checks,
	}, nil
}

type HttpCheckConfig struct {
	checks []IHttpCheck
}

func (hcc HttpCheckConfig) Check(serviceUrl string) error {
	for i, check := range hcc.checks {
		err := check.Check(serviceUrl)
		if err != nil {
			return fmt.Errorf("unable to check http %d: %v", i, err)
		}
	}
	return nil
}
