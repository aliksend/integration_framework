package testing

import (
	"fmt"
	"integration_framework/application_config"
	"integration_framework/plugins"
	"sort"
)

func ParseConfig(config *application_config.Config) (*ParsedConfig, error) {
	res := &ParsedConfig{
		config:    config,
		onlyCases: getOnlyCases(config.Cases, ""),
	}
	err := res.createEnvironmentInitializers()
	if err != nil {
		return nil, fmt.Errorf("unable to create environment initializers: %v", err)
	}
	err = res.createServices()
	if err != nil {
		return nil, fmt.Errorf("unable to create services: %v", err)
	}
	err = res.createRequesterConstructor()
	if err != nil {
		return nil, fmt.Errorf("unable to create requester constructor: %v", err)
	}
	err = res.createGeneralCasesRequesters()
	if err != nil {
		return nil, fmt.Errorf("unable to create general cases requesters: %v", err)
	}
	err = res.createTesters(config.Cases, nil, nil, "", false)
	if err != nil {
		return nil, fmt.Errorf("unable to create testers: %v", err)
	}

	return res, nil
}

type ParsedConfig struct {
	Services                map[string]plugins.IService
	Testers                 []Tester
	EnvironmentInitializers []plugins.IEnvironmentInitializer

	onlyCases              []string
	config                 *application_config.Config
	requesterConstructor   plugins.RequesterConstructor
	generalCasesRequesters map[string]plugins.IRequester
}

func (pc *ParsedConfig) createEnvironmentInitializers() error {
	for environmentInitializerName, environmentInitializerParams := range pc.config.Environment {
		environmentInitializer, err := plugins.NewEnvironmentInitializer(environmentInitializerName, environmentInitializerParams)
		if err != nil {
			return fmt.Errorf("unable to create environment initializer %q: %v", environmentInitializerName, err)
		}
		pc.EnvironmentInitializers = append(pc.EnvironmentInitializers, environmentInitializer)
	}
	return nil
}

func (pc *ParsedConfig) createServices() error {
	pc.Services = make(map[string]plugins.IService)
	port := 9000
	var serviceNames []string
	for serviceName := range pc.config.Services {
		serviceNames = append(serviceNames, serviceName)
	}
	sort.Strings(serviceNames)
	for _, serviceName := range serviceNames {
		serviceConfig := pc.config.Services[serviceName]
		service, err := plugins.NewService(serviceName, serviceConfig.Type, port, serviceConfig.Env, serviceConfig.Params)
		if err != nil {
			return fmt.Errorf("unable to create service %q: %v", serviceName, err)
		}
		pc.Services[serviceName] = service
		port++
	}
	return nil
}

func (pc *ParsedConfig) createRequesterConstructor() error {
	var ok bool
	pc.requesterConstructor, ok = plugins.GetRequesterConstructor(pc.config.Application.RequestType)
	if !ok {
		return fmt.Errorf("unable to find requester for type %q", pc.config.Application.RequestType)
	}
	return nil
}

func (pc *ParsedConfig) createGeneralCasesRequesters() error {
	pc.generalCasesRequesters = make(map[string]plugins.IRequester)
	for generalCaseName, generalCase := range pc.config.GeneralCases {
		for testCaseName, testCase := range generalCase.Cases {
			if testCase.Request != nil {
				generalCaseTestName := generalCaseName + " " + testCaseName
				requester, err := pc.requesterConstructor(testCase.Request, pc.config.Application.RequestDefaults)
				if err != nil {
					return fmt.Errorf("unable to create requester for general case %q: %v", generalCaseTestName, err)
				}
				pc.generalCasesRequesters[generalCaseTestName] = requester
			}
		}
	}
	return nil
}

func getOnlyCases(cases application_config.TestCases, casePrefix string) (res []string) {
	for testCaseName, testCase := range cases {
		if testCase.Only {
			res = append(res, casePrefix+testCaseName)
		}
		if testCase.Cases != nil {
			onlyCases := getOnlyCases(testCase.Cases, casePrefix+testCaseName+" ")
			res = append(res, onlyCases...)
		}
	}
	return
}

