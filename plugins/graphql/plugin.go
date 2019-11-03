package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"integration_framework/application_config"
	"integration_framework/helper"
	"integration_framework/plugins"
	"io/ioutil"
	"net/http"
)

func init() {
	plugins.DefineRequester("graphql", func(request interface{}, defaults application_config.RequestDefaults) (plugins.IRequester, error) {
		requestMap, ok := request.(helper.YamlMap)
		if !ok {
			return nil, fmt.Errorf("request should be map")
		}

		requester := GraphqlRequester{
			method:   "",
			url:      "",
			headers:  nil,
			defaults: defaults,
		}

		query, ok := requestMap["query"]
		if ok {
			queryString, ok := query.(string)
			if !ok {
				return nil, fmt.Errorf("request query should be string")
			}
			requester.query = queryString
		}

		headers, ok := requestMap["headers"]
		if ok {
			headersYamlMap, ok := headers.(helper.YamlMap)
			if !ok {
				return nil, fmt.Errorf("request headers should be map")
			}
			headersMap := helper.YamlMapToJsonMap(headersYamlMap)
			fmt.Printf(">> new graphql requester with headers map %#v\n", headersMap)
			requester.headers = make(map[string]string)
			for headerName, headerValue := range headersMap {
				requester.headers[headerName] = fmt.Sprintf("%v", headerValue)
			}
		}

		method, ok := requestMap["method"]
		if ok {
			methodString, ok := method.(string)
			if !ok {
				return nil, fmt.Errorf("method should be string")
			}
			requester.method = methodString
		}
		return &requester, nil
	})
}

type GraphqlRequester struct {
	query   string
	headers map[string]string
	method  string
	url     string

	defaults application_config.RequestDefaults
}

func (r *GraphqlRequester) applyDefaults() {
	fmt.Printf("r apply defaults %#v\n", r)
	if r.method == "" {
		r.method = r.defaults.Method
	}
	if r.url == "" {
		r.url = r.defaults.Url
	}
	if r.headers == nil {
		r.headers = r.defaults.Headers
	}
	fmt.Printf("r applied defaults %#v\n", r)
}

func (r *GraphqlRequester) MakeRequest() (responseBody []byte, statusCode int, err error) {
	r.applyDefaults()
	fmt.Printf(">>>>>>>>>>>> MAKE REQUEST %q %q %q %#v\n", r.query, r.url, r.method, r.headers)
	payload, err := json.Marshal(map[string]interface{}{
		"query": r.query,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("unable to marshal payload: %v", err)
	}
	request, err := http.NewRequest(r.method, r.url, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("unable to create http request: %v", err)
	}

	if r.headers != nil {
		for headerName, headerValue := range r.headers {
			request.Header.Add(headerName, headerValue)
		}
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to make request to application: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to read response body: %v", err)
	}
	return respBody, resp.StatusCode, nil
}

func (r GraphqlRequester) Join(joinWithRequester plugins.IRequester) (plugins.IRequester, error) {
	requester, ok := joinWithRequester.(*GraphqlRequester)
	if !ok {
		return nil, fmt.Errorf("cannot join graphql requester with provided %#v", joinWithRequester)
	}
	newRequester := GraphqlRequester{
		query:    r.query,
		headers:  r.headers,
		method:   r.method,
		url:      r.url,
		defaults: r.defaults,
	}
	if requester.query != "" {
		newRequester.query = requester.query
	}
	if requester.headers != nil {
		if newRequester.headers == nil {
			newRequester.headers = make(map[string]string)
		}
		for headerName, headerValue := range requester.headers {
			newRequester.headers[headerName] = headerValue
		}
	}
	if requester.method != "" {
		newRequester.method = requester.method
	}
	if requester.url != "" {
		newRequester.url = requester.url
	}
	return &newRequester, nil
}
