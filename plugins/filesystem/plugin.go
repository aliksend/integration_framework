package filesystem

import (
	"fmt"
	"integration_framework/application_config"
	"integration_framework/helper"
	"integration_framework/plugins"
)

func init() {
	plugins.DefineService("filesystem", func(serviceName string, port int, env application_config.ServiceDefinitionEnv, params map[string]interface{}) (plugins.IService, error) {
		paramsMounts, ok := params["mounts"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("mounts should be list, but it is %T (%#v)", params["mounts"], params["mounts"])
		}
		var mounts []Mount
		for _, mount := range paramsMounts {
			mountYamlMap, ok := mount.(helper.YamlMap)
			if !ok {
				return nil, fmt.Errorf("mount should be map, but it is %T (%#v)", mount, mount)
			}
			mountMap := helper.YamlMapToJsonMap(mountYamlMap)
			name, ok := mountMap["name"].(string)
			if !ok {
				return nil, fmt.Errorf("mount name (string) should be defined for mount %#v", mountMap)
			}
			path, ok := mountMap["path"].(string)
			if !ok {
				return nil, fmt.Errorf("mount path (string) should be defined for mount %#v", mountMap)
			}
			mounts = append(mounts, Mount{
				Name: name,
				Path: path,
			})
		}
		return NewService(serviceName, mounts), nil
	})
}
