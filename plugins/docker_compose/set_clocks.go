package docker_compose

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
	"time"
)

func init() {
	plugins.DefineEnvironmentInitializer("clocks", func(params interface{}) (plugins.IEnvironmentInitializer, error) {
		timeStr, ok := params.(string)
		if !ok || timeStr == "" {
			return nil, fmt.Errorf("clocks param should be non-empty string")
		}
		parsedTime, err := time.Parse("2006-01-02T15:04:05Z", timeStr)
		if err != nil {
			return nil, fmt.Errorf("unable to parse time")
		}
		return NewSetClocksPlugin(parsedTime), nil
	})
}

func NewSetClocksPlugin(timeToSet time.Time) SetClocksPlugin {
	return SetClocksPlugin{
		timeToSet: timeToSet,
	}
}

type SetClocksPlugin struct {
	timeToSet time.Time
}

func (p SetClocksPlugin) InitEnvironment() error {
	return nil
}

func (p SetClocksPlugin) ModifyApplicationConfig(applicationConfig *DockerComposeService) error {
	applicationConfig.Environment["FAKE_TIME"] = p.timeToSet.Format(helper.TimeLayout)
	return nil
}
