package testing

import (
	"encoding/json"
	"fmt"
	"integration_framework/application_config"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type createTesterContext struct {
	EnvironmentInitializers []plugins.IEnvironmentInitializer
	Services                map[string]plugins.IService
}

type Tester struct {
	Name                    string
	servicePreparers        []plugins.IServicePreparer
	serviceCheckers         []plugins.IServiceChecker
	environmentInitializers []plugins.IEnvironmentInitializer
	requester               plugins.IRequester
	expectedResponse        *map[interface{}]interface{}
	expectedCode            int
	saveResponseTo          string
}

func (pc *ParsedConfig) createTester(testCaseName string, servicePreparers []plugins.IServicePreparer, serviceCheckers []plugins.IServiceChecker, testCase *application_config.TestCase, requester plugins.IRequester) error {
	pc.Testers = append(pc.Testers, Tester{
		Name:                    testCaseName,
		servicePreparers:        servicePreparers,
		serviceCheckers:         serviceCheckers,
		environmentInitializers: pc.EnvironmentInitializers,
		requester:               requester,
		expectedResponse:        testCase.ExpectedResponse,
		expectedCode:            testCase.ExpectedCode,
		saveResponseTo:          testCase.SaveResponseTo,
	})
	return nil
}

func (t Tester) Exec() error {
	for _, environmentInitializer := range t.environmentInitializers {
		err := environmentInitializer.InitEnvironment()
		if err != nil {
			return fmt.Errorf("unable to initialize environment: %v", err)
		}
	}

	for _, servicePreparer := range t.servicePreparers {
		err := servicePreparer.PrepareService()
		if err != nil {
			return fmt.Errorf("unable to prepare service: %v", err)
		}
	}

	variables := make(map[string]interface{})
	saveResult := func(key string, value interface{}) {
		variables[key] = value
	}

	err := t.checkRequest(saveResult, variables)
	if err != nil {
		return fmt.Errorf("unable to check request: %v", err)
	}

	for _, serviceChecker := range t.serviceCheckers {
		err := serviceChecker.CheckService(saveResult, variables)
		if err != nil {
			return fmt.Errorf("unable to check service: %v", err)
		}
	}

	return nil
}

func (t Tester) checkRequest(saveResult plugins.FnResultSaver, variables map[string]interface{}) error {
	responseBody, statusCode, err := t.requester.MakeRequest()
	if err != nil {
		return fmt.Errorf("unable to make request: %v", err)
	}

	actualBody := make(map[string]interface{})
	err = json.Unmarshal(responseBody, &actualBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response body %s: %v", responseBody, err)
	}

	if t.expectedResponse != nil {
		fmt.Printf(">> actual response %#v\n", actualBody)
		// expected response defined ...
		if *t.expectedResponse == nil {
			// ... but it defined like "null", not like map
			if len(responseBody) != 0 {
				return fmt.Errorf("invalid response: expected empty response, but actual response is %s", responseBody)
			}
		} else {
			// ... and it defined like map
			expectedBody := helper.YamlMapToJsonMap(*t.expectedResponse)
			err = IsEqual(actualBody, expectedBody)
			if err != nil {
				return fmt.Errorf("invalid response: %v", err)
			}
		}
	}
	if t.expectedCode != 0 {
		if statusCode != t.expectedCode {
			return fmt.Errorf("invalid response status code: expected %d to equal %d", statusCode, t.expectedCode)
		}
	}
	if t.saveResponseTo != "" {
		saveResult(t.saveResponseTo, actualBody)
	}

	return nil
}
