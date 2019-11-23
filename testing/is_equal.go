package testing

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"integration_framework/helper"
	"reflect"
	"strconv"
	"time"
)

func IsEqual(actualResult interface{}, expectedResult interface{}) error {
	err := anyTypeMatcherFunc(actualResult, expectedResult, "")
	if err != nil {
		return err
	}
	// TODO один проход функции - только матчинг actual-а на expected. Но хочется проверять полную вложенность.
	//      как вариант - вызывать второй раз функцию "наоброт", но тогда нужно поправить возвращаемые ошибки
	// err := anyTypeMatcherFunc(expectedResult, actualResult, "")
	// if err != nil {
	// 	return err
	// }
	return nil
}

func actualExpectedError(actual interface{}, expected interface{}, keyMsg string) error {
	return fmt.Errorf("%s\n Actual value  : %#v (%T)\n Expected value: %#v (%T)", keyMsg, actual, actual, expected, expected)
}

func anyTypeMatcherFunc(actualInterface interface{}, expectedInterface interface{}, actualKey string) (err error) {
	// fmt.Printf("\nmatch any type %q\nactual   %#v\nexpected %#v\n", actualKey, actualInterface, expectedInterface)

	keyMsg := ""
	if actualKey != "" {
		keyMsg = fmt.Sprintf("invalid value of key %q.", actualKey)
	}
	switch expected := expectedInterface.(type) {
	case []interface{}:
		actualArray, ok := actualInterface.([]interface{})
		if !ok {
			return actualExpectedError(actualInterface, expected, keyMsg)
		}
		if len(expected) != len(actualArray) {
			return actualExpectedError(len(actualArray), len(expected), fmt.Sprintf("invalid length of key %q.", actualKey))
		}
		for i, actualArrayValue := range actualArray {
			err = anyTypeMatcherFunc(actualArrayValue, expected[i], fmt.Sprintf("%s.[%d]", actualKey, i))
			if err != nil {
				return
			}
		}
	case []map[string]interface{}:
		actualArray, ok := actualInterface.([]interface{})
		if !ok {
			return actualExpectedError(actualInterface, expected, keyMsg)
		}
		if len(expected) != len(actualArray) {
			return actualExpectedError(len(actualArray), len(expected), fmt.Sprintf("invalid length of key %q.", actualKey))
		}
		for i, actualArrayValue := range actualArray {
			err = anyTypeMatcherFunc(actualArrayValue, expected[i], fmt.Sprintf("%s.[%d]", actualKey, i))
			if err != nil {
				return
			}
		}
	case map[string]interface{}:
		switch actual := actualInterface.(type) {
		case map[string]interface{}:
			for key, expectedValue := range expected {
				actualValue, ok := actual[key]
				if !ok {
					return fmt.Errorf("unexpected value of key %q: %#v", actualKey+"."+key, actualValue)
				}
				err = anyTypeMatcherFunc(actualValue, expectedValue, actualKey+"."+key)
				if err != nil {
					return
				}
			}
		case []byte:
			parsedActual := make(map[string]interface{})
			err := json.Unmarshal(actual, &parsedActual)
			if err != nil {
				return fmt.Errorf("unable to parse []byte actual value %q: %v", string(actual), err)
			}
			return anyTypeMatcherFunc(parsedActual, expected, actualKey)
		default:
			return actualExpectedError(actualInterface, expected, keyMsg)
		}
	case nil:
		if actualInterface == nil {
			return nil
		}
		switch reflect.TypeOf(actualInterface).Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if !reflect.ValueOf(actualInterface).IsNil() {
				return fmt.Errorf("value isn't nil")
			}
		}
	case int:
		var actualInt int
		switch actual := actualInterface.(type) {
		case int:
			actualInt = actual
		case int64:
			actualInt = int(actual)
		case float64:
			actualInt = int(actual)
		}
		if expected != actualInt {
			return actualExpectedError(actualInterface, expected, keyMsg)
		}
	case float64:
		var actualFloat float64
		switch actual := actualInterface.(type) {
		case int:
			actualFloat = float64(actual)
		case int64:
			actualFloat = float64(actual)
		case float64:
			actualFloat = actual
		}
		if expected != actualFloat {
			return actualExpectedError(actualInterface, expected, keyMsg)
		}
	case []byte:
		expectedStr := string(expected)
		var actualStr string
		switch actualInterface.(type) {
		case string, []byte:
			actualStr = fmt.Sprintf("%s", actualInterface)
		default:
			actualStr = fmt.Sprintf("%v", actualInterface)
		}
		if expectedStr != actualStr {
			return actualExpectedError(actualStr, expectedStr, keyMsg)
		}
	case string:
		switch actual := actualInterface.(type) {
		case time.Time:
			return isEqualTime(actual, expected, keyMsg)
		case *time.Time:
			return isEqualTime(*actual, expected, keyMsg)
		case []byte:
			if expected != string(actual) {
				return actualExpectedError(string(actual), expected, keyMsg)
			}
		case float64:
			expectedFloat, err := strconv.ParseFloat(expected, 64)
			if err != nil {
				return fmt.Errorf("unable to parse float: %#v", expected)
			}
			if expectedFloat != actual {
				return actualExpectedError(actual, expectedFloat, keyMsg)
			}
		default:
			if expected != actualInterface {
				return actualExpectedError(actualInterface, expected, keyMsg)
			}
		}
	default:
		if expected != actualInterface {
			return actualExpectedError(actualInterface, expected, keyMsg)
		}
	}
	return nil
}

func isEqualTime(actual time.Time, expected string, keyMsg string) error {
	t, err := time.Parse(helper.TimeLayout, expected)
	if err != nil {
		return fmt.Errorf("unable to parse time: %v", err)
	}
	formattedActual := actual.Format(helper.TimeLayout)
	if !t.Equal(actual) {
		return fmt.Errorf("%s\n Actual value  : %#v (%T)\n Expected value: %#v (%T)", keyMsg, formattedActual, formattedActual, expected, expected)
	}
	return nil
}

func IsEqualYamlMaps(actual []byte, expected []byte) (bool, error) {
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
	err = IsEqual(helper.YamlMapToJsonMap(actualValue), helper.YamlMapToJsonMap(expectedValue))
	if err != nil {
		fmt.Println(">> not equal!", err)
	}
	return err != nil, nil
}
