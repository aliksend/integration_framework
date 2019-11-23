package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

func NewClearPrepare() *ClearPrepare {
	return &ClearPrepare{}
}

type ClearPrepare struct {
}

func (pp ClearPrepare) Prepare(conn *sqlx.DB) error {
	fmt.Println(".. mysql preparer clear")
	var tableNames []string
	err := conn.Select(&tableNames, "SHOW TABLES")
	if err != nil {
		return fmt.Errorf("unable to get all table names from db: %v", err)
	}
	for _, tableName := range tableNames {
		_, err = conn.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName))
		if err != nil {
			return fmt.Errorf("unable to truncate table %q: %v", tableName, err)
		}
	}
	return nil
}
