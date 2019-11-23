package smtp

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins/docker_compose"
)

func (s *Service) GenerateDockerComposeConfig(tmpDirectory string, serviceName string, applicationService *docker_compose.DockerComposeService) (dockerComposeServiceName string, dockerComposeService *docker_compose.DockerComposeService, configs *docker_compose.ServiceConfigs, err error) {
	dockerComposeServiceName = "smtp_" + serviceName
	service := docker_compose.DockerComposeService{
		Build: docker_compose.DockerComposeServiceBuild{
			Context: "/home/aliksend/Documents/simple_smtp", // TODO use image
		},
		Restart: "on-failure",
		Environment: map[string]string{
			"HTTP_PORT": "8080",
			"SMTP_PORT": "25",
		},
		Ports: []string{
			fmt.Sprintf("%d:8080", s.port), // expose only http port
		},
	}
	applicationService.AddDependency(dockerComposeServiceName, "tcp", 8080)
	if s.env.EnvMap != nil {
		err := helper.FillEnvironment(&applicationService.Environment, s.env.EnvMap, map[string]string{
			"host": dockerComposeServiceName,
			"port": "25",
		})
		if err != nil {
			return "", nil, nil, fmt.Errorf("unable to fill environment for smtp %q: %v", serviceName, err)
		}
	}
	return dockerComposeServiceName, &service, nil, nil
}
