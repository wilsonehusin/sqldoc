package md_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.husin.dev/sqldoc/config"
	"go.husin.dev/sqldoc/db"
	"go.husin.dev/sqldoc/md"
	"gotest.tools/v3/golden"
)

func goldenFileName(str string) string {
	ext := filepath.Ext(str)
	return "testdata/" + str[:len(str)-len(ext)] + ".golden" + ext
}

func migrate(t *testing.T, ctx context.Context, conn db.Database) {
	t.Helper()
	schemaPath := filepath.Join("testdata", t.Name()+"_schema.sql")
	log.Printf("migrate: %s", schemaPath)

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("reading up migration: %s", err.Error())
	}
	if _, err := conn.ExecContext(ctx, string(schema)); err != nil {
		t.Fatalf("executing up migration: %s", err.Error())
	}
}

func TestRenderTable(t *testing.T) {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL environment variable is not set")
	}
	conn, err := db.New(dsn)
	if err != nil {
		t.Fatalf("connecting database: %s", err.Error())
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Logf("[WARNING] error closing database connection: %s", err.Error())
		}
	})

	runTest := func(t *testing.T) {
		migrate(t, ctx, conn)

		dir := t.TempDir()
		out := strings.ReplaceAll(t.Name(), "/", "__") + ".md"

		conf := config.Default()
		conf.Documentation.Stdout = false
		conf.Documentation.Directory = dir
		conf.Documentation.Filename = out

		m := md.New(conn, conf)
		if err := m.Generate(ctx); err != nil {
			t.Fatalf("generating documentation: %s", err.Error())
		}
		golden.Assert(t, filepath.Join(dir, out), goldenFileName(out))
	}

	t.Run("SingleTableWithColumns", runTest)
}
