package docker_compose

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"integration_framework/application_config"
	"integration_framework/helper"
	"integration_framework/plugins"
	"integration_framework/testing"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func NewLauncher(tmpDirectory string) *Launcher {
	return &Launcher{
		tmpDirectory: tmpDirectory,
	}
}

type Launcher struct {
	tmpDirectory                     string
	lastGeneratedDockerComposeConfig []byte
	lastWritedFiles                  map[string][]byte
	cmd                              *exec.Cmd
	shutdownRequested                bool
	cancelContext                    func()
}

func (l *Launcher) createConfig(config *application_config.Config, services map[string]plugins.IService) (configChanged bool, err error) {
	dockerComposeConfig, servicesConfigs, err := NewDockerComposeConfig(l.tmpDirectory, config, services)
	if err != nil {
		return false, fmt.Errorf("unable to generate docker-compose config: %v", err)
	}
	dockerComposeBytes, err := yaml.Marshal(dockerComposeConfig)
	if err != nil {
		return false, fmt.Errorf("unable to marshal docker compose config: %v", err)
	}

	dockerComposeConfigUpdated := true
	if l.lastGeneratedDockerComposeConfig != nil {
		dockerComposeConfigUpdated, err = testing.IsEqualYamlMaps(l.lastGeneratedDockerComposeConfig, dockerComposeBytes)
		if err != nil {
			return false, fmt.Errorf("unable to check docker compose config equality: %v", err)
		}
	}
	if dockerComposeConfigUpdated {
		configChanged = true
		l.lastGeneratedDockerComposeConfig = dockerComposeBytes
		err = ioutil.WriteFile(l.tmpDirectory+"/docker-compose.generated-test.yml", dockerComposeBytes, os.ModePerm)
		if err != nil {
			return false, fmt.Errorf("unable to write docker compose config: %v", err)
		}
	}

	if l.lastWritedFiles == nil {
		l.lastWritedFiles = make(map[string][]byte)
	}
	for _, serviceConfigs := range servicesConfigs {
		for fileName, fileContents := range serviceConfigs.Files {
			alreadyWrited := l.lastWritedFiles[fileName]
			err = testing.IsEqual(alreadyWrited, fileContents)
			if err != nil {
				configChanged = true
				l.lastWritedFiles[fileName] = fileContents
				err = ioutil.WriteFile(filepath.Join(l.tmpDirectory, fileName), fileContents, os.ModePerm)
				if err != nil {
					return false, fmt.Errorf("unable to write docker compose config: %v", err)
				}
			}
		}
	}
	return
}

func (l *Launcher) launchApplication(services map[string]plugins.IService) error {
	fmt.Println("--------------------> launch app")
	cmd, err := launchApplication(l.tmpDirectory)
	if err != nil {
		return fmt.Errorf("unable to launch docker compose: %v", err)
	}
	l.cmd = cmd
	fmt.Println("---> launched")

	for {
		portAvailable := helper.IsHttpPortAvailable("localhost", 8080)
		if portAvailable {
			break
		}
		if l.shutdownRequested {
			break
		}
		time.Sleep(5 * time.Second)
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.cancelContext = cancel
	for _, service := range services {
		if l.shutdownRequested {
			return nil
		}
		err := service.WaitForPortAvailable(ctx)
		if err != nil {
			return fmt.Errorf("wait for port error: %v", err)
		}
	}
	fmt.Println("---> return from launch application")
	return nil
}

func (l *Launcher) ConfigUpdated(config *application_config.Config, services map[string]plugins.IService) error {
	updated, err := l.createConfig(config, services)
	if err != nil {
		return fmt.Errorf("unable to create config: %v", err)
	}
	if !updated {
		return nil
	}
	err = l.Shutdown()
	if err != nil {
		return fmt.Errorf("unable to shutdown application")
	}
	l.shutdownRequested = false
	err = l.launchApplication(services)
	if err != nil {
		return fmt.Errorf("unable to launch application: %v", err)
	}
	return nil
}

func (l *Launcher) Shutdown() error {
	l.shutdownRequested = true
	if l.cancelContext != nil {
		l.cancelContext()
	}
	if l.cmd != nil {
		err := l.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			return fmt.Errorf("unable to interrupt cmd: %v", err)
		}
		err = l.cmd.Wait()
		if err != nil {
			fmt.Printf("unable to wait for cmd: %v\n", err)
		}
		l.cmd = nil
		shutdownCmd, err := shutdownApplication(l.tmpDirectory)
		if err != nil {
			return fmt.Errorf("unable to shutdown application: %v", err)
		}
		err = shutdownCmd.Wait()
		if err != nil {
			fmt.Printf("unable to wait for shutdown cmd: %v\n", err)
		}
	}
	return nil
}
