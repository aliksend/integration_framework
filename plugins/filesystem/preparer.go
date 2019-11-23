package filesystem

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type IPrepare interface {
	Prepare(mountsRoot string, mounts []Mount) error
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
			case "clear":
				prepares = append(prepares, NewClearPrepare())
			case "file":
				params, ok := value.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("filesystem prepare file must be map, but it is %T (%#v)", value, value)
				}
				name, ok := params["name"].(string)
				if !ok {
					return nil, fmt.Errorf("filesystem prepare file.name must be string, but it is %T (%#v)", params["name"], params["name"])
				}
				content, ok := params["content"].(string)
				if !ok {
					return nil, fmt.Errorf("filesystem prepare file.content must be string, but it is %T (%#v)", params["content"], params["content"])
				}
				prepares = append(prepares, NewFilePrepare(name, content))
			default:
				return nil, fmt.Errorf("cannot create filesystem prepare %s", key)
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
		err := prepare.Prepare(ppc.service.mountsRoot, ppc.service.mounts)
		if err != nil {
			return fmt.Errorf("unable to prepare filesystem %d: %v", i, err)
		}
	}
	return nil
}
