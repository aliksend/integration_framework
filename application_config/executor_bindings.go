package application_config

import (
	"fmt"
)

func (s ExternalServiceDefinitionSection) Start() error {
	if s.executor != nil {
		return s.executor.Start()
	}
	return fmt.Errorf("executor not initialized")
}
