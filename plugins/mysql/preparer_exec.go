package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"integration_framework/helper"
)

func NewExecPrepare(exec string) *ExecPrepare {
	return &ExecPrepare{
		exec: exec,
	}
}

type ExecPrepare struct {
	exec string
}

func (pp ExecPrepare) Prepare(conn *sqlx.DB) error {
	query, err := helper.ApplyInterpolation(pp.exec, nil)
	if err != nil {
		return fmt.Errorf("unable to interpolate query: %v", err)
	}
	fmt.Println(".. mysql preparer exec", query)
	_, err = conn.Exec(query)
	if err != nil {
		return fmt.Errorf("unable to run %q on mysql: %v", pp.exec, err)
	}
	return nil
}
