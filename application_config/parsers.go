package application_config

import (
	"fmt"
	"github.com/jinzhu/copier"
	"time"
	"sort"
)

func (s *ExternalServiceDefinitionSection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias ExternalServiceDefinitionSection
	alias := Alias{}
	if err := unmarshal(&alias); err != nil {
		return err
	}
	*s = ExternalServiceDefinitionSection(alias)
	switch s.Type {
	case "postgres":
		s.executor = NewPostgresExternalServiceExecutor(s)
	case "http":
		s.executor = NewHttpExternalServiceExecutor(s)
	default:
		return fmt.Errorf("unsupported type: %s", s.Type)
	}
	return nil
}

func (tc *TestCase) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias TestCase
	alias := Alias{}
	if err := unmarshal(&alias); err != nil {
		return err
	}
	*tc = TestCase(alias)

	if tc.ExpectedCode == 0 {
		tc.ExpectedCode = 200
	}

	return nil
}

func (c *ApplicationConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias ApplicationConfig
	alias := struct {
		Alias        `yaml:",inline"`
		GeneralCases []map[string]TestCase `yaml:"general_cases"`
	}{}
	if err := unmarshal(&alias); err != nil {
		return err
	}
	*c = ApplicationConfig(alias.Alias)

	serviceTypeByName := func(name string) (string, error) {
		service, ok := c.ExternalServices[name]
		if !ok {
			return "", fmt.Errorf("service not found for name %q", name)
		}
		return service.Type, nil
	}
	err := c.InitEnvironment.PrepareExternalServices.ParseMap(serviceTypeByName)
	if err != nil {
		return fmt.Errorf("unable to check prepare external services config: %v", err)
	}
	for _, cases := range alias.GeneralCases {
		for name, oneCase := range cases {
			oneCase.Name = name
			c.GeneralCases = append(c.GeneralCases, oneCase)
		}
	}
	for sectionName, section := range c.Tests {
		err := section.BeforeEach.PrepareExternalServices.ParseMap(serviceTypeByName)
		if err != nil {
			return fmt.Errorf("unable to parse prepare external services config: %v", err)
		}

		var cases []TestCase
		for _, generalCase := range c.GeneralCases {
			extendedGeneralCase := TestCase{}
			err := copier.Copy(&extendedGeneralCase, &generalCase)
			if err != nil {
				return fmt.Errorf("unable to copy general case: %v", err)
			}
			extendedGeneralCase.RequestQuery = section.GeneralQuery
			cases = append(cases, extendedGeneralCase)
		}
		cases = append(cases, section.Cases...)
		section.Cases = cases

		for i, oneCase := range section.Cases {
			err := oneCase.PrepareExternalServices.ParseMap(serviceTypeByName)
			if err != nil {
				return fmt.Errorf("unable to parse prepare external services config: %v", err)
			}
			err = oneCase.CheckExternalServices.ParseMap(serviceTypeByName)
			if err != nil {
				return fmt.Errorf("unable to parse check external services config: %v", err)
			}
			section.Cases[i] = oneCase
		}

		c.Tests[sectionName] = section
	}

	// назначаем порты сервисам в алфавитном порядке чтобы каждый раз сервису назначался один и тот же порт чтобы зря не перезапускать docker-compose
	var externalServicesNames []string
	for serviceName := range c.ExternalServices {
		externalServicesNames = append(externalServicesNames, serviceName)
	}
	sort.Strings(externalServicesNames)

	bindToPort := 2000
	for _, serviceName := range externalServicesNames {
		c.ExternalServices[serviceName].BindToPort = fmt.Sprintf("%d", bindToPort)
		bindToPort++
	}

	return nil
}

func (s *EnvironmentInitializationSection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias EnvironmentInitializationSection
	dst := struct {
		Alias     `yaml:",inline"`
		SetClocks string `yaml:"set_clocks"`
	}{}
	err := unmarshal(&dst)
	if err != nil {
		return err
	}
	*s = EnvironmentInitializationSection(dst.Alias)

	if dst.SetClocks == "" {
		s.SetClocks = nil
	} else {
		parsedTime, err := time.Parse("2006-01-02T15:04:05Z", dst.SetClocks)
		if err != nil {
			return fmt.Errorf("unable to parse time")
		}
		s.SetClocks = &parsedTime
	}
	return nil
}

func (s *TestSection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias TestSection
	alias := struct {
		Alias `yaml:",inline"`

		// to be able to define like
		// - test case key:
		//     request_query: ...
		// - other test case key:
		//     request_query: ...
		Cases []map[string]TestCase `yaml:"cases"`
	}{}
	if err := unmarshal(&alias); err != nil {
		return err
	}
	*s = TestSection(alias.Alias)

	for _, cases := range alias.Cases {
		for name, oneCase := range cases {
			oneCase.Name = name
			s.Cases = append(s.Cases, oneCase)
		}
	}

	return nil
}

func (e *ExternalServiceDefinitionEnv) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

func (c *PrepareExternalServicesConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	prepareMap := make(map[string]interface{})
	err := unmarshal(&prepareMap)
	if err != nil {
		return err
	}
	c.prepareMap = prepareMap
	return nil
}

func (c *PrepareExternalServicesConfig) ParseMap(serviceTypeByName func(string) (string, error)) error {
	if c.prepareMap == nil {
		return nil
	}
	prepareConfigs := make(map[string]IPrepareExternalServiceConfig)
	for serviceName, prepareConfig := range c.prepareMap {
		serviceType, err := serviceTypeByName(serviceName)
		if err != nil {
			return fmt.Errorf("unable to get type for service %q: %v", serviceName, err)
		}
		switch serviceType {
		case "postgres":
			prepareConfigs[serviceName], err = NewPostgresPrepareConfig(prepareConfig)
			if err != nil {
				return fmt.Errorf("unable to create prepare config for service %q (%s): %v", serviceName, serviceType, err)
			}
		case "http":
			prepareConfigs[serviceName], err = NewHttpPrepareConfig(prepareConfig)
			if err != nil {
				return fmt.Errorf("unable to create prepare config for service %q (%s): %v", serviceName, serviceType, err)
			}
		default:
			return fmt.Errorf("unable to prepare service %q with type %q", serviceName, serviceType)
		}
	}
	c.prepareConfigs = prepareConfigs
	return nil
}

func (c *CheckExternalServicesConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	checkMap := make(map[string]interface{})
	err := unmarshal(&checkMap)
	if err != nil {
		return err
	}
	c.checkMap = checkMap
	return nil
}

func (c *CheckExternalServicesConfig) ParseMap(serviceTypeByName func(string) (string, error)) error {
	if c.checkMap == nil {
		return nil
	}
	checkConfigs := make(map[string]ICheckExternalServiceConfig)
	for serviceName, checkConfig := range c.checkMap {
		serviceType, err := serviceTypeByName(serviceName)
		if err != nil {
			return fmt.Errorf("unable to get type for service %q: %v", serviceName, err)
		}
		switch serviceType {
		case "postgres":
			checkConfigs[serviceName], err = NewPostgresCheckConfig(checkConfig)
			if err != nil {
				return fmt.Errorf("unable to create check config for service %q (%s): %v", serviceName, serviceType, err)
			}
		case "http":
			checkConfigs[serviceName], err = NewHttpCheckConfig(checkConfig)
			if err != nil {
				return fmt.Errorf("unable to create check config for service %q (%s): %v", serviceName, serviceType, err)
			}
		default:
			return fmt.Errorf("unable to check service %q with type %q", serviceName, serviceType)
		}
	}
	c.checkConfigs = checkConfigs
	return nil
}
