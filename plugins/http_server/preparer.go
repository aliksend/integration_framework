package http_server

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type IPrepare interface {
	Prepare(serviceUrl string) error
}

func (s *Service) Preparer(params interface{}) (plugins.IServicePreparer, error) {
	prepareYaml, ok := params.(helper.YamlMap)
	if !ok {
		return nil, fmt.Errorf("http prepare config should be map")
	}
	var prepares []IPrepare
	prepare := helper.YamlMapToJsonMap(prepareYaml)
	for action, value := range prepare {
		switch action {
		case "calls":
			if value != nil {
				return nil, fmt.Errorf("can only reset calls in prepare: value should be null")
			}
			prepares = append(prepares, NewResetCallsPrepare())
		case "config":
			if value == nil {
				prepares = append(prepares, NewConfigPrepare(nil))
				break
			}
			valueMap, ok := value.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("http service config should be map, but it is %T (%#v)", value, value)
			}
			prepares = append(prepares, NewConfigPrepare(valueMap))
		default:
			return nil, fmt.Errorf("invalid http prepare action %q", action)
		}
	}
	return &PrepareConfig{
		service:  s,
		prepares: prepares,
	}, nil
}

type PrepareConfig struct {
	service  *Service
	prepares []IPrepare
}

func (hpc PrepareConfig) PrepareService() error {
	for i, prepare := range hpc.prepares {
		err := prepare.Prepare(fmt.Sprintf("http://localhost:%d/", hpc.service.port))
		if err != nil {
			return fmt.Errorf("unable to prepare http %d: %v", i, err)
		}
	}
	return nil
}
