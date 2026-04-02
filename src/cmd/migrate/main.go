// migrate runs goose database migrations, including Go migrations registered
// in src/internal/migrations.
//
// Usage:
//
//	go run ./src/cmd/migrate/ [goose args...]
//	go run ./src/cmd/migrate/ up
//	go run ./src/cmd/migrate/ down
//	go run ./src/cmd/migrate/ status
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/alecdray/wax/src/internal/migrations"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

const migrationsDir = "db/migrations"

func main() {
	dbPath := flag.String("db", os.Getenv("GOOSE_DBSTRING"), "Path to SQLite database")
	flag.Parse()

	if *dbPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: migrate -db <path> [command]")
		fmt.Fprintln(os.Stderr, "Or set GOOSE_DBSTRING environment variable")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"up"}
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("Failed to set dialect", "error", err)
		os.Exit(1)
	}

	if err := goose.RunContext(context.Background(), args[0], db, migrationsDir, args[1:]...); err != nil {
		slog.Error("Migration failed", "error", err)
		os.Exit(1)
	}
}
