package application_config

import (
	"fmt"
)

func (s EnvironmentInitializationSection) Init(externalServices map[string]*ExternalServiceDefinitionSection) error {
	err := s.PrepareExternalServices.Prepare(externalServices)
	if err != nil {
		return fmt.Errorf("unable to prepare external services: %v", err)
	}
	return nil
}
