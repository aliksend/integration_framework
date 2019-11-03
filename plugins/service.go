package plugins

import (
	"context"
	"fmt"
	"integration_framework/application_config"
)

type IService interface {
	Preparer(params interface{}) (IServicePreparer, error)
	Checker(params interface{}) (IServiceChecker, error)
	WaitForPortAvailable(ctx context.Context) error
	Start() error
}

type IServicePreparer interface {
	PrepareService() error
}

type FnResultSaver func(key string, value interface{})

type IServiceChecker interface {
	CheckService(saveResult FnResultSaver, variables map[string]interface{}) error
}

type serviceConstructor func(name string, port int, env application_config.ServiceDefinitionEnv, params map[string]interface{}) (IService, error)

var serviceConstructors map[string]serviceConstructor

func init() {
	serviceConstructors = make(map[string]serviceConstructor)
}

func DefineService(serviceType string, constructor serviceConstructor) error {
	_, ok := serviceConstructors[serviceType]
	if ok {
		return fmt.Errorf("service with serviceType %q already defined", serviceType)
	}
	serviceConstructors[serviceType] = constructor
	return nil
}

func NewService(serviceName string, serviceType string, port int, env application_config.ServiceDefinitionEnv, params map[string]interface{}) (IService, error) {
	constructor, ok := serviceConstructors[serviceType]
	if !ok {
		return nil, fmt.Errorf("service with type %q not defined", serviceType)
	}
	service, err := constructor(serviceName, port, env, params)
	if err != nil {
		return nil, fmt.Errorf("unable to create service %q: %v", serviceName, err)
	}

	return service, nil
}
