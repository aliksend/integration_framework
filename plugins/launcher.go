package plugins

import (
	"fmt"
	"integration_framework/application_config"
)

type ILauncher interface {
	ConfigUpdated(config *application_config.Config, services map[string]IService) error
	Shutdown() error
}

type launchersConstructor func(tmpDirectory string) (ILauncher, error)

var launchersConstructors map[string]launchersConstructor

func init() {
	launchersConstructors = make(map[string]launchersConstructor)
}

func DefineLauncher(name string, constructor launchersConstructor) {
	_, ok := launchersConstructors[name]
	if ok {
		panic(fmt.Errorf("launcher with name %q already defined", name))
	}
	launchersConstructors[name] = constructor
}

func NewLauncher(name string, tmpDirectory string) (ILauncher, error) {
	constructor, ok := launchersConstructors[name]
	if !ok {
		return nil, fmt.Errorf("launcher %q not defined", name)
	}
	launcher, err := constructor(tmpDirectory)
	if err != nil {
		return nil, fmt.Errorf("unable to create launcher: %v", err)
	}
	return launcher, nil
}
