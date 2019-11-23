package application_config

import (
	"fmt"
)

type Config struct {
	Launcher     string                   `yaml:"launcher"`
	Application  *ApplicationConfig       `yaml:"application"`
	Environment  map[string]string        `yaml:"environment"`
	GeneralCases map[string]GeneralCase   `yaml:"general_cases"`
	Services     map[string]ServiceConfig `yaml:"services"`
	Cases        TestCases                `yaml:"cases"`
}

type TestCases map[string]*TestCase

type GeneralCase struct {
	AutoInclude bool      `yaml:"auto_include"`
	Cases       TestCases `yaml:"cases"`
}

type TestCase struct {
	PrepareServices map[string]interface{} `yaml:"prepare_services"`
	Request         interface{}            `yaml:"request"`
	// if ModifyRequest is empty then default headers will be added
	// else headers will not be added and ModifyRequest will be executed
	ModifyRequest *string `yaml:"modify_request"`
	// expected response can not be defined (so field will be <nil>)
	// also it can be defined as `null` (so field will be pointer to <nil>)
	// also it can be map (so field will be pointer to map)
	ExpectedResponse *map[interface{}]interface{} `yaml:"expected_response"`
	ExpectedCode     int                          `yaml:"expected_code"`
	CheckServices    []map[string]interface{}     `yaml:"check_services"`
	Only             bool                         `yaml:"only"`
	Skip             bool                         `yaml:"skip"`
	Cases            TestCases                    `yaml:"cases"`
	GeneralCases     *GeneralCasesSelector        `yaml:"general_cases"`
}

type GeneralCasesSelector struct {
	Request interface{} `yaml:"request"`
	Include []string    `yaml:"include"`
	Exclude []string    `yaml:"exclude"`
}

type ApplicationConfig struct {
	Path            string          `yaml:"path"`
	RequestType     string          `yaml:"request_type"`
	RequestDefaults RequestDefaults `yaml:"request_defaults"`
	Dockerize       bool            `yaml:"dockerize"`
}

type RequestDefaults struct {
	Method  string            `yaml:"method"`
	Url     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

type ServiceConfig struct {
	Type   string                 `yaml:"type"`
	Env    ServiceDefinitionEnv   `yaml:"env"`
	Params map[string]interface{} `yaml:",inline"`
}

type ServiceDefinitionEnv struct {
	EnvStr string
	EnvMap map[string]interface{}
}

func (c *Config) Join(otherConfig *Config) error {
	for otherGeneralCaseName, otherGeneralCase := range otherConfig.GeneralCases {
		for generalCaseName := range c.GeneralCases {
			if generalCaseName == otherGeneralCaseName {
				return fmt.Errorf("general cases can not be re-defined")
			}
		}
		c.GeneralCases[otherGeneralCaseName] = otherGeneralCase
	}

	for otherEnvironmentValueName, otherEnvironmentValue := range otherConfig.Environment {
		for environmentVariableName := range c.Environment {
			if environmentVariableName == otherEnvironmentValueName {
				return fmt.Errorf("environment values can not be re-defined")
			}
		}
		c.Environment[otherEnvironmentValueName] = otherEnvironmentValue
	}

	if otherConfig.Launcher != "" {
		if c.Launcher != "" {
			return fmt.Errorf("launcher can not be re-defined")
		}
		c.Launcher = otherConfig.Launcher
	}

	if otherConfig.Application != nil {
		if c.Application != nil {
			return fmt.Errorf("application config can not be re-defined")
		}
		c.Application = otherConfig.Application
	}

	for otherServiceName, otherService := range otherConfig.Services {
		for serviceName := range c.Services {
			if serviceName == otherServiceName {
				return fmt.Errorf("services can not be re-defined")
			}
		}
		c.Services[otherServiceName] = otherService
	}

	err := c.Cases.Join(otherConfig.Cases, "")
	if err != nil {
		return err
	}

	return nil
}

func (c Config) Validate() error {
	if c.Launcher == "" {
		return fmt.Errorf("launcher not specified")
	}
	if c.Application == nil {
		return fmt.Errorf("application config not specified")
	}
	if c.Application.RequestType == "" {
		return fmt.Errorf("application.request_type not specified")
	}
	if c.Application.RequestDefaults.Method == "" {
		return fmt.Errorf("application.request_defaults.method not specified")
	}
	if c.Application.RequestDefaults.Url == "" {
		return fmt.Errorf("application.request_defaults.url not specified")
	}
	for generalCaseName, generalCase := range c.GeneralCases {
		err := generalCase.Validate()
		if err != nil {
			return fmt.Errorf("general case %q invalid: %v", generalCaseName, err)
		}
	}
	for serviceName, service := range c.Services {
		err := service.Validate()
		if err != nil {
			return fmt.Errorf("service %q invalid: %v", serviceName, err)
		}
	}
	for testCaseName, testCase := range c.Cases {
		err := testCase.Validate()
		if err != nil {
			return fmt.Errorf("test case %q invalid: %v", testCaseName, err)
		}
	}
	return nil
}

func (gc GeneralCase) Validate() error {
	for testCaseName, testCase := range gc.Cases {
		err := testCase.Validate()
		if err != nil {
			return fmt.Errorf("test case %q invalid: %v", testCaseName, err)
		}
	}
	return nil
}

func (sc ServiceConfig) Validate() error {
	if sc.Type == "" {
		return fmt.Errorf("service type not specified")
	}
	return nil
}

func (tcs *TestCases) Join(otherTestCases TestCases, prefix string) error {
	for otherTestCaseName, otherTestCase := range otherTestCases {
		found := false
		for testCaseName, testCase := range *tcs {
			if found {
				break
			}
			if testCaseName == otherTestCaseName {
				err := testCase.Join(otherTestCase, prefix+testCaseName+".")
				if err != nil {
					return fmt.Errorf("unable to join test case %s: %v", prefix+testCaseName, err)
				}
				found = true
			}
		}
		if !found {
			if *tcs == nil {
				*tcs = make(TestCases)
			}
			(*tcs)[otherTestCaseName] = otherTestCase
		}
	}
	return nil
}

func (tcs *TestCases) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var testCases []map[string]*TestCase
	err := unmarshal(&testCases)
	if err != nil {
		return err
	}
	res := make(TestCases)
	for _, testCasesMap := range testCases {
		for testCaseName, testCase := range testCasesMap {
			res[testCaseName] = testCase
		}
	}
	*tcs = res
	return nil
}

