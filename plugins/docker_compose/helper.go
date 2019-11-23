package docker_compose

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
)

func launchApplication(tmpDirectory string) (*exec.Cmd, error) {
	cmd := exec.Command("docker-compose", []string{
		"--file", path.Join(tmpDirectory, "docker-compose.generated-test.yml"),
		"up", "--build",
	}...)

	// https://github.com/moby/moby/issues/3206#issuecomment-152682860
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get current user info: %v", err)
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "UID", currentUser.Uid), fmt.Sprintf("%s=%s", "GID", currentUser.Gid))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
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

	// https://github.com/moby/moby/issues/3206#issuecomment-152682860
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get current user info: %v", err)
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "UID", currentUser.Uid), fmt.Sprintf("%s=%s", "GID", currentUser.Gid))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start cmd: %v", err)
	}

	return cmd, nil
}
