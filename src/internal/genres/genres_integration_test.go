package genres

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/genregraph"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

// Real Wikidata Q-ids present in the embedded genre graph.
const (
	qHyperpop   = "Q104695865"
	qDeathMetal = "Q483251"
	qAmbient    = "Q193207"
	qBebop      = "Q105513"
	qSoul       = "Q131272"
	qNeoSoul    = "Q268253"
	qFunk       = "Q164444"
)

func newGenresService(t *testing.T) (*Service, *sql.DB) {
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

	graph, err := genregraph.Load()
	if err != nil {
		t.Fatalf("load genre graph: %v", err)
	}

	// discogs is nil — the read/persistence paths under test never touch it.
	return NewService(db.WrapSqlDB(sqlDB), nil, graph), sqlDB
}

func seedAlbum(t *testing.T, sqlDB *sql.DB, id string) {
	t.Helper()
	if _, err := sqlDB.Exec(
		`INSERT INTO albums (id, spotify_id, title) VALUES (?, ?, ?)`,
		id, "sp-"+id, "Album "+id,
	); err != nil {
		t.Fatalf("seed album: %v", err)
	}
}

func seedAlbumGenre(t *testing.T, sqlDB *sql.DB, albumID, genreID, label string) {
	t.Helper()
	if _, err := sqlDB.Exec(
		`INSERT INTO album_genres (id, album_id, genre_id, genre_label) VALUES (?, ?, ?, ?)`,
		albumID+"-"+genreID, albumID, genreID, label,
	); err != nil {
		t.Fatalf("seed album_genre: %v", err)
	}
}

func primaryLabels(ps []genregraph.Primary) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Label
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAlbumPrimaries(t *testing.T) {
	svc, sqlDB := newGenresService(t)
	ctx := context.Background()

	seedAlbum(t, sqlDB, "hp")
	seedAlbumGenre(t, sqlDB, "hp", qHyperpop, "hyperpop")

	seedAlbum(t, sqlDB, "dm")
	seedAlbumGenre(t, sqlDB, "dm", qDeathMetal, "death metal")

	seedAlbum(t, sqlDB, "amb")
	seedAlbumGenre(t, sqlDB, "amb", qAmbient, "ambient")

	// Album with leaves on two unrelated branches → union of primaries.
	seedAlbum(t, sqlDB, "mix")
	seedAlbumGenre(t, sqlDB, "mix", qBebop, "bebop")
	seedAlbumGenre(t, sqlDB, "mix", qDeathMetal, "death metal")

	got, err := svc.AlbumPrimaries(ctx, []string{"hp", "dm", "amb", "mix"})
	if err != nil {
		t.Fatalf("AlbumPrimaries: %v", err)
	}

	if labels := primaryLabels(got["hp"]); !equal(labels, []string{"pop", "electronic"}) {
		t.Errorf("hyperpop album primaries = %v, want [pop electronic]", labels)
	}
	if labels := primaryLabels(got["dm"]); !equal(labels, []string{"metal"}) {
		t.Errorf("death metal album primaries = %v, want [metal]", labels)
	}
	if _, ok := got["amb"]; ok {
		t.Errorf("ambient album should be absent (uncategorized), got %v", primaryLabels(got["amb"]))
	}
	// metal precedes jazz in curated order.
	if labels := primaryLabels(got["mix"]); !equal(labels, []string{"metal", "jazz"}) {
		t.Errorf("mixed album primaries = %v, want [metal jazz]", labels)
	}
}

func TestAlbumPrimaries_DominanceOrder(t *testing.T) {
	svc, sqlDB := newGenresService(t)
	ctx := context.Background()

	// Two leaves map to soul (soul itself + neo soul), one to funk → soul leads.
	seedAlbum(t, sqlDB, "dom")
	seedAlbumGenre(t, sqlDB, "dom", qSoul, "soul")
	seedAlbumGenre(t, sqlDB, "dom", qNeoSoul, "neo soul")
	seedAlbumGenre(t, sqlDB, "dom", qFunk, "funk")

	got, err := svc.AlbumPrimaries(ctx, []string{"dom"})
	if err != nil {
		t.Fatalf("AlbumPrimaries: %v", err)
	}
	if labels := primaryLabels(got["dom"]); !equal(labels, []string{"soul", "funk"}) {
		t.Errorf("dominance order = %v, want [soul funk] (soul has more leaf support)", labels)
	}
}

func TestAlbumPrimaries_CapsAtThree(t *testing.T) {
	svc, sqlDB := newGenresService(t)
	ctx := context.Background()

	// Four supported primaries: soul (x2 via neo soul), metal, jazz, funk.
	// Only the top 3 by dominance survive — funk (weakest, latest curated) drops.
	seedAlbum(t, sqlDB, "cap")
	seedAlbumGenre(t, sqlDB, "cap", qSoul, "soul")
	seedAlbumGenre(t, sqlDB, "cap", qNeoSoul, "neo soul")
	seedAlbumGenre(t, sqlDB, "cap", qDeathMetal, "death metal")
	seedAlbumGenre(t, sqlDB, "cap", qBebop, "bebop")
	seedAlbumGenre(t, sqlDB, "cap", qFunk, "funk")

	got, err := svc.AlbumPrimaries(ctx, []string{"cap"})
	if err != nil {
		t.Fatalf("AlbumPrimaries: %v", err)
	}
	if labels := primaryLabels(got["cap"]); !equal(labels, []string{"soul", "metal", "jazz"}) {
		t.Errorf("capped primaries = %v, want [soul metal jazz] (funk dropped)", labels)
	}
}

func TestEnrichmentMarkerAndReplace(t *testing.T) {
	svc, sqlDB := newGenresService(t)
	ctx := context.Background()
	seedAlbum(t, sqlDB, "a1")

	// Not enriched yet.
	enriched, err := svc.repo.EnrichedAlbumIDs(ctx, []string{"a1"})
	if err != nil {
		t.Fatalf("EnrichedAlbumIDs: %v", err)
	}
	if enriched["a1"] {
		t.Fatal("album should not be enriched before processing")
	}

	// Persist genres then mark enriched (what EnrichAlbum does in a tx).
	if err := svc.repo.ReplaceAlbumGenres(ctx, "a1", []AlbumGenreDTO{{GenreID: qBebop, Label: "bebop"}}); err != nil {
		t.Fatalf("ReplaceAlbumGenres: %v", err)
	}
	if err := svc.repo.MarkEnriched(ctx, "a1"); err != nil {
		t.Fatalf("MarkEnriched: %v", err)
	}

	enriched, err = svc.repo.EnrichedAlbumIDs(ctx, []string{"a1"})
	if err != nil {
		t.Fatalf("EnrichedAlbumIDs: %v", err)
	}
	if !enriched["a1"] {
		t.Fatal("album should be enriched after marking")
	}

	// Replace is a clear-and-set: a second call with a different genre wins.
	if err := svc.repo.ReplaceAlbumGenres(ctx, "a1", []AlbumGenreDTO{{GenreID: qDeathMetal, Label: "death metal"}}); err != nil {
		t.Fatalf("ReplaceAlbumGenres (replace): %v", err)
	}
	byAlbum, err := svc.repo.GetAlbumGenresByAlbumIDs(ctx, []string{"a1"})
	if err != nil {
		t.Fatalf("GetAlbumGenresByAlbumIDs: %v", err)
	}
	if len(byAlbum["a1"]) != 1 || byAlbum["a1"][0].GenreID != qDeathMetal {
		t.Fatalf("expected only death metal after replace, got %+v", byAlbum["a1"])
	}
}
