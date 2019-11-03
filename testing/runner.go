package testing

import (
	"fmt"
)

func (pc *ParsedConfig) RunTests() (exitCode int) {
	failedTests := make(map[string]error)

	if len(pc.Testers) == 0 {
		fmt.Println("==== no tests to run")
		return 3
	}

	for _, tester := range pc.Testers {
		fmt.Printf("---- %s\n", tester.Name)
		if len(pc.onlyCases) != 0 {
			found := false
			for _, onlyCase := range pc.onlyCases {
				if onlyCase == tester.Name {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("skipped\n")
				continue
			}
		}

		err := tester.Exec()
		if err != nil {
			fmt.Printf("====> test failed: %v\n", err)
			failedTests[tester.Name] = err
		} else {
			fmt.Printf("====> test passed\n")
		}
	}

	if len(failedTests) == 0 {
		fmt.Println("==== all tests passed")
	} else {
		fmt.Printf("%d test(s) fails\n", len(failedTests))
		for failedTestName, failedTestError := range failedTests {
			fmt.Printf("==== #%q\n", failedTestName)
			fmt.Printf("%v\n", failedTestError)
		}
		exitCode = 2
	}

	return
}