func (tc *TestCase) Join(otherTestCase *TestCase, prefix string) error {
	if otherTestCase.PrepareServices != nil {
		if tc.PrepareServices != nil {
			return fmt.Errorf("prepare_services section can not be re-defined")
		}
		tc.PrepareServices = otherTestCase.PrepareServices
	}

	if otherTestCase.CheckServices != nil {
		if tc.CheckServices != nil {
			return fmt.Errorf("check_services section can not be re-defined")
		}
		tc.CheckServices = otherTestCase.CheckServices
	}

	if otherTestCase.Request != nil {
		if tc.Request != nil {
			return fmt.Errorf("request can not be re-defined")
		}
		tc.Request = otherTestCase.Request
	}

	if otherTestCase.ModifyRequest != nil {
		if tc.ModifyRequest != nil {
			return fmt.Errorf("modify_request can not be re-defined")
		}
		tc.ModifyRequest = otherTestCase.ModifyRequest
	}

	if otherTestCase.ExpectedResponse != nil {
		if tc.ExpectedResponse != nil {
			return fmt.Errorf("expected_response can not be re-defined")
		}
		tc.ExpectedResponse = otherTestCase.ExpectedResponse
	}

	if otherTestCase.ExpectedCode != 0 {
		if tc.ExpectedCode != 0 {
			return fmt.Errorf("expected code can not be re-defined")
		}
		tc.ExpectedCode = otherTestCase.ExpectedCode
	}

	if otherTestCase.GeneralCases != nil {
		if tc.GeneralCases != nil {
			return fmt.Errorf("expected code can not be re-defined")
		}
		tc.GeneralCases = otherTestCase.GeneralCases
	}

	if otherTestCase.Only {
		tc.Only = true
	}

	err := tc.Cases.Join(otherTestCase.Cases, prefix+".")
	if err != nil {
		return err
	}

	return nil
}

func (tc TestCase) Validate() error {
	for testCaseName, testCase := range tc.Cases {
		err := testCase.Validate()
		if err != nil {
			return fmt.Errorf("test case %q invalid: %v", testCaseName, err)
		}
	}
	return nil
}

func (e *ServiceDefinitionEnv) UnmarshalYAML(unmarshal func(interface{}) error) error {
	strErr := unmarshal(&e.EnvStr)
	if strErr == nil {
		return nil
	}
	mapErr := unmarshal(&e.EnvMap)
	if mapErr == nil {
		return nil
	}
	return fmt.Errorf("str unmarshal error: %v\nmap unmarshal error: %v", strErr, mapErr)
}
