package smtp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type MailCheck struct {
	from    *string
	subject *string
	content *string
}

type ActualMail struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

type MailsChecker struct {
	mailbox string
	mails   []MailCheck
}

func NewMailsChecker(mailbox string, mails []MailCheck) *MailsChecker {
	return &MailsChecker{
		mailbox: mailbox,
		mails:   mails,
	}
}

func (mc MailsChecker) Check(httpServiceUrl string, variables map[string]interface{}) error {
	resp, err := http.Get(httpServiceUrl + "__mails")
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %v", err)
	}
	var actualMails []ActualMail
	err = json.Unmarshal(respBody, &actualMails)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response body: %v", err)
	}
	var actualMailsForThisMailbox []ActualMail
	for _, actualMail := range actualMails {
		if actualMail.To == mc.mailbox {
			actualMailsForThisMailbox = append(actualMailsForThisMailbox, actualMail)
		}
	}
	if len(actualMailsForThisMailbox) != len(mc.mails) {
		return fmt.Errorf("different mails count: actual is %d, expected is %d", len(actualMailsForThisMailbox), len(mc.mails))
	}
	for i, mailCheck := range mc.mails {
		actualMail := actualMailsForThisMailbox[i]
		if mailCheck.from != nil {
			if *mailCheck.from != actualMail.From {
				return fmt.Errorf("check #%d failed: from not matched. expected %q, actual %q", i, *mailCheck.from, actualMail.From)
			}
		}
		if mailCheck.subject != nil {
			if *mailCheck.subject != actualMail.Subject {
				return fmt.Errorf("check #%d failed: subject not matched. expected %q, actual %q", i, *mailCheck.subject, actualMail.Subject)
			}
		}
		if mailCheck.content != nil {
			if *mailCheck.content != actualMail.Content {
				return fmt.Errorf("check #%d failed: content not matched. expected %q, actual %q", i, *mailCheck.content, actualMail.Content)
			}
		}
	}
	return nil
}
