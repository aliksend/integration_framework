package docker_compose_launcher

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func IsPortAvailable(host string, port string) bool {
	_, err := http.Get(fmt.Sprintf("http://%s:%s", host, port))
	if err != nil {
		// fmt.Printf("is port available connection err: %v\n", err)
		return false
	}
	return true
}

func LaunchApplication(dstDirectory string) (*exec.Cmd, error) {
	cmd := exec.Command("docker-compose", []string{
		"--file", path.Join(dstDirectory, "docker-compose.generated-test.yml"),
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
