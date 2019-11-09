package application_config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var confiFileNameRegexp = regexp.MustCompile(`\.ya?ml$`)

func absPath(path string, rootToResolveRelativePaths string) (res string, err error) {
	if filepath.IsAbs(path) {
		res = path
	} else {
		res, err = filepath.Abs(filepath.Join(rootToResolveRelativePaths, path))
		if err != nil {
			return
		}
	}
	return
}

func LoadConfig(pathToConfig string, rootToResolveRelativePaths string) (*Config, error) {
	absolutePathToConfig, err := absPath(pathToConfig, rootToResolveRelativePaths)
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path to config: %v", err)
	}
	configs, err := loadAllFilesInPath(absolutePathToConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to load files in path %s: %v", absolutePathToConfig, err)
	}
	config, err := joinConfigs(configs)
	if err != nil {
		return nil, fmt.Errorf("unable to join configs: %v", err)
	}
	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("config invalid: %v", err)
	}

	config.Application.Path, err = absPath(config.Application.Path, filepath.Dir(absolutePathToConfig))
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path to application: %v", err)
	}

	return config, nil
}

func IterateOverConfigFiles(path string, callback func(filename string) error) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("unable to stat file %s: %v", path, err)
	}
	if fileInfo.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return fmt.Errorf("unable to readdir %s: %v", path, err)
		}

		for _, file := range files {
			if file.IsDir() {
				shouldReadDir := shouldReadDirWithName(file.Name())
				if !shouldReadDir {
					continue
				}
			} else {
				shouldLoadFile := shouldLoadFileWithName(file.Name())
				if !shouldLoadFile {
					continue
				}
			}
			err := IterateOverConfigFiles(filepath.Join(path, file.Name()), callback)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err = callback(path)
	if err != nil {
		return fmt.Errorf("unable to process callback for path %s: %v", path, err)
	}
	return nil
}

func loadAllFilesInPath(path string) ([]*Config, error) {
	var configs []*Config
	err := IterateOverConfigFiles(path, func(filename string) error {
		config, err := loadConfigFile(filename)
		if err != nil {
			return fmt.Errorf("unable to load config file %s: %v", filename, err)
		}
		configs = append(configs, config)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func loadConfigFile(path string) (*Config, error) {
	config := Config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %v", err)
	}
	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	return &config, nil
}

func joinConfigs(configs []*Config) (*Config, error) {
	res := &Config{
		GeneralCases: make(map[string]GeneralCase),
		Environment:  make(map[string]interface{}),
		Services:     make(map[string]ServiceConfig),
	}
	for _, config := range configs {
		err := res.Join(config)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func shouldLoadFileWithName(filename string) bool {
	if !confiFileNameRegexp.MatchString(filename) {
		return false
	}
	if filename[0] == '_' {
		return false
	}
	return true
}

func shouldReadDirWithName(dirname string) bool {
	if dirname[0] == '_' {
		return false
	}
	return true
}
