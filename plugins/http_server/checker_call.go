package http_server

import (
	"encoding/json"
	"fmt"
	"integration_framework/helper"
	"integration_framework/testing"
	"io/ioutil"
	"net/http"
)

func NewCallsCheck(calls []CheckCall) *CallsCheck {
	return &CallsCheck{
		calls: calls,
	}
}

type CallsCheck struct {
	calls []CheckCall
}

type ActualCall struct {
	Route string `json:"route"`
	Body  string `json:"body"`
}

func (c CallsCheck) Check(serviceUrl string, variables map[string]interface{}) error {
	resp, err := http.Get(serviceUrl + "__calls")
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %v", err)
	}
	var actualCalls []ActualCall
	err = json.Unmarshal(respBody, &actualCalls)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response body: %v", err)
	}
	if len(actualCalls) != len(c.calls) {
		return fmt.Errorf("different calls count: actual is %d, expected is %d", len(actualCalls), len(c.calls))
	}
	for i, check := range c.calls {
		actualCall := actualCalls[i]
		var parsedActualBody interface{}
		err := check.unmarshal([]byte(actualCall.Body), &parsedActualBody)
		if err != nil {
			return fmt.Errorf("check #%d failed: unable to parse actual call body: %v", i, err)
		}
		if actualCall.Route != check.Route {
			return fmt.Errorf("check #%d failed: route not matched. expcted %q, actual %q", i, check.Route, actualCall.Route)
		}
		checkBody, err := helper.ApplyInterpolationForObject(check.Body, variables)
		if err != nil {
			return fmt.Errorf("unable to apply interpolation: %v", err)
		}
		err = testing.IsEqual(parsedActualBody, checkBody)
		if err != nil {
			return fmt.Errorf("check #%d failed: body not matched: %v", i, err)
		}
	}
	return nil
}
