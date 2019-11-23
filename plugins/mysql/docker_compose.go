package mysql

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins/docker_compose"
)

func (s *Service) GenerateDockerComposeConfig(tmpDirectory string, serviceName string, applicationService *docker_compose.DockerComposeService) (dockerComposeServiceName string, dockerComposeService *docker_compose.DockerComposeService, configs *docker_compose.ServiceConfigs, err error) {
	dockerComposeServiceName = "mysql_" + serviceName
	s.serviceUrl = fmt.Sprintf("root:root@tcp(%s)/%s", dockerComposeServiceName, dbName)
	service := docker_compose.DockerComposeService{
		Image:   "mysql:8.0",
		Restart: "always",
		Environment: map[string]string{
			"MYSQL_DATABASE":      dbName,
			"MYSQL_ROOT_PASSWORD": "root",
		},
		Ports: []string{
			fmt.Sprintf("%d:3306", s.port),
		},
	}
	applicationService.AddDependency(dockerComposeServiceName, "tcp", 3306)
	if s.env.EnvStr != "" {
		applicationService.Environment[s.env.EnvStr] = s.serviceUrl
	} else if s.env.EnvMap != nil {
		err := helper.FillEnvironment(&applicationService.Environment, s.env.EnvMap, map[string]string{
			"login":    "root",
			"password": "root",
			"host":     dockerComposeServiceName,
			"port":     "3306",
			"db":       dbName,
		})
		if err != nil {
			return "", nil, nil, fmt.Errorf("unable to fill environment for mysql %q: %v", serviceName, err)
		}
	}
	return dockerComposeServiceName, &service, nil, nil
}
