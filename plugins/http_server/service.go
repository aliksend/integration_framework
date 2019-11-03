package http_server

import (
	"context"
	"integration_framework/application_config"
	"integration_framework/helper"
	"time"
)

type Service struct {
	name   string
	env    application_config.ServiceDefinitionEnv
	params map[string]interface{}
	port   int
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
	return nil
}

func (s Service) WaitForPortAvailable(ctx context.Context) error {
	for {
		portAvailable := helper.IsHttpPortAvailable("localhost", s.port)
		if portAvailable {
			return nil
		}

		select {
		case _, ok := <-ctx.Done():
			if ok {
				// value readed
				return nil
			} else {
				// channel closed
				return nil
			}
		default:
			// chanel is empty
			// continue checking
		}

		time.Sleep(5 * time.Second)
	}
	return nil
}
