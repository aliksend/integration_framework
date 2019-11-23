package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	"integration_framework/application_config"
	"integration_framework/plugins"
)

const dbName = "test"

func init() {
	plugins.DefineService("mysql", func(name string, port int, env application_config.ServiceDefinitionEnv, params map[string]interface{}) (plugins.IService, error) {
		return NewService(name, port, env, params), nil
	})
}
