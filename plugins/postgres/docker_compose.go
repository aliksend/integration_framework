package postgres

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins/docker_compose"
)

func (s *Service) GenerateDockerComposeConfig(tmpDirectory string, serviceName string, applicationService *docker_compose.DockerComposeService) (dockerComposeServiceName string, dockerComposeService *docker_compose.DockerComposeService, configs *docker_compose.ServiceConfigs, err error) {
	dockerComposeServiceName = "postgres_" + serviceName
	s.serviceUrl = fmt.Sprintf("postgres://postgres:postgres@%s/%s?sslmode=disable", dockerComposeServiceName, dbName)
	service := docker_compose.DockerComposeService{
		Image:   "postgres:9.6",
		Restart: "always",
		Environment: map[string]string{
			"POSTGRES_DB": dbName,
		},
		Ports: []string{
			fmt.Sprintf("%d:5432", s.port),
		},
		TmpFs: []string{"/var/lib/postgresql/data"},
	}
	applicationService.AddDependency(dockerComposeServiceName, "tcp", 5432)
	if s.env.EnvStr != "" {
		applicationService.Environment[s.env.EnvStr] = s.serviceUrl
	} else if s.env.EnvMap != nil {
		err := helper.FillEnvironment(&applicationService.Environment, s.env.EnvMap, map[string]string{
			"login":    "postgres",
			"password": "postgres",
			"schema":   "postgres",
			"host":     dockerComposeServiceName,
			"port":     "5432",
			"db":       dbName,
		})
		if err != nil {
			return "", nil, nil, fmt.Errorf("unable to fill environment for postgres %q: %v", serviceName, err)
		}
	}
	return dockerComposeServiceName, &service, nil, nil
}
