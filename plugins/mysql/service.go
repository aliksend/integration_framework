package mysql

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_framework/application_config"
)

type Service struct {
	name       string
	env        application_config.ServiceDefinitionEnv
	params     map[string]interface{}
	port       int
	serviceUrl string
	conn       *sqlx.DB
}

func NewService(name string, port int, env application_config.ServiceDefinitionEnv, params map[string]interface{}) *Service {
	return &Service{
		name:   name,
		env:    env,
		port:   port,
		params: params,
	}
}

func (s *Service) Start() error {
	conn, err := sqlx.Connect("mysql", fmt.Sprintf("root:root@tcp(localhost:%d)/%s", s.port, dbName))
	if err != nil {
		return fmt.Errorf("unable to connect to mysql: %v", err)
	}
	s.conn = conn
	return nil
}

func (s Service) WaitForPortAvailable(ctx context.Context) error {
	return nil
}
