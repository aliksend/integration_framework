package plugins

import (
	"fmt"
	"integration_framework/application_config"
)

type IRequester interface {
	MakeRequest() (responseBody []byte, statusCode int, err error)

	// joins caller requester with provided requester returning new requester.
	// caller requester should remain unchanged
	Join(IRequester) (IRequester, error)
}

type RequesterConstructor func(params interface{}, defaults application_config.RequestDefaults) (IRequester, error)

var requesterConstructors map[string]RequesterConstructor

func init() {
	requesterConstructors = make(map[string]RequesterConstructor)
}

func DefineRequester(name string, constructor RequesterConstructor) {
	_, ok := requesterConstructors[name]
	if ok {
		panic(fmt.Errorf("requester with name %q already defined", name))
	}
	requesterConstructors[name] = constructor
}

func GetRequesterConstructor(name string) (RequesterConstructor, bool) {
	constructor, ok := requesterConstructors[name]
	return constructor, ok
}
