package helper

import (
	"regexp"
)

var converterRegexp = regexp.MustCompile(`$$_([\w.]+)`)

func ApplyConverters(input interface{}) (interface{}, error) {
	// TODO тут только один вариант - перебирать вложенные map-ы и массивы, искать все поля $$_ и как-то модицифировать конфиг
	return input, nil
}