func (pc *ParsedConfig) createTesters(cases application_config.TestCases, parentPreparers []plugins.IServicePreparer, parentCheckers []plugins.IServiceChecker, casePrefix string, ignoreOnly bool) error {
	for tCN, testCase := range cases {
		testCaseName := casePrefix + tCN

		if testCase.Skip {
			continue
		}

		// process only-cases
		// if this test not exists in onlyCases then preparers and checkers should be created but only for children
		allowedToProcess := true
		if !ignoreOnly && len(pc.onlyCases) != 0 {
			found := false
			for _, onlyCaseName := range pc.onlyCases {
				if testCaseName == onlyCaseName {
					found = true
					break
				}
			}
			allowedToProcess = found
		}

		// create preparers
		servicePreparers := append([]plugins.IServicePreparer{}, parentPreparers...)
		for serviceName, servicePreparerParams := range testCase.PrepareServices {
			service, ok := pc.Services[serviceName]
			if !ok {
				return fmt.Errorf("unable to find service with name %q", serviceName)
			}
			// fmt.Printf("-- create service %q preparer %#v for test case %q\n", serviceName, servicePreparerParams, testCaseName)
			servicePreparer, err := service.Preparer(servicePreparerParams)
			if err != nil {
				return fmt.Errorf("unable to create preparer for service %q: %v", serviceName, err)
			}
			servicePreparers = append(servicePreparers, servicePreparer)
		}

		// create checkers
		serviceCheckers := append([]plugins.IServiceChecker{}, parentCheckers...)
		for _, checkServicesMap := range testCase.CheckServices {
			for serviceName, serviceCheckerParams := range checkServicesMap {
				service, ok := pc.Services[serviceName]
				if !ok {
					return fmt.Errorf("unable to find service with name %q", serviceName)
				}
				serviceChecker, err := service.Checker(serviceCheckerParams)
				if err != nil {
					return fmt.Errorf("unable to create checker for service %q: %v", serviceName, err)
				}
				serviceCheckers = append(serviceCheckers, serviceChecker)
			}
		}

		if allowedToProcess {
			// create testers for general cases
			if testCase.GeneralCases != nil {
				err := pc.createGeneralTesters(testCaseName, servicePreparers, serviceCheckers, *testCase.GeneralCases)
				if err != nil {
					return fmt.Errorf("unable to create general cases for %q: %v", testCaseName, err)
				}
			}

			// create tester for this test case
			if testCase.Request != nil || testCase.ModifyRequest != nil || testCase.ExpectedCode != 0 || testCase.ExpectedResponse != nil {
				if testCase.Request == nil {
					return fmt.Errorf("unable to create tester %q: request should be set", testCaseName)
				}
				requester, err := pc.requesterConstructor(testCase.Request, pc.config.Application.RequestDefaults)
				if err != nil {
					return fmt.Errorf("unable to create request for tester %q: %v", testCaseName, err)
				}
				// fmt.Printf("**** created requster for test case %q: %#v\n", testCaseName, requester)
				err = pc.createTester(testCaseName, servicePreparers, serviceCheckers, testCase, requester)
				if err != nil {
					return fmt.Errorf("unable to create tester %q: %v", testCaseName, err)
				}
			}
		}

		// create testers for sub-cases
		if testCase.Cases != nil {
			// ignore `only` flag in all children if test case is allowed to process
			err := pc.createTesters(testCase.Cases, servicePreparers, serviceCheckers, testCaseName+" ", allowedToProcess)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (pc *ParsedConfig) createGeneralTesters(testCaseName string, servicePreparers []plugins.IServicePreparer, serviceCheckers []plugins.IServiceChecker, selector application_config.GeneralCasesSelector) (err error) {
	for generalCaseName, generalCase := range pc.config.GeneralCases {
		if !generalCase.AutoInclude {
			include := false
			for _, inlcudeWithName := range selector.Include {
				if inlcudeWithName == generalCaseName {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}

		excluded := false
		for _, excludeWithName := range selector.Exclude {
			if excludeWithName == generalCaseName {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		for generalTestCaseName, generalTestCase := range generalCase.Cases {
			selectorRequester, err := pc.requesterConstructor(selector.Request, pc.config.Application.RequestDefaults)
			if err != nil {
				return fmt.Errorf("unable to create requester for selector's request for general case %q for test case %q: %v", generalCaseName+" "+generalTestCaseName, testCaseName, err)
			}
			generalCaseRequester, ok := pc.generalCasesRequesters[generalCaseName+" "+generalTestCaseName]
			if !ok {
				return fmt.Errorf("unable to find requester for general case %q for test case %q: %v", generalCaseName+" "+generalTestCaseName, testCaseName, err)
			}
			requester, err := generalCaseRequester.Join(selectorRequester)
			if err != nil {
				return fmt.Errorf("unable to join requesters for general case %q for test case %q: %v", generalCaseName+" "+generalTestCaseName, testCaseName, err)
			}
			err = pc.createTester(testCaseName+" "+generalCaseName+" "+generalTestCaseName, servicePreparers, serviceCheckers, generalTestCase, requester)
			if err != nil {
				return fmt.Errorf("unable to create tester for general case %q for test case %q: %v", generalCaseName+" "+generalTestCaseName, testCaseName, err)
			}
		}
	}
	return
}
