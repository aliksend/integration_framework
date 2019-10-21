package main

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"integration_tests/application_config"
	"integration_tests/docker_compose_launcher"
	"integration_tests/helper"
	"integration_tests/testing"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

const (
	dstDirectory    = "./integration_environment"
	exitAfterOneRun = true
)

var existingDockerComposeBytes []byte

func generateDockerComposeConfig() (*application_config.ApplicationConfig, bool, error) {
	err := helper.EnsureDirectory(dstDirectory)
	if err != nil {
		return nil, true, fmt.Errorf("unable to ensure directory: %v", err)
	}

	config, err := application_config.LoadConfig("./integration_environment.yml")
	if err != nil {
		return nil, true, fmt.Errorf("unable to load config: %v", err)
	}
	dockerComposeConfig, err := config.DockerComposeConfig(dstDirectory, config)
	if err != nil {
		return nil, true, fmt.Errorf("unable to generate docker-compose config: %v", err)
	}
	dockerComposeBytes, err := yaml.Marshal(dockerComposeConfig)
	if err != nil {
		return nil, true, fmt.Errorf("unable to marshal docker compose config: %v", err)
	}
	if existingDockerComposeBytes != nil {
		dockerComposeConfigUpdated, err := isEqualYamlMaps(existingDockerComposeBytes, dockerComposeBytes)
		if err != nil {
			return nil, true, fmt.Errorf("unable to check docker compose config equality: %v", err)
		}
		if dockerComposeConfigUpdated {
			return config, false, nil
		}
	}
	existingDockerComposeBytes = dockerComposeBytes
	err = ioutil.WriteFile(dstDirectory+"/docker-compose.generated-test.yml", dockerComposeBytes, os.ModePerm)
	if err != nil {
		return nil, true, fmt.Errorf("unable to write docker compose config: %v", err)
	}

	return config, true, nil
}

func isEqualYamlMaps(actual []byte, expected []byte) (bool, error) {
	// fmt.Printf(">> is equal yaml %s %s", actual, expected)
	var actualValue, expectedValue map[interface{}]interface{}
	err := yaml.Unmarshal(actual, &actualValue)
	if err != nil {
		return false, fmt.Errorf("unable to unmarshal actual: %v", err)
	}
	err = yaml.Unmarshal(expected, &expectedValue)
	if err != nil {
		return false, fmt.Errorf("unable to unmarshal expected: %v", err)
	}
	err = testing.IsEqual(helper.YamlMapToJsonMap(actualValue), helper.YamlMapToJsonMap(expectedValue))
	if err != nil {
		fmt.Println(">> not equal!", err)
	}
	return err != nil, nil
}

