package md

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/glamour"
	"github.com/nao1215/markdown"
	"go.husin.dev/sqldoc/config"
	"go.husin.dev/sqldoc/db"
)

type MD struct {
	db   db.Database
	conf *config.Config
}

const (
	openFileFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	openFilePerm  = 0644
)

func New(db db.Database, conf *config.Config) *MD {
	return &MD{db: db, conf: conf}
}

func (m *MD) Generate(ctx context.Context) error {
	if err := os.MkdirAll(m.conf.Documentation.Directory, 0755); err != nil {
		return fmt.Errorf("precreating directory: %w", err)
	}
	if m.conf.Documentation.Strategy == "per_table" {
		return m.execPerTable(ctx)
	}
	return m.execUnified(ctx)
}

func (m *MD) execUnified(ctx context.Context) error {
	fpath := filepath.Join(m.conf.Documentation.Directory, m.conf.Documentation.Filename)
	f, err := os.OpenFile(fpath, openFileFlags, openFilePerm)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	for _, schema := range m.conf.Database.Schemas {
		tables, err := m.db.ListTables(ctx, schema)
		if err != nil {
			return fmt.Errorf("fetching tables: %w", err)
		}

		for i, table := range tables {
			if slices.Contains(m.conf.Database.ExcludeTables, table) {
				continue
			}
			if err := m.renderTable(ctx, f, schema, table); err != nil {
				return fmt.Errorf("writing markdown: %w", err)
			}
			if i < len(tables)-1 {
				f.WriteString("\n")
			}
		}
	}
	return nil
}

func (m *MD) execPerTable(ctx context.Context) error {
	for _, schema := range m.conf.Database.Schemas {
		tables, err := m.db.ListTables(ctx, schema)
		if err != nil {
			return fmt.Errorf("fetching tables: %w", err)
		}

		for _, table := range tables {
			if slices.Contains(m.conf.Database.ExcludeTables, table) {
				continue
			}
			fpath := filepath.Join(m.conf.Documentation.Directory, table+".md")
			f, err := os.OpenFile(fpath, openFileFlags, openFilePerm)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}
			defer f.Close()

			if err := m.renderTable(ctx, f, schema, table); err != nil {
				return fmt.Errorf("writing markdown: %w", err)
			}
			f.Close()
		}
	}
	return nil
}

func (m *MD) renderTable(ctx context.Context, w io.Writer, schema, tableName string) error {
	canvas := markdown.NewMarkdown(w)

	canvas.H1(markdown.Code(tableName))
	canvas.PlainText("")
	mdTableRows := [][]string{}
	columns, err := m.db.ListColumns(ctx, schema, tableName)
	if err != nil {
		return fmt.Errorf("fetching columns for '%s.%s': %w", schema, tableName, err)
	}
	for _, column := range columns {
		nullable := ""
		if !column.Nullable {
			nullable = "NOT NULL"
		}
		d := column.Default
		if d != "" {
			d = markdown.Code(d)
		}
		mdTableRows = append(mdTableRows, []string{
			markdown.Code(column.Name),
			column.Type,
			nullable,
			d,
		})
	}
	canvas.Table(markdown.TableSet{
		Header: []string{"Name", "Type", "Nullable", "Default"},
		Rows:   mdTableRows,
	})
	if m.conf.Documentation.Stdout {
		out, err := glamour.Render(canvas.String(), "dark")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering markdown to terminal: %v\n", err)
		} else {
			fmt.Println(out)
		}
	}

	return canvas.Build()
}
