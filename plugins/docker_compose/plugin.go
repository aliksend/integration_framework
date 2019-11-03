package docker_compose

import (
	"integration_framework/plugins"
)

func init() {
	plugins.DefineLauncher("docker compose", func(tmpDirectory string) (plugins.ILauncher, error) {
		return NewLauncher(tmpDirectory), nil
	})
}
