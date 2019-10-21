package application_config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func LoadConfig(filename string) (*ApplicationConfig, error) {
	config := ApplicationConfig{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file error: %v", err)
	}
	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("config invalid: %v", err)
	}
	return &config, nil
}
