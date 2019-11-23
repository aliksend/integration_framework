package smtp

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type IPrepare interface {
	Prepare(httpServiceUrl string) error
}

func (s *Service) Preparer(param interface{}) (plugins.IServicePreparer, error) {
	prepareConfigYaml, ok := param.(helper.YamlMap)
	if !ok {
		return nil, fmt.Errorf("prepare must be map, but it is %T (%#v)", param, param)
	}
	var prepares []IPrepare
	for prepareName := range helper.YamlMapToJsonMap(prepareConfigYaml) {
		switch prepareName {
		case "clear":
			prepares = append(prepares, NewClearAllPreparer())
		default:
			return nil, fmt.Errorf("preparer %q not defined for smtp", prepareName)
		}
	}

	return &Preparer{
		service:  s,
		prepares: prepares,
	}, nil
}

type Preparer struct {
	service  *Service
	prepares []IPrepare
}

func (ppc Preparer) PrepareService() error {
	for i, prepare := range ppc.prepares {
		err := prepare.Prepare(fmt.Sprintf("http://localhost:%d/", ppc.service.port))
		if err != nil {
			return fmt.Errorf("unable to prepare smtp %d: %v", i, err)
		}
	}
	return nil
}
