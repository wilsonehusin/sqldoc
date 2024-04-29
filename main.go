package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"

	"go.husin.dev/sqldoc/config"
	"go.husin.dev/sqldoc/db"
	"go.husin.dev/sqldoc/md"
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
			ver()
			return
		}
	}
	if err := run(); err != nil {
		panic(err)
	}
}

var version = "dev"

func ver() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("sqldoc: unknown version")
		return
	}

	if v := bi.Main.Version; v != "(devel)" {
		// Use version from `go install`.
		version = v
	}
	rev := "dev"
	for _, i := range bi.Settings {
		if i.Key == "vcs.revision" {
			rev = i.Value
			break
		}
	}
	fmt.Printf("sqldoc: %s (%s)\n", version, rev)
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

	m := md.New(db, conf)
	if err := m.Generate(ctx); err != nil {
		return fmt.Errorf("generating markdown: %w", err)
	}
	return nil
}
