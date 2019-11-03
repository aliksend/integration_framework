package postgres

import (
	"fmt"
	"github.com/jmoiron/sqlx"
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
	_, err := conn.Exec(pp.exec)
	if err != nil {
		return fmt.Errorf("unable to run %q on postgres: %v", pp.exec, err)
	}
	return nil
}
