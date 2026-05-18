package library

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

// GetProvisionalAlbums must succeed regardless of whether each provisional
// album also has an album_rating_log row. Until the underlying query was
// fixed, a LEFT JOIN onto the log produced a NULL rating column that errored
// on scan into a non-nullable Go float64 — surfacing as HTTP 500 on the
// provisional carousel.

func TestRepoGetProvisionalAlbums_NoLogRowYieldsNilRating(t *testing.T) {
	repo, sqlDB := newRepoTestDB(t)
	ctx := context.Background()

	seedLibraryUser(t, sqlDB, "u1")
	seedAlbum(t, sqlDB, "a1", "Backfilled Album")
	seedProvisionalState(t, sqlDB, "u1", "a1")
	// Deliberately no album_rating_log row — this is the shape that the
	// stalled→provisional backfill produces and that used to 500.

	rows, err := repo.GetProvisionalAlbums(ctx, "u1")
	if err != nil {
		t.Fatalf("GetProvisionalAlbums: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 provisional album, got %d", len(rows))
	}
	if rows[0].ID != "a1" {
		t.Fatalf("expected album a1 in result, got %q", rows[0].ID)
	}
	if rows[0].Rating != nil {
		t.Fatalf("expected Rating to be nil for log-less provisional album, got %v", *rows[0].Rating)
	}
}

func TestRepoGetProvisionalAlbums_WithLogRowYieldsLatestRating(t *testing.T) {
	repo, sqlDB := newRepoTestDB(t)
	ctx := context.Background()

	seedLibraryUser(t, sqlDB, "u1")
	seedAlbum(t, sqlDB, "a1", "Rated Album")
	seedProvisionalState(t, sqlDB, "u1", "a1")
	seedRatingLog(t, sqlDB, "log-old", "u1", "a1", 5.0, "2026-04-01 12:00:00")
	seedRatingLog(t, sqlDB, "log-new", "u1", "a1", 7.5, "2026-05-01 12:00:00")

	rows, err := repo.GetProvisionalAlbums(ctx, "u1")
	if err != nil {
		t.Fatalf("GetProvisionalAlbums: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 provisional album, got %d", len(rows))
	}
	if rows[0].Rating == nil {
		t.Fatalf("expected Rating to be set for provisional album with log row, got nil")
	}
	if *rows[0].Rating != 7.5 {
		t.Fatalf("expected most-recent rating 7.5, got %v", *rows[0].Rating)
	}
}

func TestRepoGetProvisionalAlbums_MixedLibrary(t *testing.T) {
	repo, sqlDB := newRepoTestDB(t)
	ctx := context.Background()

	seedLibraryUser(t, sqlDB, "u1")

	// Backfilled-from-stalled shape: state row but no log row.
	seedAlbum(t, sqlDB, "a-empty", "Backfilled")
	seedProvisionalState(t, sqlDB, "u1", "a-empty")

	// Normal shape: state row with a log row.
	seedAlbum(t, sqlDB, "a-rated", "Rated")
	seedProvisionalState(t, sqlDB, "u1", "a-rated")
	seedRatingLog(t, sqlDB, "log-1", "u1", "a-rated", 8.25, "2026-05-10 09:00:00")

	rows, err := repo.GetProvisionalAlbums(ctx, "u1")
	if err != nil {
		t.Fatalf("GetProvisionalAlbums: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 provisional albums, got %d", len(rows))
	}

	byID := make(map[string]ProvisionalAlbumDTO, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}

	empty, ok := byID["a-empty"]
	if !ok {
		t.Fatalf("expected backfilled album in result, missing: %+v", rows)
	}
	if empty.Rating != nil {
		t.Fatalf("expected nil Rating for backfilled album, got %v", *empty.Rating)
	}

	rated, ok := byID["a-rated"]
	if !ok {
		t.Fatalf("expected rated album in result, missing: %+v", rows)
	}
	if rated.Rating == nil || *rated.Rating != 8.25 {
		t.Fatalf("expected Rating=8.25 for rated album, got %v", rated.Rating)
	}
}

// --- helpers ---

// newRepoTestDB opens a fresh sqlite DB, applies every migration, and returns
// a library Repo plus the raw *sql.DB for fixture seeding.
func newRepoTestDB(t *testing.T) (*Repo, *sql.DB) {
	t.Helper()

	migrationsDir, err := filepath.Abs("../../../db/migrations")
	if err != nil {
		t.Fatalf("resolve migrations dir: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	return NewRepo(sqlc.New(sqlDB)), sqlDB
}

func seedLibraryUser(t *testing.T, sqlDB *sql.DB, userID string) {
	t.Helper()
	if _, err := sqlDB.Exec(
		`INSERT INTO users (id, spotify_id) VALUES (?, ?)`,
		userID, "spotify-"+userID,
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedAlbum(t *testing.T, sqlDB *sql.DB, albumID, title string) {
	t.Helper()
	if _, err := sqlDB.Exec(
		`INSERT INTO albums (id, spotify_id, title) VALUES (?, ?, ?)`,
		albumID, "spotify-"+albumID, title,
	); err != nil {
		t.Fatalf("seed album: %v", err)
	}
}

func seedProvisionalState(t *testing.T, sqlDB *sql.DB, userID, albumID string) {
	t.Helper()
	if _, err := sqlDB.Exec(
		`INSERT INTO album_rating_state (id, user_id, album_id, state)
		 VALUES (?, ?, ?, 'provisional')`,
		"ars-"+userID+"-"+albumID, userID, albumID,
	); err != nil {
		t.Fatalf("seed provisional state: %v", err)
	}
}

func seedRatingLog(t *testing.T, sqlDB *sql.DB, id, userID, albumID string, rating float64, createdAt string) {
	t.Helper()
	if _, err := sqlDB.Exec(
		`INSERT INTO album_rating_log (id, user_id, album_id, rating, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		id, userID, albumID, rating, createdAt,
	); err != nil {
		t.Fatalf("seed rating log: %v", err)
	}
}
