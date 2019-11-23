package postgres

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_framework/helper"
	"integration_framework/plugins"
	"integration_framework/testing"
)

type QueryChecker struct {
	query          string
	expectedResult interface{}
	saveResultTo   string
}

func (pc QueryChecker) makeRequest(conn *sqlx.DB, query string) ([]interface{}, error) {
	actualResult := make([]interface{}, 0)
	rows, err := conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("unable to make sql request: %v", err)
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("unable to get result columns: %v", err)
	}
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("unable to scan result: %v", err)
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		actualResult = append(actualResult, m)
	}
	return actualResult, nil
}

func NewQueryChecker(query string, expectedResult interface{}, saveResultTo string) *QueryChecker {
	return &QueryChecker{
		query:          query,
		expectedResult: expectedResult,
		saveResultTo:   saveResultTo,
	}
}

func (pc QueryChecker) Check(conn *sqlx.DB, saveResult plugins.FnResultSaver, variables map[string]interface{}) error {
	fmt.Println(".. postgres checker query", pc.query)

	query, err := helper.ApplyInterpolation(pc.query, variables)
	if err != nil {
		return fmt.Errorf("unable to interpolate query: %v", err)
	}
	actualResult, err := pc.makeRequest(conn, query)
	if err != nil {
		return fmt.Errorf("unable to make requset to db: %v", err)
	}
	// fmt.Printf("\n!! query: %s\n", pc.query)
	fmt.Printf("..   actual  : %#v\n", actualResult)

	if pc.expectedResult != nil {
		fmt.Printf("..   expected: %#v\n", pc.expectedResult)
		// separate variable to be able to convert to slice
		expectedResult := pc.expectedResult
		_, expectingSlice := expectedResult.([]interface{})
		if !expectingSlice {
			expectedResult = []interface{}{expectedResult}
		}
		expectedResult, err = helper.ApplyInterpolationForObject(expectedResult, variables)
		if err != nil {
			return fmt.Errorf("unable to apply interpolation: %v", err)
		}
		expectedResult, err = testing.ApplyConverters(expectedResult)
		if err != nil {
			return fmt.Errorf("unable to apply converters: %v", err)
		}
		err := testing.IsEqual(actualResult, expectedResult)
		// fmt.Println("!! is equal", err)
		// fmt.Printf("\n")
		if err != nil {
			return fmt.Errorf("db state not valid: %v", err)
		}
	}
	if pc.saveResultTo != "" {
		saveResult(pc.saveResultTo, actualResult)
	}

	return nil
}
