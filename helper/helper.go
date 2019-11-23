package helper

import (
	"fmt"
	"net/http"
	"os"
)

const TimeLayout = "2006-01-02T15:04:05Z"

func EnsureDirectory(pathToDirectory string) error {
	info, err := os.Stat(pathToDirectory)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("dstDirectory isn't directory")
		}
	} else if os.IsNotExist(err) {
		os.MkdirAll(pathToDirectory, os.ModePerm)
	} else {
		return fmt.Errorf("stat dstDirectory error: %v", pathToDirectory)
	}
	return nil
}

func FillEnvironment(dst *map[string]string, paramToEnvVarNameMap map[string]interface{}, paramValues map[string]string) error {
	for desiredParamName, envNameInterface := range paramToEnvVarNameMap {
		envName, ok := envNameInterface.(string)
		if !ok {
			return fmt.Errorf("env name for param %q must be string", desiredParamName)
		}
		paramValueAssigned := false
		for definedParamName, paramValue := range paramValues {
			if definedParamName == desiredParamName {
				(*dst)[envName] = paramValue
				paramValueAssigned = true
				break
			}
		}
		if !paramValueAssigned {
			return fmt.Errorf("unsupported env param name %q", desiredParamName)
		}
	}
	return nil
}

type YamlMap map[interface{}]interface{}

func yamlValueToJsonValue(v interface{}) interface{} {
	yamlMap, ok := IsYamlMap(v)
	if ok {
		return yamlMap.ToMap()
	}
	slice, ok := v.([]interface{})
	if ok {
		r := make([]interface{}, len(slice))
		for i, item := range slice {
			r[i] = yamlValueToJsonValue(item)
		}
		return r
	}
	return v
}

func (m YamlMap) ToMap() map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range m {
		res[fmt.Sprintf("%v", k)] = yamlValueToJsonValue(v)
	}
	return res
}

func IsYamlMap(v interface{}) (YamlMap, bool) {
	m, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	return YamlMap(m), true
}

func IsHttpPortAvailable(host string, port int) bool {
	_, err := http.Get(fmt.Sprintf("http://%s:%d", host, port))
	if err != nil {
		// fmt.Printf("is port available connection err: %v\n", err)
		return false
	}
	return true
}
