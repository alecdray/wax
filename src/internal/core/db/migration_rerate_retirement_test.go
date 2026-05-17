package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

// The retirement migration: backfills any historical 'stalled' rows in
// album_rating_state to 'provisional', drops the time-based columns, and
// must be safe to re-run. It also must NOT touch album_rating_log rows
// (history is immutable).
const retirementMigrationVersion int64 = 20260517000001
const prevMigrationVersion int64 = 20260510023140

// migrationsDirFromTest resolves the repo-root migrations directory from the
// test's working directory (src/internal/core/db).
func migrationsDirFromTest(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../../../../db/migrations")
	if err != nil {
		t.Fatalf("resolve migrations dir: %v", err)
	}
	return abs
}

// openMigratedToPrevious opens a fresh SQLite DB and applies every migration
// up to (and including) the one immediately before the rerate-retirement
// migration. The returned DB is positioned for fixture inserts that exercise
// the retirement migration on realistic-shape data.
func openMigratedToPrevious(t *testing.T) (*sql.DB, string) {
	t.Helper()
	migrationsDir := migrationsDirFromTest(t)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	if err := goose.UpTo(sqlDB, migrationsDir, prevMigrationVersion); err != nil {
		t.Fatalf("goose up to %d: %v", prevMigrationVersion, err)
	}
	return sqlDB, migrationsDir
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func seedUserAndAlbum(t *testing.T, db *sql.DB, userID, spotifyUser, albumID, spotifyAlbum string) {
	t.Helper()
	mustExec(t, db, `INSERT INTO users (id, spotify_id) VALUES (?, ?)`, userID, spotifyUser)
	mustExec(t, db, `INSERT INTO albums (id, spotify_id, title) VALUES (?, ?, ?)`, albumID, spotifyAlbum, "Album "+albumID)
}

// Criterion 1 + 7: stalled rows are migrated to provisional; log rows with
// state='stalled' are untouched.
func TestRetirementMigration_StalledRowsMoveToProvisional_LogUntouched(t *testing.T) {
	db, dir := openMigratedToPrevious(t)

	seedUserAndAlbum(t, db, "u1", "su1", "a1", "sa1")
	seedUserAndAlbum(t, db, "u2", "su2", "a2", "sa2")

	mustExec(t, db,
		`INSERT INTO album_rating_state (id, user_id, album_id, state, snooze_count, next_rerate_at) VALUES (?, ?, ?, 'stalled', 3, NULL)`,
		"rs1", "u1", "a1",
	)
	mustExec(t, db,
		`INSERT INTO album_rating_state (id, user_id, album_id, state, snooze_count, next_rerate_at) VALUES (?, ?, ?, 'provisional', 0, datetime('now', '+30 days'))`,
		"rs2", "u2", "a2",
	)

	// A historical log row recorded under the stalled lifecycle. Captured
	// values to assert untouched after migration.
	const (
		logID        = "log-stalled-1"
		logScore     = 6.4
		logCreatedAt = "2025-12-01 12:34:56"
		logState     = "stalled"
	)
	mustExec(t, db,
		`INSERT INTO album_rating_log (id, user_id, album_id, rating, state, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		logID, "u1", "a1", logScore, logState, logCreatedAt,
	)

	// Apply the retirement migration.
	if err := goose.UpTo(db, dir, retirementMigrationVersion); err != nil {
		t.Fatalf("goose up to %d: %v", retirementMigrationVersion, err)
	}

	// album_rating_state: zero rows in 'stalled', the previously-stalled row is now 'provisional'.
	var stalledCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM album_rating_state WHERE state = 'stalled'`).Scan(&stalledCount); err != nil {
		t.Fatalf("count stalled state rows: %v", err)
	}
	if stalledCount != 0 {
		t.Errorf("expected 0 stalled rows in album_rating_state, got %d", stalledCount)
	}

	var rs1State string
	if err := db.QueryRow(`SELECT state FROM album_rating_state WHERE id = 'rs1'`).Scan(&rs1State); err != nil {
		t.Fatalf("read rs1: %v", err)
	}
	if rs1State != "provisional" {
		t.Errorf("previously-stalled row should now be 'provisional', got %q", rs1State)
	}

	var rs2State string
	if err := db.QueryRow(`SELECT state FROM album_rating_state WHERE id = 'rs2'`).Scan(&rs2State); err != nil {
		t.Fatalf("read rs2: %v", err)
	}
	if rs2State != "provisional" {
		t.Errorf("untouched row should still be 'provisional', got %q", rs2State)
	}

	// album_rating_log row with state='stalled' must be intact: same id, score,
	// state, and created_at instant.
	var (
		gotID, gotState string
		gotScore        float64
		gotCreatedAt    time.Time
	)
	if err := db.QueryRow(`SELECT id, rating, state, created_at FROM album_rating_log WHERE id = ?`, logID).
		Scan(&gotID, &gotScore, &gotState, &gotCreatedAt); err != nil {
		t.Fatalf("read historical log row: %v", err)
	}
	wantCreatedAt, err := time.Parse("2006-01-02 15:04:05", logCreatedAt)
	if err != nil {
		t.Fatalf("parse seeded created_at: %v", err)
	}
	if gotID != logID || gotScore != logScore || gotState != logState || !gotCreatedAt.Equal(wantCreatedAt) {
		t.Errorf("historical log row mutated: got (id=%q score=%v state=%q createdAt=%s), want (%q %v %q %s)",
			gotID, gotScore, gotState, gotCreatedAt, logID, logScore, logState, wantCreatedAt)
	}
}

// Criterion 2: re-running the migration is a no-op and does not write any
// rows to album_rating_log.
func TestRetirementMigration_IsIdempotent_AndWritesNoLogRows(t *testing.T) {
	db, dir := openMigratedToPrevious(t)

	seedUserAndAlbum(t, db, "u1", "su1", "a1", "sa1")
	mustExec(t, db,
		`INSERT INTO album_rating_state (id, user_id, album_id, state, snooze_count, next_rerate_at) VALUES (?, ?, ?, 'provisional', 0, datetime('now', '+30 days'))`,
		"rs1", "u1", "a1",
	)
	mustExec(t, db,
		`INSERT INTO album_rating_log (id, user_id, album_id, rating, state, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"log1", "u1", "a1", 7.0, "provisional", "2026-01-15 10:00:00",
	)

	// Apply the retirement migration.
	if err := goose.UpTo(db, dir, retirementMigrationVersion); err != nil {
		t.Fatalf("first goose up: %v", err)
	}

	var beforeStateCount, beforeLogCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM album_rating_state`).Scan(&beforeStateCount); err != nil {
		t.Fatalf("count state rows: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM album_rating_log`).Scan(&beforeLogCount); err != nil {
		t.Fatalf("count log rows: %v", err)
	}

	// Re-run. goose tracks applied versions, so the migration's statements
	// must not execute again — counts must not change and no log rows are written.
	if err := goose.UpTo(db, dir, retirementMigrationVersion); err != nil {
		t.Fatalf("second goose up: %v", err)
	}

	var afterStateCount, afterLogCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM album_rating_state`).Scan(&afterStateCount); err != nil {
		t.Fatalf("count state rows after rerun: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM album_rating_log`).Scan(&afterLogCount); err != nil {
		t.Fatalf("count log rows after rerun: %v", err)
	}

	if afterStateCount != beforeStateCount {
		t.Errorf("album_rating_state row count changed on re-run: %d → %d", beforeStateCount, afterStateCount)
	}
	if afterLogCount != beforeLogCount {
		t.Errorf("album_rating_log row count changed on re-run (migration wrote log rows!): %d → %d", beforeLogCount, afterLogCount)
	}
	if afterLogCount != 1 {
		t.Errorf("expected exactly the seeded log row to remain (1), got %d", afterLogCount)
	}
}

// Criterion 3: the time-based columns are no longer present on album_rating_state.
func TestRetirementMigration_DropsTimeBasedColumns(t *testing.T) {
	db, dir := openMigratedToPrevious(t)
	if err := goose.UpTo(db, dir, retirementMigrationVersion); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	rows, err := db.Query(`PRAGMA table_info(album_rating_state)`)
	if err != nil {
		t.Fatalf("pragma table_info: %v", err)
	}
	defer rows.Close()

	cols := map[string]bool{}
	for rows.Next() {
		var (
			cid      int
			name     string
			typeName string
			notnull  int
			dflt     sql.NullString
			pk       int
		)
		if err := rows.Scan(&cid, &name, &typeName, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan pragma row: %v", err)
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("pragma rows err: %v", err)
	}

	if cols["snooze_count"] {
		t.Error("album_rating_state still has column snooze_count")
	}
	if cols["next_rerate_at"] {
		t.Error("album_rating_state still has column next_rerate_at")
	}

	// Sanity check: live columns survive.
	for _, want := range []string{"id", "user_id", "album_id", "state", "created_at", "updated_at"} {
		if !cols[want] {
			t.Errorf("album_rating_state is missing required column %q", want)
		}
	}
}
