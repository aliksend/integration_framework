package postgres

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

func NewClearPrepare() *ClearPrepare {
	return &ClearPrepare{}
}

type ClearPrepare struct {
}

func (pp ClearPrepare) Prepare(conn *sqlx.DB) error {
	fmt.Println(".. postgres preparer clear")
	var tableNames []string
	err := conn.Select(&tableNames, "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema' AND tablename != 'schema_migrations'")
	if err != nil {
		return fmt.Errorf("unable to get all table names from db: %v", err)
	}
	_, err = conn.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY", strings.Join(tableNames, ",")))
	if err != nil {
		return fmt.Errorf("unable to truncate tables: %v", err)
	}
	return nil
}
