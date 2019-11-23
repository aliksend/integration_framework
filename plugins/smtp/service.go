package smtp

import (
	"context"
	"integration_framework/application_config"
)

type Service struct {
	name string
	port int
	env  application_config.ServiceDefinitionEnv
}

func NewService(name string, port int, env application_config.ServiceDefinitionEnv) (*Service, error) {
	return &Service{
		name: name,
		port: port,
		env:  env,
	}, nil
}

func (s *Service) Start() error {
	// TODO establish connect with server
	return nil
}

func (s Service) WaitForPortAvailable(ctx context.Context) error {
	return nil
}
