package docker_compose

import (
	"fmt"
	"integration_framework/application_config"
	"integration_framework/plugins"
	"sort"
)

type IDockerComposeConfigModifier interface {
	ModifyApplicationConfig(*DockerComposeService) error
}

type IServiceWithDockerComposeConfigGenerator interface {
	GenerateDockerComposeConfig(tmpDirectory string, serviceName string, applicationService *DockerComposeService) (dockerComposeServiceName string, dockerComposeService *DockerComposeService, configs *ServiceConfigs, err error)
}

func NewDockerComposeConfig(tmpDirectory string, config *application_config.Config, services map[string]plugins.IService, environmentInitializers []plugins.IEnvironmentInitializer) (*DockerComposeConfig, []ServiceConfigs, error) {
	dockerComposeConfig := DockerComposeConfig{
		Version:  "3.4",
		Services: make(map[string]*DockerComposeService),
	}
	applicationService := DockerComposeService{
		Build: DockerComposeServiceBuild{
			Context:    config.Application.Path,
			Dockerfile: "Dockerfile.prod",
		},
		Ports: []string{
			"8080:8080",
		},
		Environment: make(map[string]string),
		Restart:     "on-failure",
	}
	for _, environmentInitializer := range environmentInitializers {
		dockerComposeConfigModifier, ok := environmentInitializer.(IDockerComposeConfigModifier)
		if ok {
			err := dockerComposeConfigModifier.ModifyApplicationConfig(&applicationService)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to modify application config using environment initializer: %v", err)
			}
		}
	}
	var servicesConfigs []ServiceConfigs
	var servicesNames []string
	for serviceName := range services {
		servicesNames = append(servicesNames, serviceName)
	}
	sort.Strings(servicesNames)
	for _, serviceName := range servicesNames {
		service, ok := services[serviceName].(IServiceWithDockerComposeConfigGenerator)
		if !ok {
			return nil, nil, fmt.Errorf("service %q doesnt support generating docker compose", serviceName)
		}
		dockerComposeServiceName, dockerComposeService, serviceConfigs, err := service.GenerateDockerComposeConfig(tmpDirectory, serviceName, &applicationService)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate docker compose service for external service %q", serviceName)
		}
		dockerComposeConfig.Services[dockerComposeServiceName] = dockerComposeService
		if serviceConfigs != nil {
			servicesConfigs = append(servicesConfigs, *serviceConfigs)
		}
	}
	dockerComposeConfig.Services["application"] = &applicationService
	return &dockerComposeConfig, servicesConfigs, nil
}

type DockerComposeServiceBuild struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile,omitempty"`
}

type DockerComposeService struct {
	Image       string                    `yaml:"image,omitempty"`
	Build       DockerComposeServiceBuild `yaml:"build,omitempty"`
	Restart     string                    `yaml:"restart,omitempty"`
	Volumes     []string                  `yaml:"volumes,omitempty"`
	Environment map[string]string         `yaml:"environment,omitempty"`
	Ports       []string                  `yaml:"ports,omitempty"`
	DependsOn   []string                  `yaml:"depends_on,omitempty"`
	Command     string                    `yaml:"command,omitempty"`
	WorkingDir  string                    `yaml:"working_dir,omitempty"`
}

type DockerComposeConfig struct {
	Version  string                           `yaml:"version"`
	Services map[string]*DockerComposeService `yaml:"services"`
}

type ServiceConfigs struct {
	Files map[string][]byte
}
