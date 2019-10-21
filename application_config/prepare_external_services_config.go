package application_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_tests/helper"
	"io"
	"net/http"
	"strings"
)

type PrepareExternalServicesConfig struct {
	prepareMap     map[string]interface{}
	prepareConfigs map[string]IPrepareExternalServiceConfig
}

func (c PrepareExternalServicesConfig) Prepare(externalServices map[string]*ExternalServiceDefinitionSection) error {
	for name, prepareConfig := range c.prepareConfigs {
		externalService := externalServices[name]
		if externalService == nil {
			return fmt.Errorf("unable to find external service %q", name)
		}
		err := externalService.executor.Prepare(prepareConfig)
		if err != nil {
			return fmt.Errorf("unable to prepare external service %q: %v", name, err)
		}
	}
	return nil
}

type IPostgresPrepare interface {
	Prepare(conn *sqlx.DB) error
}

func NewExecPostgresPrepare(exec string) *ExecPostgresPrepare {
	return &ExecPostgresPrepare{
		exec: exec,
	}
}

type ExecPostgresPrepare struct {
	exec string
}

func (pp ExecPostgresPrepare) Prepare(conn *sqlx.DB) error {
	_, err := conn.Exec(pp.exec)
	if err != nil {
		return fmt.Errorf("unable to run %q on postgres: %v", pp.exec, err)
	}
	return nil
}

func NewClearPostgresPrepare() *ClearPostgresPrepare {
	return &ClearPostgresPrepare{}
}

type ClearPostgresPrepare struct {
}

func (pp ClearPostgresPrepare) Prepare(conn *sqlx.DB) error {
	var tableNames []string
	err := conn.Select(&tableNames, "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema' AND tablename != 'schema_migrations'")
	if err != nil {
		return fmt.Errorf("unable to get all table names from db: %v", err)
	}
	_, err = conn.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY", strings.Join(tableNames, ",")))
	if err != nil {
		return fmt.Errorf("unable to truncate tables: %v", err)
	}
	return nil
}

func NewPostgresPrepareConfig(params interface{}) (*PostgresPrepareConfig, error) {
	var prepares []IPostgresPrepare
	preparesList, ok := params.([]interface{})
	if !ok {
		return nil, fmt.Errorf("prepare must be list")
	}
	for _, iprepare := range preparesList {
		prepareYaml, ok := iprepare.(helper.YamlMap)
		if !ok {
			return nil, fmt.Errorf("prepare item must be map")
		}
		prepare := helper.YamlMapToJsonMap(prepareYaml)
		for key, value := range prepare {
			switch key {
			case "exec":
				exec, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf("postgres prepare exec must be string. actual: %v", value)
				}
				prepares = append(prepares, NewExecPostgresPrepare(exec))
			case "clear":
				prepares = append(prepares, NewClearPostgresPrepare())
			default:
				return nil, fmt.Errorf("cannot create postgres prepare %s", key)
			}
		}
	}

	return &PostgresPrepareConfig{
		prepares: prepares,
	}, nil
}

type PostgresPrepareConfig struct {
	prepares []IPostgresPrepare
}

func (ppc PostgresPrepareConfig) Prepare(conn *sqlx.DB) error {
	for i, prepare := range ppc.prepares {
		err := prepare.Prepare(conn)
		if err != nil {
			return fmt.Errorf("unable to prepare postgres %d: %v", i, err)
		}
	}
	return nil
}

type IHttpPrepare interface {
	Prepare(serviceUrl string) error
}

func NewResetCallsHttpPrepare() *ResetCallsHttpPrepare {
	return &ResetCallsHttpPrepare{}
}

type ResetCallsHttpPrepare struct {
}

func (p ResetCallsHttpPrepare) Prepare(serviceUrl string) error {
	resp, err := http.Post(serviceUrl+"__reset_calls", "application/json", nil)
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unsuccessfull status code: %d", resp.StatusCode)
}

func NewConfigHttpPrepare(config map[string]interface{}) *ConfigHttpPrepare {
	return &ConfigHttpPrepare{
		config: config,
	}
}

type ConfigHttpPrepare struct {
	config map[string]interface{}
}

func (p ConfigHttpPrepare) Prepare(serviceUrl string) error {
	var body io.Reader
	if p.config != nil {
		configBytes, err := json.Marshal(p.config)
		if err != nil {
			return fmt.Errorf("unable to marshal config: %v", err)
		}
		body = bytes.NewReader(configBytes)
	}
	resp, err := http.Post(serviceUrl+"__config", "application/json", body)
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unsuccessfull status code: %d", resp.StatusCode)
}

func NewHttpPrepareConfig(params interface{}) (*HttpPrepareConfig, error) {
	prepareYaml, ok := params.(helper.YamlMap)
	if !ok {
		return nil, fmt.Errorf("http prepare config should be map")
	}
	var prepares []IHttpPrepare
	prepare := helper.YamlMapToJsonMap(prepareYaml)
	for action, value := range prepare {
		switch action {
		case "calls":
			if value != nil {
				return nil, fmt.Errorf("can only reset calls in prepare: value should be null")
			}
			prepares = append(prepares, NewResetCallsHttpPrepare())
		case "config":
			if value == nil {
				prepares = append(prepares, NewConfigHttpPrepare(nil))
				break
			}
			valueMap, ok := value.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("http service config should be map, but now it is %T (%#v)", value, value)
			}
			prepares = append(prepares, NewConfigHttpPrepare(valueMap))
		default:
			return nil, fmt.Errorf("invalid http prepare action %q", action)
		}
	}
	return &HttpPrepareConfig{
		prepares: prepares,
	}, nil
}

type HttpPrepareConfig struct {
	prepares []IHttpPrepare
}

func (hpc HttpPrepareConfig) Prepare(serviceUrl string) error {
	for i, prepare := range hpc.prepares {
		err := prepare.Prepare(serviceUrl)
		if err != nil {
			return fmt.Errorf("unable to prepare http %d: %v", i, err)
		}
	}
	return nil
}
