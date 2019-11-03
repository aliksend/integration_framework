package http_server

import (
	"integration_framework/application_config"
	"integration_framework/plugins"
)

func init() {
	plugins.DefineService("http", func(name string, port int, env application_config.ServiceDefinitionEnv, params map[string]interface{}) (plugins.IService, error) {
		return NewService(name, port, env, params), nil
	})
}
