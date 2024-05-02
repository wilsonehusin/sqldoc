package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Database interface {
	Close() error

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	ListTables(ctx context.Context, schema string) ([]string, error)
	ListColumns(ctx context.Context, schema, table string) ([]TableColumn, error)
	ListConstraints(ctx context.Context, schema, table string) ([]TableConstraint, error)
}

type TableColumn struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
	PK       bool
	FK       string
	FK_ref   string
	UK       bool
}

/*
	tc.constraint_schema
	, tc.constraint_name
	, tc.constraint_type
	, kcu1.constraint_schema
	, kcu1.constraint_name
	, kcu1.table_name from_table
	, kcu1.column_name from_column
	, kcu1.ordinal_position from_position
	, kcu2.constraint_schema
	, kcu2.constraint_name
	, kcu2.table_name to_table
	, kcu2.column_name to_column
	, kcu2.ordinal_position to_position
*/
type TableConstraint struct {
	Schema       string
	Name         string
	Type         string
	TableColumn  string // referring column
	TargetSchema string // referred schema
	TargetTable  string // referred table
	TargetColumn string // referred column
}

func New(url string) (Database, error) {
	protocol := strings.Split(url, "://")[0]
	switch protocol {
	case "postgres", "postgresql", "pgx":
		return NewPostgres(url)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
