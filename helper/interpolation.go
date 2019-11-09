package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"text/template"
)

func ApplyInterpolationForObject(input interface{}, variables map[string]interface{}) (interface{}, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal input: %v", err)
	}

	encodedAndInterpolated, err := ApplyInterpolation(string(encoded), variables)
	if err != nil {
		return nil, err
	}

	var output interface{}
	err = json.Unmarshal([]byte(encodedAndInterpolated), &output)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal result: %v", err)
	}
	return output, nil
}

func ApplyInterpolation(str string, variables map[string]interface{}) (string, error) {
	funcMap := template.FuncMap{
		"hash_password": func(s string) (string, error) {
			// TODO move to yaml configuration
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
			if err != nil {
				return "", err
			}
			return string(hashedPassword), nil
		},
		"quote": func(s string) string {
			return fmt.Sprintf(`"%s"`, strings.ReplaceAll(s, `"`, `\"`))
		},
		"singlequote": func(s string) string {
			return fmt.Sprintf(`'%s'`, strings.ReplaceAll(s, `'`, `\'`))
		},
		"string": func(val interface{}) string {
			switch val.(type) {
			case []byte, string:
				return fmt.Sprintf("%s", val)
			default:
				return fmt.Sprintf("%v", val)
			}
		},
	}

	// fmt.Println(">> parse encoded interpolated", string(str))
	tmpl, err := template.New("input").Funcs(funcMap).Parse(string(str))
	if err != nil {
		return "", fmt.Errorf("unable to parse template: %v", err)
	}

	var filledTemplateBuffer bytes.Buffer
	err = tmpl.Execute(&filledTemplateBuffer, variables)
	if err != nil {
		return "", fmt.Errorf("unable to execute template: %v", err)
	}
	// fmt.Println(">> interpolated template", filledTemplateBuffer.String())

	return filledTemplateBuffer.String(), nil
}
