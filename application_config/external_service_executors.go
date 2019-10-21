package application_config

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_tests/docker_compose_config"
	"integration_tests/helper"
	"io/ioutil"
	"os"
)

const (
	postgresDbName = "test"
)

type IPrepareExternalServiceConfig interface{}
type ICheckExternalServiceConfig interface{}

func NewPostgresExternalServiceExecutor(definition *ExternalServiceDefinitionSection) *PostgresExternalServiceExecutor {
	return &PostgresExternalServiceExecutor{
		definition: definition,
	}
}

type PostgresExternalServiceExecutor struct {
	name       string
	definition *ExternalServiceDefinitionSection
	conn       *sqlx.DB
}

func (g *PostgresExternalServiceExecutor) Generate(dstDirectory string, name string, applicationService *docker_compose_config.DockerComposeService) (string, *docker_compose_config.DockerComposeService, error) {
	g.name = name
	serviceName := "postgres_" + name
	service := docker_compose_config.DockerComposeService{
		Image:   "postgres:9.6",
		Restart: "always",
		Environment: map[string]string{
			"POSTGRES_DB": postgresDbName,
		},
		Ports: []string{
			fmt.Sprintf("%s:5432", g.definition.BindToPort),
		},
	}
	applicationService.DependsOn = append(applicationService.DependsOn, serviceName)
	if g.definition.Env.EnvStr != "" {
		applicationService.Environment[g.definition.Env.EnvStr] = fmt.Sprintf("postgres://postgres:postgres@%s/%s?sslmode=disable", serviceName, postgresDbName)
	} else if g.definition.Env.EnvMap != nil {
		err := helper.FillEnvironment(&applicationService.Environment, g.definition.Env.EnvMap, map[string]string{
			"login":    "postgres",
			"password": "postgres",
			"schema":   "postgres",
			"host":     serviceName,
			"port":     "5432",
			"db":       postgresDbName,
		})
		if err != nil {
			return "", nil, fmt.Errorf("unable to fill environment for postgres %q: %v", name, err)
		}
	}
	return serviceName, &service, nil
}

func (g *PostgresExternalServiceExecutor) Start() error {
	fmt.Println(">> postgres start!!", g.definition.BindToPort, postgresDbName)
	conn, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://postgres:postgres@localhost:%s/%s?sslmode=disable", g.definition.BindToPort, postgresDbName))
	if err != nil {
		return fmt.Errorf("unable to connect to postgres: %v", err)
	}
	g.conn = conn
	return nil
}

func (g *PostgresExternalServiceExecutor) Prepare(prepareConfig IPrepareExternalServiceConfig) error {
	postgresPrepareConfig, ok := prepareConfig.(*PostgresPrepareConfig)
	if !ok {
		return fmt.Errorf("unable to prepare postgres external service with this prepare config: %#v", prepareConfig)
	}
	err := postgresPrepareConfig.Prepare(g.conn)
	if err != nil {
		return fmt.Errorf("unable to prepare postgres external service: %v", err)
	}
	return nil
}

func (g *PostgresExternalServiceExecutor) Check(checkConfig ICheckExternalServiceConfig, saveResult FnResultSaver, variables map[string]interface{}) error {
	postgresCheckConfig, ok := checkConfig.(*PostgresCheckConfig)
	if !ok {
		return fmt.Errorf("unable to check postgres external service with this check config: %#v", checkConfig)
	}
	err := postgresCheckConfig.Check(g.conn, saveResult, variables)
	if err != nil {
		return fmt.Errorf("unable to check postgres external service: %v", err)
	}
	return nil
}

func NewHttpExternalServiceExecutor(definition *ExternalServiceDefinitionSection) *HttpExternalServiceExecutor {
	return &HttpExternalServiceExecutor{
		definition: definition,
	}
}

type HttpExternalServiceExecutor struct {
	definition *ExternalServiceDefinitionSection
}

func (g *HttpExternalServiceExecutor) Generate(dstDirectory string, name string, applicationService *docker_compose_config.DockerComposeService) (string, *docker_compose_config.DockerComposeService, error) {
	serviceName := "http_" + name
	servicePort := "8080"
	simpleServerConfig := make(map[string]interface{})
	m, ok := g.definition.Params["routes"].(helper.YamlMap)
	if ok {
		simpleServerConfig["routes"] = helper.YamlMapToJsonMap(m)
	}
	simpleServerConfig["service_name"] = name
	simpleServerConfigBytes, err := json.Marshal(simpleServerConfig)
	if err != nil {
		return "", nil, fmt.Errorf("unable to marshal simple server %s config: %v", name, err)
	}
	serviceConfigFilename := dstDirectory + "/" + serviceName + ".simple-server.conf"
	err = ioutil.WriteFile(serviceConfigFilename, simpleServerConfigBytes, os.ModePerm)
	if err != nil {
		return "", nil, fmt.Errorf("unable to write simple-server.conf file: %v", err)
	}
	service := docker_compose_config.DockerComposeService{
		Build: docker_compose_config.DockerComposeServiceBuild{
			Context: "../../simple_server",
		},
		Volumes: []string{
			fmt.Sprintf("%s:%s", "../"+serviceConfigFilename, "/config.json"),
		},
		WorkingDir: "/app",
		Restart:    "on-failure",
		Environment: map[string]string{
			"PORT": servicePort,
		},
		Ports: []string{
			fmt.Sprintf("%s:%s", g.definition.BindToPort, servicePort),
		},
	}
	applicationService.DependsOn = append(applicationService.DependsOn, serviceName)
	if g.definition.Env.EnvStr != "" {
		applicationService.Environment[g.definition.Env.EnvStr] = fmt.Sprintf("http://%s:%s/", serviceName, servicePort)
	} else if g.definition.Env.EnvMap != nil {
		err := helper.FillEnvironment(&applicationService.Environment, g.definition.Env.EnvMap, map[string]string{
			"schema": "http",
			"host":   serviceName,
			"port":   servicePort,
		})
		if err != nil {
			return "", nil, fmt.Errorf("unable to fill environment for http %q: %v", name, err)
		}
	}
	return serviceName, &service, nil
}

func (g *HttpExternalServiceExecutor) Start() error {
	return nil
}

func (g *HttpExternalServiceExecutor) Prepare(prepareConfig IPrepareExternalServiceConfig) error {
	httpPrepareConfig, ok := prepareConfig.(*HttpPrepareConfig)
	if !ok {
		return fmt.Errorf("unable to prepare http external service with this prepare config: %#v", prepareConfig)
	}
	err := httpPrepareConfig.Prepare(fmt.Sprintf("http://localhost:%s/", g.definition.BindToPort))
	if err != nil {
		return fmt.Errorf("unable to prepare http external service: %v", err)
	}
	return nil
}

func (g *HttpExternalServiceExecutor) Check(checkConfig ICheckExternalServiceConfig, saveResult FnResultSaver, variables map[string]interface{}) error {
	httpCheckConfig, ok := checkConfig.(*HttpCheckConfig)
	if !ok {
		return fmt.Errorf("unable to check http external service with this check config: %#v", checkConfig)
	}
	err := httpCheckConfig.Check(fmt.Sprintf("http://localhost:%s/", g.definition.BindToPort))
	if err != nil {
		return fmt.Errorf("unable to check http external service: %v", err)
	}
	return nil
}
