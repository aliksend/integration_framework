package docker_compose

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func launchApplication(tmpDirectory string) (*exec.Cmd, error) {
	cmd := exec.Command("docker-compose", []string{
		"--file", path.Join(tmpDirectory, "docker-compose.generated-test.yml"),
		"up", "--build",
	}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start cmd: %v", err)
	}

	return cmd, nil
}

func shutdownApplication(tmpDirectory string) (*exec.Cmd, error) {
	cmd := exec.Command("docker-compose", []string{
		"--file", path.Join(tmpDirectory, "docker-compose.generated-test.yml"),
		"down",
	}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start cmd: %v", err)
	}

	return cmd, nil
}
