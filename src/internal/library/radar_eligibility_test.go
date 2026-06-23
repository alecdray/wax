package library

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/alecdray/wax/src/internal/core/db"
)

// Radar eligibility (ADR 0005): an album is radar-eligible unless the user
// currently owns or wishlists it. A `removed` release no longer blocks the
// radar, so a discarded album can be put back on it.

func newRadarTestService(t *testing.T) (*Service, *sql.DB) {
	t.Helper()
	repo, sqlDB := newRepoTestDB(t)
	return &Service{db: db.WrapSqlDB(sqlDB), repo: repo}, sqlDB
}

func seedUserRelease(t *testing.T, sqlDB *sql.DB, userID, albumID, releaseID, status string) {
	t.Helper()
	if _, err := sqlDB.Exec(
		`INSERT INTO releases (id, album_id, format) VALUES (?, ?, 'digital')`,
		releaseID, albumID,
	); err != nil {
		t.Fatalf("seed release: %v", err)
	}
	if _, err := sqlDB.Exec(
		`INSERT INTO user_releases (id, user_id, release_id, status, created_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		"ur-"+releaseID, userID, releaseID, status,
	); err != nil {
		t.Fatalf("seed user_release: %v", err)
	}
}

func TestAddAlbumToRadar_AllowsAlbumWithOnlyRemovedRelease(t *testing.T) {
	svc, sqlDB := newRadarTestService(t)
	ctx := context.Background()
	seedLibraryUser(t, sqlDB, "u1")
	seedAlbum(t, sqlDB, "a1", "Discarded")
	seedUserRelease(t, sqlDB, "u1", "a1", "rel1", "removed")

	if err := svc.AddAlbumToRadar(ctx, "u1", "a1"); err != nil {
		t.Fatalf("removed-release album should be radar-eligible: %v", err)
	}

	albums, err := svc.GetRadarAlbums(ctx, "u1")
	if err != nil {
		t.Fatalf("GetRadarAlbums: %v", err)
	}
	if len(albums) != 1 || albums[0].ID != "a1" {
		t.Fatalf("expected removed-release album a1 on radar, got %+v", albums)
	}
}

func TestAddAlbumToRadar_RefusesOwnedAndWishlistedAlbums(t *testing.T) {
	for _, status := range []string{"owned", "wishlist"} {
		t.Run(status, func(t *testing.T) {
			svc, sqlDB := newRadarTestService(t)
			ctx := context.Background()
			seedLibraryUser(t, sqlDB, "u1")
			seedAlbum(t, sqlDB, "a1", "Decided")
			seedUserRelease(t, sqlDB, "u1", "a1", "rel1", status)

			err := svc.AddAlbumToRadar(ctx, "u1", "a1")
			if !errors.Is(err, ErrAlbumAlreadyDecided) {
				t.Fatalf("expected ErrAlbumAlreadyDecided for %s album, got %v", status, err)
			}
		})
	}
}

func TestGetRadarAlbums_ExcludesOwnedAndWishlistedButKeepsRemoved(t *testing.T) {
	svc, sqlDB := newRadarTestService(t)
	ctx := context.Background()
	seedLibraryUser(t, sqlDB, "u1")

	// A radar row exists for all three, but owned/wishlisted must be filtered
	// out at query time while the removed one survives.
	for _, tc := range []struct {
		album, release, status string
	}{
		{"owned-alb", "rel-o", "owned"},
		{"wish-alb", "rel-w", "wishlist"},
		{"removed-alb", "rel-r", "removed"},
	} {
		seedAlbum(t, sqlDB, tc.album, tc.album)
		seedUserRelease(t, sqlDB, "u1", tc.album, tc.release, tc.status)
		if _, err := sqlDB.Exec(
			`INSERT INTO user_album_radar (id, user_id, album_id) VALUES (?, ?, ?)`,
			"rad-"+tc.album, "u1", tc.album,
		); err != nil {
			t.Fatalf("seed radar row: %v", err)
		}
	}

	albums, err := svc.GetRadarAlbums(ctx, "u1")
	if err != nil {
		t.Fatalf("GetRadarAlbums: %v", err)
	}
	if len(albums) != 1 || albums[0].ID != "removed-alb" {
		t.Fatalf("expected only removed-alb on radar, got %+v", albums)
	}
}
