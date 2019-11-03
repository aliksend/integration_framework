package plugins

import (
	"fmt"
)

type IEnvironmentInitializer interface {
	InitEnvironment() error
}

type environmentInitializersConstructor func(params interface{}) (IEnvironmentInitializer, error)

var environmentInitializersConstructors map[string]environmentInitializersConstructor

func init() {
	environmentInitializersConstructors = make(map[string]environmentInitializersConstructor)
}

func DefineEnvironmentInitializer(name string, constructor environmentInitializersConstructor) error {
	_, ok := environmentInitializersConstructors[name]
	if ok {
		return fmt.Errorf("environment initializer with name %q already defined", name)
	}
	environmentInitializersConstructors[name] = constructor
	return nil
}

func NewEnvironmentInitializer(name string, params interface{}) (IEnvironmentInitializer, error) {
	constructor, ok := environmentInitializersConstructors[name]
	if !ok {
		return nil, fmt.Errorf("environment initializer %q not defined", name)
	}
	environmentInitializer, err := constructor(params)
	if err != nil {
		return nil, fmt.Errorf("unable to create environment initializer: %v", err)
	}
	return environmentInitializer, nil
}
