package http_server

import (
	"fmt"
	"net/http"
)

func NewResetCallsPrepare() *ResetCallsPrepare {
	return &ResetCallsPrepare{}
}

type ResetCallsPrepare struct {
}

func (p ResetCallsPrepare) Prepare(serviceUrl string) error {
	resp, err := http.Post(serviceUrl+"__reset_calls", "application/json", nil)
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unsuccessfull status code: %d", resp.StatusCode)
}
