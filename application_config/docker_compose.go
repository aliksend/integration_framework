package application_config

import (
	"fmt"
	"integration_tests/docker_compose_config"
	"integration_tests/helper"
)

func (c ApplicationConfig) DockerComposeConfig(dstDirectory string, config *ApplicationConfig) (*docker_compose_config.DockerComposeConfig, error) {
	dockerComposeConfig := docker_compose_config.DockerComposeConfig{
		Version:  "3.4",
		Services: make(map[string]*docker_compose_config.DockerComposeService),
	}
	environment := make(map[string]string)
	if config.InitEnvironment.SetClocks != nil {
		environment["FAKE_TIME"] = config.InitEnvironment.SetClocks.Format(helper.TimeLayout)
	}
	applicationService := docker_compose_config.DockerComposeService{
		Build: docker_compose_config.DockerComposeServiceBuild{
			Context:    "../..",
			Dockerfile: "Dockerfile.prod",
		},
		Ports: []string{
			"8080:8080",
		},
		Environment: environment,
		Restart:     "on-failure",
	}
	for name, definition := range config.ExternalServices {
		serviceName, dockerComposeService, err := definition.executor.Generate(dstDirectory, name, &applicationService)
		if err != nil {
			return nil, fmt.Errorf("unable to generate docker compose service for external service %q", name)
		}
		dockerComposeConfig.Services[serviceName] = dockerComposeService
	}
	dockerComposeConfig.Services["application"] = &applicationService
	return &dockerComposeConfig, nil
}
