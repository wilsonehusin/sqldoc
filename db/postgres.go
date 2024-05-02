package db

import (
	"context"
	"database/sql"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &Postgres{db}, nil
}

func (p *Postgres) Close() error {
	if p.db == nil {
		return nil
	}
	if err := p.db.Close(); err != nil {
		return err
	}
	p.db = nil
	return nil
}

func (p *Postgres) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

const pgListTables = `SELECT DISTINCT table_name FROM information_schema.tables WHERE table_schema = $1 ORDER BY table_name ASC`

func (p *Postgres) ListTables(ctx context.Context, schema string) ([]string, error) {
	rows, err := p.db.QueryContext(ctx, pgListTables, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

const pgListColumns = `SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2`

func (p *Postgres) ListColumns(ctx context.Context, schema, table string) ([]TableColumn, error) {
	rows, err := p.db.QueryContext(ctx, pgListColumns, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []TableColumn
	for rows.Next() {
		var c TableColumn
		var nullable string
		var defaultVal sql.NullString
		if err := rows.Scan(&c.Name, &c.Type, &nullable, &defaultVal); err != nil {
			return nil, err
		}
		c.Nullable = nullable == "YES"
		if defaultVal.Valid {
			c.Default = strings.TrimSpace(defaultVal.String)
		}
		columns = append(columns, c)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}
