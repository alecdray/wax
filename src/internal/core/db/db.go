package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	_ "github.com/alecdray/wax/src/internal/migrations"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

const migrationsDir = "db/migrations"

type DB struct {
	sql     *sql.DB
	queries *sqlc.Queries
}

func NewDB(filepath string) (*DB, error) {
	sqlDb, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err := sqlDb.Ping(); err != nil {
		return nil, err
	}

	sqlDb.SetMaxOpenConns(25)
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetConnMaxLifetime(5 * time.Minute)

	queries := sqlc.New(sqlDb)
	db := &DB{sql: sqlDb, queries: queries}

	if err := db.runMigrations(); err != nil {
		return nil, err
	}

	return db, nil
}

func newDBWithTx(db DB, tx *sql.Tx) *DB {
	queries := sqlc.New(tx)
	db.queries = queries
	dbTx := &db
	return dbTx
}

func (db *DB) Sql() *sql.DB {
	return db.sql
}

func (db *DB) Queries() *sqlc.Queries {
	return db.queries
}

func (db *DB) Close() error {
	return db.sql.Close()
}

func (db *DB) WithTx(fn func(*DB) error) error {
	tx, err := db.sql.Begin()
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	dbTx := newDBWithTx(*db, tx)
	err = fn(dbTx)

	return err
}

func (db *DB) runMigrations() error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	pending, err := goose.CollectMigrations(migrationsDir, 0, goose.MaxVersion)
	if err != nil && !errors.Is(err, goose.ErrNoMigrationFiles) {
		return err
	}

	current, err := goose.GetDBVersion(db.sql)
	if err != nil {
		return err
	}

	for _, m := range pending {
		if m.Version > current {
			return fmt.Errorf("database is not up to date (at %d, %d migration(s) pending) — run: task db/up", current, len(pending)-indexOf(pending, current))
		}
	}

	return nil
}

func indexOf(migrations goose.Migrations, version int64) int {
	for i, m := range migrations {
		if m.Version > version {
			return i
		}
	}
	return len(migrations)
}