func launchApplication(config *application_config.ApplicationConfig) (*exec.Cmd, error) {
	fmt.Println("---> launch app")
	cmd, err := docker_compose_launcher.LaunchApplication(dstDirectory)
	if err != nil {
		return nil, fmt.Errorf("unable to launch docker compose: %v", err)
	}
	fmt.Println("---> launched")
	for {
		portAvailable := docker_compose_launcher.IsPortAvailable("localhost", "8080")
		if portAvailable {
			fmt.Println("---> port available 8080")
			break
		}
		if shutdownRequested {
			break
		}
		time.Sleep(5 * time.Second)
	}
	for _, serviceDefinition := range config.ExternalServices {
		if serviceDefinition.Type != "http" {
			continue
		}
		for {
			portAvailable := docker_compose_launcher.IsPortAvailable("localhost", serviceDefinition.BindToPort)
			if portAvailable {
				fmt.Println("---> port available", serviceDefinition.BindToPort)
				break
			}
			if shutdownRequested {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
	fmt.Println("---> return from launch application")
	return cmd, nil
}

var shutdownRequested bool
var globalCmd *exec.Cmd

func main() {
	rand.Seed(time.Now().UnixNano())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("\n\n----> Received an interrupt")
		shutdown()
	}()

	exitCode, err := start()
	fmt.Println(">> returned from start", exitCode, err)
	if err != nil {
		log.Println(err)
	}

	time.Sleep(2 * time.Second)
	shutdown()

	os.Exit(exitCode)
}

func shutdown() {
	shutdownRequested = true
	if globalCmd != nil {
		err := globalCmd.Process.Signal(os.Interrupt)
		if err != nil {
			log.Println("unable to interrupt cmd: %v", err)
		}
		err = globalCmd.Wait()
		if err != nil {
			log.Printf("unable to wait for cmd: %v", err)
		}
		globalCmd = nil
	}
}

type OnlyTestCase struct {
	Section   string
	TestIndex int
}

type FailedTestCase struct {
	Section   string
	TestIndex int
	Error     error
}

var configUpdated bool

func startWatcher(pathToWatch string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("unable to create watcher: %v", err)
	}

	if err := watcher.Add(pathToWatch); err != nil {
		return fmt.Errorf("unable to add file to watcher: %v", err)
	}

	go func() {
		defer watcher.Close()
		for {
			if shutdownRequested {
				return
			}

			select {
			// watch for events
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Printf("FSNOTIFY EVENT! %#v\n", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("--> config update detected!")
					configUpdated = true
				}

				// watch for errors
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("watcher error", err)
			}
		}
	}()

	return nil
}

func start() (exitCode int, err error) {
	err = startWatcher("./integration_environment.yml")
	if err != nil {
		return 1, fmt.Errorf("unable to start watcher: %v", err)
	}

	for {
		fmt.Println("--> start The Cycle")
		config, dockerComposeConfigUpdated, err := generateDockerComposeConfig()
		if err != nil {
			return 1, fmt.Errorf("unable to generate docker-compose config: %v", err)
		}
		if dockerComposeConfigUpdated {
			fmt.Println("--> docker compose config updated")
			if globalCmd != nil {
				fmt.Println("--> shutdown")
				shutdown()
				shutdownRequested = false
			}
			globalCmd, err = launchApplication(config)
			if err != nil {
				return 1, fmt.Errorf("unable to launch docker-compose application: %v", err)
			}
		}

		for serviceName, service := range config.ExternalServices {
			err := service.Start()
			if err != nil {
				return 1, fmt.Errorf("unable to start service %q: %v", serviceName, err)
			}
		}

		exitCode = runTests(config)

		if exitAfterOneRun {
			return exitCode, nil
		}

		for {
			time.Sleep(5 * time.Second)
			fmt.Println("--> wait for config updated")
			if configUpdated {
				configUpdated = false
				break
			}
			if shutdownRequested {
				return 0, nil
			}
		}
	}
}

func runTests(config *application_config.ApplicationConfig) (exitCode int) {
	var onlyCases []OnlyTestCase
	for name, section := range config.Tests {
		for i, testCase := range section.Cases {
			if testCase.Only {
				onlyCases = append(onlyCases, OnlyTestCase{
					Section:   name,
					TestIndex: i,
				})
			}
		}
	}

	var failedTests []FailedTestCase
	for sectionName, section := range config.Tests {
		fmt.Printf("==== %s\n", sectionName)
		for testCaseIndex, testCase := range section.Cases {
			fmt.Printf("---- %s\n", testCase.Name)
			if len(onlyCases) != 0 {
				found := false
				for _, onlyCase := range onlyCases {
					if onlyCase.Section == sectionName && onlyCase.TestIndex == testCaseIndex {
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("skipped\n")
					continue
				}
			}

			err := runTest(config, section, testCase)
			if err != nil {
				fmt.Printf("====> test failed: %v\n", err)
				failedTests = append(failedTests, FailedTestCase{
					Section:   sectionName,
					TestIndex: testCaseIndex,
					Error:     err,
				})
			} else {
				fmt.Printf("====> test passed\n")
			}
		}
	}

	if len(failedTests) == 0 {
		fmt.Println("==== all tests passed")
	} else {
		fmt.Printf("%d test(s) fails\n", len(failedTests))
		for i, failedTest := range failedTests {
			fmt.Printf("==== #%d\n", i)
			fmt.Printf("%s\n", failedTest.Section)
			fmt.Printf("%s\n", config.Tests[failedTest.Section].Cases[failedTest.TestIndex].Name)
			fmt.Printf("%v\n", failedTest.Error)
		}
		exitCode = 2
	}

	return
}

func runTest(config *application_config.ApplicationConfig, testSection application_config.TestSection, testCase application_config.TestCase) error {
	err := config.InitEnvironment.Init(config.ExternalServices)
	if err != nil {
		return fmt.Errorf("unable to initialize environment: %v", err)
	}
	err = testSection.BeforeEach.PrepareExternalServices.Prepare(config.ExternalServices)
	if err != nil {
		return fmt.Errorf("unable to prepare external services in before_each: %v", err)
	}
	err = testCase.PrepareExternalServices.Prepare(config.ExternalServices)
	if err != nil {
		return fmt.Errorf("unable to prepare external services in case: %v", err)
	}

	response, err := testing.MakeRequest(testCase.RequestQuery, testCase.ModifyRequest)
	if err != nil {
		return fmt.Errorf("unable to make response: %v", err)
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %v", err)
	}

	if testCase.ExpectedResponse != nil {
		// expected response defined ...
		if *testCase.ExpectedResponse == nil {
			// ... but it defined like "null", not like map
			if len(responseBody) != 0 {
				return fmt.Errorf("invalid response: expected empty response, but actual response is %s\n", responseBody)
			}
		} else {
			// ... and it defined like map
			expectedBody := helper.YamlMapToJsonMap(*testCase.ExpectedResponse)
			actualBody := make(map[string]interface{})
			err := json.Unmarshal(responseBody, &actualBody)
			if err != nil {
				return fmt.Errorf("unable to unmarshal response body %s: %v", responseBody, err)
			}
			err = testing.IsEqual(actualBody, expectedBody)
			if err != nil {
				return fmt.Errorf("invalid response: %v\n", err)
			}
		}
	}
	if response.StatusCode != testCase.ExpectedCode {
		return fmt.Errorf("invalid response status code: expected %d to equal %d\n", response.StatusCode, testCase.ExpectedCode)
	}

	err = testCase.CheckExternalServices.Check(config.ExternalServices)
	if err != nil {
		return fmt.Errorf("unable to check external services: %v", err)
	}

	return nil
}
