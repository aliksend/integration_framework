package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type IPrepare interface {
	Prepare(conn *sqlx.DB) error
}

func (s *Service) Preparer(param interface{}) (plugins.IServicePreparer, error) {
	configsList, ok := param.([]interface{})
	if !ok {
		return nil, fmt.Errorf("prepare must be list")
	}
	var prepares []IPrepare
	for _, iconfig := range configsList {
		configYaml, ok := iconfig.(helper.YamlMap)
		if !ok {
			return nil, fmt.Errorf("prepare item must be map")
		}
		config := helper.YamlMapToJsonMap(configYaml)
		for key, value := range config {
			switch key {
			case "exec":
				exec, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf("mysql prepare exec must be string. actual: %v", value)
				}
				prepares = append(prepares, NewExecPrepare(exec))
			case "clear":
				prepares = append(prepares, NewClearPrepare())
			default:
				return nil, fmt.Errorf("cannot create mysql prepare %s", key)
			}
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
		err := prepare.Prepare(ppc.service.conn)
		if err != nil {
			return fmt.Errorf("unable to prepare mysql %d: %v", i, err)
		}
	}
	return nil
}
