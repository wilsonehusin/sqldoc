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

const pgListColumns = `SELECT
	c.column_name
	, c.data_type
	, c.is_nullable
	, c.column_default

FROM information_schema.columns c
WHERE c.table_schema = $1 AND c.table_name = $2
ORDER BY ordinal_position`

/*
const pgListColumns = `
SELECT
	c.column_name
	, c.data_type
	, c.is_nullable
	, c.column_default
	, tc.constraint_type,
	, *
FROM information_schema.columns c
LEFT OUTER JOIN information_schema.key_column_usage k
	ON c.column_name = k.column_name
	AND c.table_name = k.table_name
LEFT OUTER JOIN information_schema.table_constraints tc
	ON tc.constraint_name = k.constraint_name
WHERE c.table_name = $1
	AND c.table_schema = $2
`
*/
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

const pgListConstraints = `
SELECT
	  tc.constraint_schema
	, tc.constraint_name
	, tc.constraint_type
--	, kcu1.constraint_schema
--	, kcu1.constraint_name
--	, kcu1.table_name from_table
	, kcu1.column_name from_column
--	, kcu1.ordinal_position from_position
--	, kcu2.constraint_schema
--	, kcu2.constraint_name
	, kcu2.table_schema to_schema
	, kcu2.table_name to_table
	, kcu2.column_name to_column
--  , kcu2.ordinal_position to_position
FROM information_schema.table_constraints tc
LEFT OUTER JOIN information_schema.referential_constraints rc
  	USING(constraint_schema, constraint_name)
JOIN information_schema.key_column_usage as kcu1
    ON  kcu1.constraint_catalog = tc.constraint_catalog
    AND kcu1.constraint_schema  = tc.constraint_schema
    AND kcu1.constraint_name    = tc.constraint_name
LEFT OUTER JOIN information_schema.key_column_usage as kcu2
    ON kcu2.constraint_catalog = rc.unique_constraint_catalog
    AND kcu2.constraint_schema = rc.unique_constraint_schema
    AND kcu2.constraint_name   = rc.unique_constraint_name
    AND kcu2.ordinal_position  = kcu1.ordinal_position
WHERE   tc.table_name = $2
	AND tc.table_schema = $1
`

func (p *Postgres) ListConstraints(ctx context.Context, schema, table string) ([]TableConstraint, error) {
	rows, err := p.db.QueryContext(ctx, pgListConstraints, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var constraints []TableConstraint
	for rows.Next() {
		var c TableConstraint
		var tblColumn sql.NullString
		var tgtSchema sql.NullString
		var tgtTable sql.NullString
		var tgtColumn sql.NullString
		if err := rows.Scan(&c.Schema, &c.Name, &c.Type, &tblColumn, &tgtSchema, &tgtTable, &tgtColumn); err != nil {
			return nil, err
		}
		if tblColumn.Valid {
			c.TableColumn = tblColumn.String
		}
		if tgtSchema.Valid {
			c.TargetSchema = tgtSchema.String
		}
		if tgtTable.Valid {
			c.TargetTable = tgtTable.String
		}
		if tgtColumn.Valid {
			c.TargetColumn = tgtColumn.String
		}

		constraints = append(constraints, c)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return constraints, nil
}
