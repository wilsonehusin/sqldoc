package md

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
				_, _ = f.WriteString("\n")
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
	constraints, err := m.db.ListConstraints(ctx, schema, tableName)

	if err != nil {
		return fmt.Errorf("fetching constraints for '%s.%s': %w", schema, tableName, err)
	}

	columns, err := m.db.ListColumns(ctx, schema, tableName)
	if err != nil {
		return fmt.Errorf("fetching columns for '%s.%s': %w", schema, tableName, err)
	}
	for idx := range columns {
		nullable := ""
		if !columns[idx].Nullable {
			nullable = "NOT NULL"
		}
		d := columns[idx].Default
		if d != "" {
			d = markdown.Code(d)
		}
		c := []string{}
		t := ""
		for _, constraint := range constraints {
			if constraint.TableColumn == columns[idx].Name {
				if constraint.Type == "PRIMARY KEY" {
					c = append(c, "PK")
					columns[idx].PK = true
				}
				if constraint.Type == "UNIQUE" {
					c = append(c, "UK")
					columns[idx].UK = true
				}
				if constraint.Type == "FOREIGN KEY" {
					c = append(c, "FK")
					t = fmt.Sprintf("%s.%s", constraint.TargetSchema, constraint.TargetTable)
					columns[idx].FK = t
					t = fmt.Sprintf("%s.%s", t, constraint.TargetColumn)
					columns[idx].FK_ref = constraint.TargetColumn

				}
			}

		}
		mdTableRows = append(mdTableRows, []string{
			markdown.Code(columns[idx].Name),
			columns[idx].Type,
			nullable,
			d,
			strings.Join(c[:], ", "),
			t,
		})
	}
	canvas.Table(markdown.TableSet{
		Header: []string{"Name", "Type", "Nullable", "Default", "Key", "Target"},
		Rows:   mdTableRows,
	})

	if m.conf.Documentation.Mermaid {
		mermaid_dep := ""

		mermaid := "```mermaid\nerDiagram\n"
		mermaid = fmt.Sprintf("%s\"%s.%s\"{\n", mermaid, schema, tableName)
		for _, column := range columns {
			// nullable := ""
			// if !column.Nullable {
			// 	nullable = "NN"
			// }
			k := []string{}
			if column.PK {
				k = append(k, "PK")
			}
			if column.FK != "" {
				k = append(k, "FK")
				mermaid_dep = fmt.Sprintf("%s\n\"%s.%s\" o|--o| \"%s\":%s",
					mermaid_dep,
					schema, tableName,
					column.FK,
					column.FK_ref)
			}

			if column.UK {
				k = append(k, "UK")
			}

			mermaid = fmt.Sprintf("%s%s %s %s\n", mermaid,
				strings.Replace(column.Type, " ", "_", -1),
				column.Name, strings.Join(k[:], ","))
		}

		mermaid = fmt.Sprintf("%s}\n%s\n```", mermaid, mermaid_dep)

		canvas.PlainText(mermaid)
	}

	if len(constraints) > 0 {
		canvas.PlainText("")
		canvas.H2("Constraints")
		canvas.PlainText("")
		mdConstraintsRows := [][]string{}
		for _, constraint := range constraints {

			col := ""
			if constraint.TableColumn != "" {
				col = markdown.Code(constraint.TableColumn)
			}
			tgtsch := ""
			if constraint.TargetSchema != "" {
				tgtsch = markdown.Code(constraint.TargetSchema)
			}
			tgttbl := ""
			if constraint.TargetTable != "" {
				tgttbl = markdown.Code(constraint.TargetTable)
			}
			tgtcol := ""
			if constraint.TargetColumn != "" {
				tgtcol = markdown.Code(constraint.TargetColumn)
			}
			mdConstraintsRows = append(mdConstraintsRows, []string{
				markdown.Code(constraint.Schema),
				markdown.Code(constraint.Name),
				markdown.Code(constraint.Type),
				col,
				tgtsch,
				tgttbl,
				tgtcol,
			})
		}
		canvas.Table(markdown.TableSet{
			Header: []string{"Schema", "Name", "Type", "Referring Column", "Target Schema", "Target Table", "Target Column"},
			Rows:   mdConstraintsRows,
		})
	}

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
