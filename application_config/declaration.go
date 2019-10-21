package application_config

import (
	"integration_tests/docker_compose_config"
	"time"
)

type ExternalServiceDefinitionEnv struct {
	EnvStr string
	EnvMap map[string]interface{}
}

type IExternalServiceDefinitionExecutor interface {
	Start() error
	Generate(dstDirectory string, name string, applicationService *docker_compose_config.DockerComposeService) (serviceName string, result *docker_compose_config.DockerComposeService, err error)
	Prepare(prepareConfig IPrepareExternalServiceConfig) error
	Check(checkConfig ICheckExternalServiceConfig, saveResult FnResultSaver, variables map[string]interface{}) error
}

type ExternalServiceDefinitionSection struct {
	Type       string                       `yaml:"type"`
	Env        ExternalServiceDefinitionEnv `yaml:"env"`
	Params     map[string]interface{}       `yaml:",inline"`
	BindToPort string                       `yaml:"-"`
	executor   IExternalServiceDefinitionExecutor
}

type BeforeEachSection struct {
	PrepareExternalServices PrepareExternalServicesConfig `yaml:"prepare_external_services"`
}

type TestCase struct {
	// parsed manually
	Name                    string                        `yaml:"-"`
	PrepareExternalServices PrepareExternalServicesConfig `yaml:"prepare_external_services"`
	RequestQuery            string                        `yaml:"request_query"`
	// if ModifyRequest is empty then default headers will be added
	// else headers will not be added and ModifyRequest will be executed
	ModifyRequest *string `yaml:"modify_request"`
	// expected response can not be defined (so field will be <nil>)
	// also it can be defined as `null` (so field will be pointer to <nil>)
	// also it can be map (so field will be pointer to map)
	ExpectedResponse      *map[interface{}]interface{} `yaml:"expected_response"`
	ExpectedCode          int                          `yaml:"expected_code"`
	CheckExternalServices CheckExternalServicesConfig  `yaml:"check_external_services"`
	Only                  bool                         `yaml:"only"`
}

type TestSection struct {
	BeforeEach   BeforeEachSection `yaml:"before_each"`
	GeneralQuery string            `yaml:"general_query"`
	// parsed manually
	Cases []TestCase `yaml:"-"`
}

type EnvironmentInitializationSection struct {
	SetClocks               *time.Time                    `yaml:"-"`
	PrepareExternalServices PrepareExternalServicesConfig `yaml:"prepare_external_services"`
}

type ApplicationConfig struct {
	ExternalServices map[string]*ExternalServiceDefinitionSection `yaml:"external_services"`
	Tests            map[string]TestSection                       `yaml:"tests"`
	InitEnvironment  EnvironmentInitializationSection             `yaml:"init_environment"`
	GeneralCases     []TestCase                                   `yaml:"-"`
}
