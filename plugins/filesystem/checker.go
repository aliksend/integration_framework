package filesystem

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type ICheck interface {
	Check(mountsRoot string, saveResult plugins.FnResultSaver, variables map[string]interface{}) error
}

func (s *Service) Checker(checkConfig interface{}) (plugins.IServiceChecker, error) {
	configsList, ok := checkConfig.([]interface{})
	if !ok {
		return nil, fmt.Errorf("service check for filesystem must be list, but it is %T (%#v)", checkConfig, checkConfig)
	}
	var checkers []ICheck
	for _, configInterface := range configsList {
		configYaml, ok := configInterface.(helper.YamlMap)
		if !ok {
			return nil, fmt.Errorf("service check for filesystem must be map, but it is %T (%#v)", configInterface, configInterface)
		}
		configMap := helper.YamlMapToJsonMap(configYaml)
		for checkName, params := range configMap {
			switch checkName {
			case "exists":
				paramsMap, ok := params.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("param for filesystem exists checker should be map, but it is %T (%#v)", params, params)
				}
				filename, ok := paramsMap["name"].(string)
				if !ok {
					return nil, fmt.Errorf("name of file should be set for filesystem exists checker")
				}
				exists, ok := paramsMap["exists"].(bool)
				if !ok {
					exists = true
				}
				checkers = append(checkers, NewFileExistsChecker(filename, exists))
			default:
				return nil, fmt.Errorf("checker %q not exists for filesystem service", checkName)
			}
		}
	}

	return &Checker{
		service:  s,
		checkers: checkers,
	}, nil
}

type Checker struct {
	service  *Service
	checkers []ICheck
}

func (pcc Checker) CheckService(saveResult plugins.FnResultSaver, variables map[string]interface{}) error {
	for i, check := range pcc.checkers {
		err := check.Check(pcc.service.mountsRoot, saveResult, variables)
		if err != nil {
			return fmt.Errorf("unable to check filesystem %d: %v", i, err)
		}
	}
	return nil
}
