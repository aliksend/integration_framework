package smtp

import (
	"fmt"
	"net/http"
)

func NewClearAllPreparer() *ClearAllPreparer {
	return &ClearAllPreparer{}
}

type ClearAllPreparer struct {
}

func (pp ClearAllPreparer) Prepare(httpServiceUrl string) error {
	fmt.Println(".. smtp preparer clear")
	resp, err := http.Post(httpServiceUrl+"__reset_mails", "application/json", nil)
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unsuccessfull status code: %d", resp.StatusCode)
}
