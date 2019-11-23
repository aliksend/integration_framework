package docker_compose

import (
	"fmt"
	"integration_framework/application_config"
	"integration_framework/plugins"
	"sort"
	"strings"
)

type IDockerComposeConfigModifier interface {
	ModifyApplicationConfig(*DockerComposeService) error
}

type IServiceWithDockerComposeConfigGenerator interface {
	GenerateDockerComposeConfig(tmpDirectory string, serviceName string, applicationService *DockerComposeService) (dockerComposeServiceName string, dockerComposeService *DockerComposeService, configs *ServiceConfigs, err error)
}

func NewDockerComposeConfig(tmpDirectory string, config *application_config.Config, services map[string]plugins.IService) (*DockerComposeConfig, []ServiceConfigs, error) {
	dockerComposeConfig := DockerComposeConfig{
		Version:  "3.4",
		Services: make(map[string]*DockerComposeService),
	}
	applicationService := DockerComposeService{
		config: config.Application,
		Build: DockerComposeServiceBuild{
			Context:    config.Application.Path,
			Dockerfile: "Dockerfile.prod",
		},
		Ports: []string{
			"8080:8080",
		},
		Environment: config.Environment,
		Restart:     "on-failure",
		User:        "${UID}:${GID}",
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
			return nil, nil, fmt.Errorf("unable to generate docker compose service for external service %q: %v", serviceName, err)
		}
		if dockerComposeService != nil {
			dockerComposeConfig.Services[dockerComposeServiceName] = dockerComposeService
			if serviceConfigs != nil {
				servicesConfigs = append(servicesConfigs, *serviceConfigs)
			}
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
	config      *application_config.ApplicationConfig
	Image       string                    `yaml:"image,omitempty"`
	Build       DockerComposeServiceBuild `yaml:"build,omitempty"`
	Restart     string                    `yaml:"restart,omitempty"`
	Volumes     []string                  `yaml:"volumes,omitempty"`
	Environment map[string]string         `yaml:"environment,omitempty"`
	Ports       []string                  `yaml:"ports,omitempty"`
	DependsOn   []string                  `yaml:"depends_on,omitempty"`
	Command     string                    `yaml:"command,omitempty"`
	WorkingDir  string                    `yaml:"working_dir,omitempty"`
	User        string                    `yaml:"user,omitempty"`
	waitFor     []string
}

type DockerComposeConfig struct {
	Version  string                           `yaml:"version"`
	Services map[string]*DockerComposeService `yaml:"services"`
}

type ServiceConfigs struct {
	Files map[string][]byte
}

func (s *DockerComposeService) AddDependency(serviceName string, protocol string, port int) {
	s.DependsOn = append(s.DependsOn, serviceName)
	if s.config.Dockerize {
		s.waitFor = append(s.waitFor, fmt.Sprintf("-wait %s://%s:%d -timeout 60s", protocol, serviceName, port))
		s.Environment["DOCKERIZE_ARGS"] = strings.Join(s.waitFor, " ")
	}
}
