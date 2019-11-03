package http_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func NewConfigPrepare(config map[string]interface{}) *ConfigPrepare {
	return &ConfigPrepare{
		config: config,
	}
}

type ConfigPrepare struct {
	config map[string]interface{}
}

func (p ConfigPrepare) Prepare(serviceUrl string) error {
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
