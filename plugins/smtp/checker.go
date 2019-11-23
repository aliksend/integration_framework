package smtp

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
)

type ICheck interface {
	Check(httpServiceUrl string, variables map[string]interface{}) error
}

func (s *Service) Checker(param interface{}) (plugins.IServiceChecker, error) {
	checkConfigYaml, ok := helper.IsYamlMap(param)
	if !ok {
		return nil, fmt.Errorf("service check for smtp must be map, but it is %T (%#v)", param, param)
	}
	var checkers []ICheck
	for checkName, check := range checkConfigYaml.ToMap() {
		switch checkName {
		case "mails":
			checkParams, ok := check.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("mails check should be map, but it is %T (%#v)", check, check)
			}
			for mailboxName, mailboxContent := range checkParams {
				mailboxContentArr, ok := mailboxContent.([]interface{})
				if !ok {
					return nil, fmt.Errorf("mailbox content check should be array, but it is %T (%#v)", mailboxContent, mailboxContent)
				}
				var mails []MailCheck
				for _, mailboxContentItem := range mailboxContentArr {
					mailboxContentMap, ok := mailboxContentItem.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("mailbox content item check should be map, but it is %T (%#v)", mailboxContent, mailboxContent)
					}
					var (
						fromPtr    *string
						subjectPtr *string
						contentPtr *string
					)
					from, ok := mailboxContentMap["from"].(string)
					if ok {
						fromPtr = &from
					}
					subject, ok := mailboxContentMap["subject"].(string)
					if ok {
						subjectPtr = &subject
					}
					content, ok := mailboxContentMap["content"].(string)
					if ok {
						contentPtr = &content
					}
					if subjectPtr == nil && contentPtr == nil {
						return nil, fmt.Errorf("mailbox content check checks nothing")
					}
					mails = append(mails, MailCheck{
						from:    fromPtr,
						subject: subjectPtr,
						content: contentPtr,
					})
				}
				checkers = append(checkers, NewMailsChecker(mailboxName, mails))
			}
		default:
			return nil, fmt.Errorf("checker %q not defined for smtp", checkName)
		}
	}

	return &Checker{
		service:  s,
		checkers: checkers,
	}, nil
}

type Checker struct {
	service  *Service
	checkers []ICheck
}

func (pcc Checker) CheckService(saveResult plugins.FnResultSaver, variables map[string]interface{}) error {
	for i, check := range pcc.checkers {
		err := check.Check(fmt.Sprintf("http://localhost:%d/", pcc.service.port), variables)
		if err != nil {
			return fmt.Errorf("unable to check smtp %d: %v", i, err)
		}
	}
	return nil
}
