package testing

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

var converterRegexp = regexp.MustCompile(`^\$\$_([\w.]+)$`)

type CustomEqualityChecker interface {
	IsEqualTo(v interface{}) (bool, error)
}

// TODO хорошо бы apply-ить конвертеры при настройке, а не при работе, чтобы ошибки выдавались перед запуском
//      но это не полуяится так как нужно учитывать интерполяцию. самая лучшая мысль - создавать объект "equality checker",
//      который будет создаваться с expected-значением на этапе настройки и сможет учитывать конвертеры и интерполяцию и будет вызываться на этапе проверки с actual-ными значениями
func ApplyConverters(input interface{}) (interface{}, error) {
	return applyConverters(input, "")
}

func applyConverters(inputInterface interface{}, objectPath string) (interface{}, error) {
	// fmt.Printf("()()()()()() input %#v\n", inputInterface)
	switch input := inputInterface.(type) {
	case map[string]interface{}:
		for k, v := range input {
			if converterRegexp.MatchString(k) {
				converterName := converterRegexp.FindStringSubmatch(k)[1]
				switch converterName {
				case "exist", "exists":
					exists, ok := v.(bool)
					if !ok {
						return nil, fmt.Errorf("%s value should be bool, but it is %T (%#v)", k, v, v)
					}
					return ExitstEqualityChecker{
						exists: exists,
					}, nil
				case "convert":
					format, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("format to convert to should be string, but it is %T (%#v)", v, v)
					}
					delete(input, "$$_convert")
					switch format {
					case "json":
						return JsonEqualityChecker{
							expectedJson: input,
						}, nil
					default:
						return nil, fmt.Errorf("unknown format to convert to: %s", format)
					}
				case "regexp":
					re, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("%s value should be bool, but it is %T (%#v)", k, v, v)
					}
					compiledRe, err := regexp.Compile(re)
					if err != nil {
						return nil, fmt.Errorf("unable to compile re %q: %v", re, err)
					}
					return RegexpEqualityChecker{
						re: compiledRe,
					}, nil
				default:
					return nil, fmt.Errorf("converter with name %q not found", converterName)
				}
			}
			switch v.(type) {
			case map[string]interface{}:
				res, err := applyConverters(v, fmt.Sprintf("%s.%s", objectPath, k))
				if err != nil {
					return nil, err // fmt.Errorf("unable to apply converters to %s.%s: %v", objectPath, k, err)
				}
				input[k] = res
			case []interface{}:
				res, err := applyConverters(v, fmt.Sprintf("%s.%s", objectPath, k))
				if err != nil {
					return nil, err // fmt.Errorf("unable to apply converters to %s.%s: %v", objectPath, k, err)
				}
				input[k] = res
			}
		}
	case []interface{}:
		for i, val := range input {
			res, err := applyConverters(val, fmt.Sprintf("%s[%d]", objectPath, i))
			if err != nil {
				return nil, err // fmt.Errorf("unable to apply converters to %s.%d: %v", objectPath, i, err)
			}
			input[i] = res
		}
	}
	return inputInterface, nil
}

type ExitstEqualityChecker struct {
	exists bool
}

func (e ExitstEqualityChecker) IsEqualTo(v interface{}) (bool, error) {
	actuallyExists := e.isExists(v)
	if e.exists && !actuallyExists {
		return false, fmt.Errorf("expected to exists, but actually not")
	}
	if !e.exists && actuallyExists {
		return false, fmt.Errorf("expected to not exists, but actually exists")
	}
	return true, nil
}

func (e ExitstEqualityChecker) isExists(v interface{}) bool {
	if v == nil {
		return false
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		isNil := reflect.ValueOf(v).IsNil()
		return !isNil
	}

	return true
}

type JsonEqualityChecker struct {
	expectedJson map[string]interface{}
}

func (e JsonEqualityChecker) IsEqualTo(v interface{}) (bool, error) {
	// fmt.Printf("[][][] is %#v equal to %#v\n", v, e.expectedJson)
	var actualMap map[string]interface{}
	switch value := v.(type) {
	case map[string]interface{}:
		actualMap = value
	case []byte:
		err := json.Unmarshal(value, &actualMap)
		if err != nil {
			return false, fmt.Errorf("unable to parse json value: %v", err)
		}
	case string:
		err := json.Unmarshal([]byte(value), &actualMap)
		if err != nil {
			return false, fmt.Errorf("unable to parse json value: %v", err)
		}
	default:
		return false, fmt.Errorf("unsupported type to compare to json map %T (%#v)", v, v)
	}
	// fmt.Printf("[][][] actually compare %#v\n", actualMap)
	err := IsEqual(actualMap, e.expectedJson)
	if err != nil {
		return false, err
	}
	return true, nil
}

type RegexpEqualityChecker struct {
	re *regexp.Regexp
}

func (e RegexpEqualityChecker) IsEqualTo(v interface{}) (bool, error) {
	switch value := v.(type) {
	case string:
		if e.re.MatchString(value) {
			return true, nil
		}
		return false, fmt.Errorf("value %v not match regexp %v", value, e.re.String())
	case []byte:
		if e.re.Match(value) {
			return true, nil
		}
		return false, fmt.Errorf("value %v not match regexp %v", value, e.re.String())
	default:
		return false, fmt.Errorf("value of type %T not supported to be checked by regexp", v)
	}
}
