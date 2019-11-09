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

		err := tester.Exec()
		if err != nil {
			fmt.Printf("====> test failed: %v\n", err)
			failedTests[tester.Name] = err
		} else {
			fmt.Printf("====> test passed\n")
		}
	}

	if len(failedTests) == 0 {
		fmt.Printf("==== %d test(s) passed\n", len(pc.Testers))
	} else {
		fmt.Printf("%d test(s) of %d fails\n", len(failedTests), len(pc.Testers))
		for failedTestName, failedTestError := range failedTests {
			fmt.Printf("==== #%q\n", failedTestName)
			fmt.Printf("%v\n", failedTestError)
		}
		exitCode = 2
	}

	return
}
