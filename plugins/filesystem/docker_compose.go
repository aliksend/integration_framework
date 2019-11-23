package filesystem

import (
	"fmt"
	"integration_framework/plugins/docker_compose"
	"os"
	"path/filepath"
)

func (s *Service) GenerateDockerComposeConfig(tmpDirectory string, serviceName string, applicationService *docker_compose.DockerComposeService) (dockerComposeServiceName string, dockerComposeService *docker_compose.DockerComposeService, configs *docker_compose.ServiceConfigs, err error) {
	s.mountsRoot = filepath.Join(tmpDirectory, serviceName)
	volumes := make([]string, len(s.mounts))
	for i, mount := range s.mounts {
		localPathToMount := filepath.Join(s.mountsRoot, mount.Name)
		// create path for volumes manually
		// or docker service will create it using root as owner
		// and it will not be accessible by current user
		err := os.MkdirAll(localPathToMount, os.ModePerm)
		if err != nil {
			return "", nil, nil, fmt.Errorf("unable to create volume mount %q: %v", localPathToMount, err)
		}
		volumes[i] = fmt.Sprintf("%s:%s", localPathToMount, mount.Path)
	}
	applicationService.Volumes = append(applicationService.Volumes, volumes...)
	return "filesystem_" + serviceName, nil, nil, nil
}
