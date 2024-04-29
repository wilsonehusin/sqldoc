package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"slices"

	"github.com/charmbracelet/glamour"
	"github.com/nao1215/markdown"

	"go.husin.dev/sqldoc/config"
	"go.husin.dev/sqldoc/db"
)

const helpText = `sqldoc generates markdown documentation for SQL databases.

Commands:
  init      Write an example config file to sqldoc.yaml
  version   Print version information
  help      Print this help text

Usage:
  sqldoc [flags]

Flags:
`

func help() {
	fmt.Fprint(os.Stderr, helpText)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = help
	flag.Parse()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := config.WriteExample("sqldoc.yaml"); err != nil {
				panic(err)
			}
			fmt.Printf("Example config file written to sqldoc.yaml\n")
			return
		case "help":
			help()
			return
		case "version":
			version()
			return
		}
	}
	if err := run(); err != nil {
		panic(err)
	}
}

func version() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("sqldoc dev")
		return
	}
	rev := "dev"
	for _, i := range bi.Settings {
		if i.Key == "vcs.revision" {
			rev = i.Value
			break
		}
	}
	fmt.Printf("sqldoc %s (%s)\n", bi.Main.Version, rev)
}

var (
	conf *config.Config

	configPath = flag.String("config", "", "config file path")
)

func run() error {
	var err error
	conf, err = config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := db.New(conf.Database.URL)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	if err := exec(ctx, db); err != nil {
		return err
	}
	return nil
}

func exec(ctx context.Context, db db.Database) error {
	if err := os.MkdirAll(conf.Documentation.Directory, 0755); err != nil {
		return fmt.Errorf("precreating directory: %w", err)
	}
	if conf.Documentation.Strategy == "per_table" {
		return execPerTable(ctx, db)
	}
	return execUnified(ctx, db)
}

func execUnified(ctx context.Context, db db.Database) error {
	f, err := os.OpenFile(filepath.Join(conf.Documentation.Directory, conf.Documentation.Filename), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	for _, schema := range conf.Database.Schemas {
		tables, err := db.ListTables(ctx, schema)
		if err != nil {
			return fmt.Errorf("fetching tables: %w", err)
		}

		for i, table := range tables {
			if slices.Contains(conf.Database.ExcludeTables, table) {
				continue
			}
			if err := tableToMarkdown(ctx, db, f, schema, table); err != nil {
				return fmt.Errorf("writing markdown: %w", err)
			}
			if i < len(tables)-1 {
				f.WriteString("\n")
			}
		}
	}
	return nil
}

func execPerTable(ctx context.Context, db db.Database) error {
	for _, schema := range conf.Database.Schemas {
		tables, err := db.ListTables(ctx, schema)
		if err != nil {
			return fmt.Errorf("fetching tables: %w", err)
		}

		for _, table := range tables {
			if slices.Contains(conf.Database.ExcludeTables, table) {
				continue
			}
			f, err := os.OpenFile(filepath.Join(conf.Documentation.Directory, table+".md"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}
			defer f.Close()

			if err := tableToMarkdown(ctx, db, f, schema, table); err != nil {
				return fmt.Errorf("writing markdown: %w", err)
			}
			f.Close()
		}
	}
	return nil
}

func tableToMarkdown(ctx context.Context, db db.Database, w io.Writer, schema, tableName string) error {
	md := markdown.NewMarkdown(w)

	md.H1(markdown.Code(tableName))
	md.PlainText("")
	mdTableRows := [][]string{}
	columns, err := db.ListColumns(ctx, schema, tableName)
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
	md.Table(markdown.TableSet{
		Header: []string{"Name", "Type", "Nullable", "Default"},
		Rows:   mdTableRows,
	})
	if conf.Documentation.Stdout {
		out, err := glamour.Render(md.String(), "dark")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering markdown to terminal: %v\n", err)
		} else {
			fmt.Println(out)
		}
	}

	return md.Build()
}
