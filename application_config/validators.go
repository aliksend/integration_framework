package application_config

import (
	"fmt"
)

func (esde ExternalServiceDefinitionEnv) Validate() error {
	if esde.EnvStr == "" && (esde.EnvMap == nil || len(esde.EnvMap) == 0) {
		return fmt.Errorf("env must be either string or map")
	}
	return nil
}

func (esds ExternalServiceDefinitionSection) Validate() error {
	if err := esds.Env.Validate(); err != nil {
		return fmt.Errorf("invalid env: %v", err)
	}
	return nil
}
func (bes BeforeEachSection) Validate() error {
	return nil
}
func (tc TestCase) Validate() error {
	return nil
}
func (ts TestSection) Validate() error {
	if err := ts.BeforeEach.Validate(); err != nil {
		return fmt.Errorf("invalid berofe each section: %v", err)
	}
	for i, testCase := range ts.Cases {
		if err := testCase.Validate(); err != nil {
			return fmt.Errorf("invalid case %d: %v", i, err)
		}
	}
	return nil
}
func (c ApplicationConfig) Validate() error {
	for name, externalServiceDefinition := range c.ExternalServices {
		err := externalServiceDefinition.Validate()
		if err != nil {
			return fmt.Errorf("invalid external service %s definition: %v", name, err)
		}
	}
	for name, test := range c.Tests {
		err := test.Validate()
		if err != nil {
			return fmt.Errorf("invalid test section %s: %v", name, err)
		}
	}
	for i, testCase := range c.GeneralCases {
		if err := testCase.Validate(); err != nil {
			return fmt.Errorf("invalid general case %d: %v", i, err)
		}
	}
	return nil
}
