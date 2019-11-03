package http_server

import (
	"encoding/json"
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type ICheck interface {
	Check(serviceUrl string) error
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

func (s *Service) Checker(checkConfig interface{}) (plugins.IServiceChecker, error) {
	checkYaml, ok := checkConfig.(helper.YamlMap)
	if !ok {
		return nil, fmt.Errorf("http check config should be map")
	}
	var checks []ICheck
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
			checks = append(checks, NewCallsCheck(calls))
		default:
			return nil, fmt.Errorf("invalid http check action %q", action)
		}
	}
	return &CheckConfig{
		service: s,
		checks:  checks,
	}, nil
}

type CheckConfig struct {
	service *Service
	checks  []ICheck
}

func (hcc CheckConfig) CheckService(saveResult plugins.FnResultSaver, variables map[string]interface{}) error {
	for i, check := range hcc.checks {
		err := check.Check(fmt.Sprintf("http://localhost:%d/", hcc.service.port))
		if err != nil {
			return fmt.Errorf("unable to check http %d: %v", i, err)
		}
	}
	return nil
}
