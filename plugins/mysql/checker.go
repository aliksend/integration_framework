package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type ICheck interface {
	Check(conn *sqlx.DB, saveResult plugins.FnResultSaver, variables map[string]interface{}) error
}

func (s *Service) Checker(param interface{}) (plugins.IServiceChecker, error) {
	configsList, ok := param.([]interface{})
	if !ok {
		return nil, fmt.Errorf("service check for mysql must be list, but it is %T (%#v)", param, param)
	}
	var checkers []ICheck
	for _, configInterface := range configsList {
		configYaml, ok := helper.IsYamlMap(configInterface)
		if !ok {
			return nil, fmt.Errorf("service check for mysql must be map, but it is %T (%#v)", configInterface, configInterface)
		}
		config := configYaml.ToMap()
		query, ok := config["query"].(string)
		if !ok {
			return nil, fmt.Errorf("service check query for mysql must be string, but it is %T %#v", config["query"], config["query"])
		}
		var expectedResult interface{}
		expectedResultStruct, ok := config["expected_result"].(map[string]interface{})
		if ok {
			expectedResult = expectedResultStruct
		}
		expectedResultSlice, ok := config["expected_result"].([]interface{})
		if ok {
			expectedResult = expectedResultSlice
		}
		saveResultTo, _ := config["save_result_to"].(string)

		checkers = append(checkers, NewQueryChecker(query, expectedResult, saveResultTo))
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
		err := check.Check(pcc.service.conn, saveResult, variables)
		if err != nil {
			return fmt.Errorf("unable to check mysql %d: %v", i, err)
		}
	}
	return nil
}
