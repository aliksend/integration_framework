package postgres

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_framework/plugins"
	"integration_framework/testing"
	"regexp"
	"strconv"
	"strings"
)

type QueryChecker struct {
	query          string
	expectedResult interface{}
	saveResultTo   string
}

func (pc QueryChecker) makeRequest(conn *sqlx.DB) ([]interface{}, error) {
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

func NewQueryChecker(query string, expectedResult interface{}, saveResultTo string) *QueryChecker {
	return &QueryChecker{
		query:          query,
		expectedResult: expectedResult,
		saveResultTo:   saveResultTo,
	}
}

func (pc QueryChecker) Check(conn *sqlx.DB, saveResult plugins.FnResultSaver, variables map[string]interface{}) error {
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
