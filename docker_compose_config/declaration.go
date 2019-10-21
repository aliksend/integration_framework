package docker_compose_config

type DockerComposeServiceBuild struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}

type DockerComposeService struct {
	Image       string                    `yaml:"image,omitempty"`
	Build       DockerComposeServiceBuild `yaml:"build,omitempty"`
	Restart     string                    `yaml:"restart,omitempty"`
	Volumes     []string                  `yaml:"volumes,omitempty"`
	Environment map[string]string         `yaml:"environment,omitempty"`
	Ports       []string                  `yaml:"ports,omitempty"`
	DependsOn   []string                  `yaml:"depends_on,omitempty"`
	Command     string                    `yaml:"command,omitempty"`
	WorkingDir  string                    `yaml:"working_dir,omitempty"`
}

type DockerComposeConfig struct {
	Version  string                           `yaml:"version"`
	Services map[string]*DockerComposeService `yaml:"services"`
}
