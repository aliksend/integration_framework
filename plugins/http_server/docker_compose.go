package http_server

import (
	"encoding/json"
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins/docker_compose"
	"path/filepath"
)

func (s *Service) GenerateDockerComposeConfig(tmpDirectory string, serviceName string, applicationService *docker_compose.DockerComposeService) (dockerComposeServiceName string, dockerComposeService *docker_compose.DockerComposeService, configs *docker_compose.ServiceConfigs, err error) {
	dockerComposeServiceName = "http_" + serviceName
	servicePort := "8080"
	simpleServerConfig := make(map[string]interface{})
	m, ok := s.params["routes"].(helper.YamlMap)
	if ok {
		simpleServerConfig["routes"] = helper.YamlMapToJsonMap(m)
	}
	simpleServerConfig["service_name"] = serviceName
	simpleServerConfig["response_content_type"] = s.params["response_content_type"]
	initialServerConfig, ok := s.params["config"].(helper.YamlMap)
	if ok {
		simpleServerConfig["config"] = helper.YamlMapToJsonMap(initialServerConfig)
	}
	simpleServerConfigBytes, err := json.Marshal(simpleServerConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("unable to marshal simple server %s config: %v", serviceName, err)
	}
	serviceConfigFilename := dockerComposeServiceName + ".simple-server.conf"
	configs = &docker_compose.ServiceConfigs{
		Files: map[string][]byte{
			serviceConfigFilename: simpleServerConfigBytes,
		},
	}
	service := docker_compose.DockerComposeService{
		Build: docker_compose.DockerComposeServiceBuild{
			Context: "/home/aliksend/Documents/simple_server", // TODO use simple_server docker image instead of build directory
		},
		Volumes: []string{
			fmt.Sprintf("%s:%s:ro", filepath.Join(tmpDirectory, serviceConfigFilename), "/config.json"),
		},
		Restart: "on-failure",
		Environment: map[string]string{
			"PORT": servicePort,
		},
		Ports: []string{
			fmt.Sprintf("%d:%s", s.port, servicePort),
		},
	}
	applicationService.AddDependency(dockerComposeServiceName, "tcp", 8080) // http checker will check that `/` returns smth like 200 and fails if it returns 404
	if s.env.EnvStr != "" {
		applicationService.Environment[s.env.EnvStr] = fmt.Sprintf("http://%s:%s/", dockerComposeServiceName, servicePort)
	} else if s.env.EnvMap != nil {
		err := helper.FillEnvironment(&applicationService.Environment, s.env.EnvMap, map[string]string{
			"schema": "http",
			"host":   dockerComposeServiceName,
			"port":   servicePort,
		})
		if err != nil {
			return "", nil, nil, fmt.Errorf("unable to fill environment for http %q: %v", serviceName, err)
		}
	}
	return dockerComposeServiceName, &service, configs, nil
}
