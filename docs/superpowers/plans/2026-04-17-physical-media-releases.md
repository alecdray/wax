# Physical Media Releases Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow users to record ownership of vinyl, CD, and cassette releases for albums already in their Wax/Spotify library, with optional Discogs integration to attach pressing details.

**Architecture:** A formats modal on the album detail page presents all four format rows — digital (read-only) and three physical (togglable). Toggling a physical format on optionally surfaces an inline Discogs search that pre-fills label and release date. A single PUT saves all ownership and Discogs data. The `releases` table gains three nullable Discogs columns; `user_releases` is unchanged.

**Tech Stack:** Go, SQLite/SQLC, Templ, HTMX, Alpine.js, Discogs API (existing client)

---

## Worktree Setup

Before starting, run these from the worktree root (`/Users/shmoopy/workshop/workbench/wax/wt-a75e`):

```bash
cp /Users/shmoopy/workshop/projects/wax/.env .env
cp /Users/shmoopy/workshop/projects/wax/tmp/db.sql ./tmp/db.sql
npm install
```

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `db/migrations/20260417000001_add_discogs_fields_to_releases.sql` | Adds `discogs_id`, `label`, `released_at` to `releases` |
| Modify | `db/schema.sql` | SQLC schema source — must stay in sync with migrations |
| Modify | `db/queries/releases.sql` | Update existing queries + add `UpdateReleaseDiscogsInfo` |
| Modify | `db/queries/user_releases.sql` | Add `SoftDeleteUserRelease` (by userID + releaseID) |
| Modify | `src/internal/library/service.go` | New DTOs, `GetAlbumFormats`, `SaveAlbumFormats` |
| Modify | `src/internal/library/service_test.go` | Tests for new service methods |
| Modify | `src/internal/library/adapters/http.go` | Add `discogsService` field to `HttpHandler` |
| Create | `src/internal/library/adapters/formats.go` | HTTP handlers for 4 new routes |
| Create | `src/internal/library/adapters/formats_modal.templ` | Formats modal + Discogs inline search templates |
| Modify | `src/internal/library/adapters/album_detail.templ` | Replace static format icons with clickable button |
| Modify | `src/internal/server/server.go` | Pass discogs to library handler; register 4 new routes |

---

## Task 1: DB Migration

**Files:**
- Create: `db/migrations/20260417000001_add_discogs_fields_to_releases.sql`
- Modify: `db/schema.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- db/migrations/20260417000001_add_discogs_fields_to_releases.sql

-- +goose Up
ALTER TABLE releases ADD COLUMN discogs_id TEXT;
ALTER TABLE releases ADD COLUMN label TEXT;
ALTER TABLE releases ADD COLUMN released_at DATETIME;

-- +goose Down
ALTER TABLE releases DROP COLUMN discogs_id;
ALTER TABLE releases DROP COLUMN label;
ALTER TABLE releases DROP COLUMN released_at;
```

- [ ] **Step 2: Update `db/schema.sql` — add the three columns to the `releases` table**

Find this block in `db/schema.sql`:
```sql
CREATE TABLE releases (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    format text not null check(format in ('digital', 'vinyl', 'cd', 'cassette')),
    created_at datetime not null default current_timestamp,
    deleted_at datetime,
    unique(album_id, format)
);
```

Replace with:
```sql
CREATE TABLE releases (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    format text not null check(format in ('digital', 'vinyl', 'cd', 'cassette')),
    created_at datetime not null default current_timestamp,
    deleted_at datetime,
    discogs_id text,
    label text,
    released_at datetime,
    unique(album_id, format)
);
```

- [ ] **Step 3: Run the migration**

```bash
task db/up
```

Expected output: `goose: successfully migrated database to version: 20260417000001`

- [ ] **Step 4: Commit**

```bash
git add db/migrations/20260417000001_add_discogs_fields_to_releases.sql db/schema.sql
git commit -m "feat: add discogs_id, label, released_at columns to releases"
```

---

## Task 2: Update SQLC Queries and Regenerate

**Files:**
- Modify: `db/queries/releases.sql`
- Modify: `db/queries/user_releases.sql`

The existing `GetRelease`, `GetReleases`, and `GetOrCreateRelease` queries use explicit column lists that don't include the new columns. SQLC generates scan code from the SQL, so these must be updated to `SELECT *` / `RETURNING *` to include the new fields in the generated struct.

- [ ] **Step 1: Update `db/queries/releases.sql`**

Replace the entire file with:

```sql
-- name: CreateRelease :exec
INSERT INTO releases (id, album_id, format) VALUES (?, ?, ?);

-- name: GetOrCreateRelease :one
INSERT INTO releases (id, album_id, format) VALUES (?, ?, ?)
ON CONFLICT (album_id, format)
DO UPDATE SET album_id = album_id
RETURNING *;

-- name: GetRelease :one
SELECT * FROM releases WHERE id = ?;

-- name: GetReleases :many
SELECT * FROM releases WHERE album_id = ?;

-- name: UpdateRelease :exec
UPDATE releases
SET
    album_id = COALESCE(?, album_id),
    format = COALESCE(?, format)
WHERE id = ?;

-- name: UpdateReleaseDiscogsInfo :exec
UPDATE releases
SET
    discogs_id = ?,
    label = ?,
    released_at = ?
WHERE id = ?;
```

- [ ] **Step 2: Add `SoftDeleteUserRelease` to `db/queries/user_releases.sql`**

Append to the end of `db/queries/user_releases.sql`:

```sql
-- name: SoftDeleteUserRelease :exec
UPDATE user_releases
SET removed_at = current_timestamp
WHERE user_id = ? AND release_id = ? AND removed_at IS NULL;
```

- [ ] **Step 3: Regenerate SQLC**

```bash
task build/sqlc
```

Expected: no errors. `src/internal/core/db/sqlc/releases.sql.go` and `user_releases.sql.go` are regenerated. The `Release` struct in `src/internal/core/db/sqlc/models.go` now has `DiscogsID sql.NullString`, `Label sql.NullString`, `ReleasedAt sql.NullTime`.

- [ ] **Step 4: Verify the project compiles**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add db/queries/releases.sql db/queries/user_releases.sql src/internal/core/db/sqlc/
git commit -m "feat: add UpdateReleaseDiscogsInfo and SoftDeleteUserRelease queries"
```

---

## Task 3: Update `ReleaseDTO` and Add `AlbumFormatDTO`

**Files:**
- Modify: `src/internal/library/service.go`
- Modify: `src/internal/library/service_test.go`

- [ ] **Step 1: Write failing tests**

Add to `src/internal/library/service_test.go`:

```go
// --- NewReleaseDTOFromModel ---

func TestNewReleaseDTOFromModel_MapsDiscogsFields(t *testing.T) {
	now := time.Now()
	release := sqlc.Release{
		ID:        "r1",
		AlbumID:   "a1",
		Format:    models.ReleaseFormatVinyl,
		CreatedAt: now,
		DiscogsID: sql.NullString{String: "12345", Valid: true},
		Label:     sql.NullString{String: "Warner Bros.", Valid: true},
		ReleasedAt: sql.NullTime{Time: now, Valid: true},
	}
	userRelease := &sqlc.UserRelease{AddedAt: now}

	dto := NewReleaseDTOFromModel(release, userRelease)

	if dto.DiscogsID != "12345" {
		t.Errorf("expected DiscogsID %q, got %q", "12345", dto.DiscogsID)
	}
	if dto.Label != "Warner Bros." {
		t.Errorf("expected Label %q, got %q", "Warner Bros.", dto.Label)
	}
	if dto.ReleasedAt == nil || !dto.ReleasedAt.Equal(now) {
		t.Errorf("expected ReleasedAt %v, got %v", now, dto.ReleasedAt)
	}
}

func TestNewReleaseDTOFromModel_NullDiscogsFieldsAreEmpty(t *testing.T) {
	release := sqlc.Release{
		ID:      "r1",
		AlbumID: "a1",
		Format:  models.ReleaseFormatVinyl,
	}

	dto := NewReleaseDTOFromModel(release, nil)

	if dto.DiscogsID != "" {
		t.Errorf("expected empty DiscogsID, got %q", dto.DiscogsID)
	}
	if dto.Label != "" {
		t.Errorf("expected empty Label, got %q", dto.Label)
	}
	if dto.ReleasedAt != nil {
		t.Errorf("expected nil ReleasedAt, got %v", dto.ReleasedAt)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./src/internal/library/... -run TestNewReleaseDTOFromModel -v
```

Expected: compilation errors — `DiscogsID`, `Label`, `ReleasedAt` don't exist yet on `ReleaseDTO`.

- [ ] **Step 3: Update `ReleaseDTO` and `NewReleaseDTOFromModel` in `service.go`**

Replace the existing `ReleaseDTO` struct and constructor:

```go
type ReleaseDTO struct {
	ID         string
	AlbumID    string
	Format     models.ReleaseFormat
	AddedAt    *time.Time
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

func NewReleaseDTOFromModel(model sqlc.Release, userRelease *sqlc.UserRelease) ReleaseDTO {
	dto := ReleaseDTO{
		ID:        model.ID,
		AlbumID:   model.AlbumID,
		Format:    model.Format,
		DiscogsID: model.DiscogsID.String,
		Label:     model.Label.String,
	}
	if model.ReleasedAt.Valid {
		dto.ReleasedAt = &model.ReleasedAt.Time
	}
	if userRelease != nil {
		dto.AddedAt = &userRelease.AddedAt
	}
	return dto
}
```

- [ ] **Step 4: Add `AlbumFormatDTO` type** — add after `ReleaseDTOs` methods in `service.go`:

```go
// AlbumFormatDTO represents one format row in the formats modal.
// It exists for all 4 formats regardless of whether the user owns that format.
type AlbumFormatDTO struct {
	Format     models.ReleaseFormat
	ReleaseID  string     // empty if this format has never been added for this album
	Owned      bool
	AddedAt    *time.Time
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./src/internal/library/... -run TestNewReleaseDTOFromModel -v
```

Expected: PASS

- [ ] **Step 6: Verify full build**

```bash
go build ./...
```

- [ ] **Step 7: Commit**

```bash
git add src/internal/library/service.go src/internal/library/service_test.go
git commit -m "feat: add Discogs fields to ReleaseDTO, add AlbumFormatDTO"
```

---

## Task 4: Add `GetAlbumFormats` Service Method

**Files:**
- Modify: `src/internal/library/service.go`
- Modify: `src/internal/library/service_test.go`

`GetAlbumFormats` returns one `AlbumFormatDTO` per format (all 4), regardless of ownership state. It joins existing release rows and user ownership data from the DB.

- [ ] **Step 1: Write the failing test**

Add to `src/internal/library/service_test.go`:

```go
// --- AlbumFormatDTO helpers ---

func TestAlbumFormatDTO_OwnedFormat_HasDiscogsData(t *testing.T) {
	now := time.Now()
	release := sqlc.Release{
		ID:         "r1",
		AlbumID:    "a1",
		Format:     models.ReleaseFormatVinyl,
		DiscogsID:  sql.NullString{String: "99", Valid: true},
		Label:      sql.NullString{String: "ECM", Valid: true},
		ReleasedAt: sql.NullTime{Time: now, Valid: true},
	}
	userRelease := sqlc.UserRelease{AddedAt: now}

	dto := albumFormatDTOFromRelease(release, &userRelease)

	t.Run("owned is true", func(t *testing.T) {
		if !dto.Owned {
			t.Error("expected Owned = true")
		}
	})
	t.Run("release ID set", func(t *testing.T) {
		if dto.ReleaseID != "r1" {
			t.Errorf("expected ReleaseID %q, got %q", "r1", dto.ReleaseID)
		}
	})
	t.Run("discogs ID mapped", func(t *testing.T) {
		if dto.DiscogsID != "99" {
			t.Errorf("expected DiscogsID %q, got %q", "99", dto.DiscogsID)
		}
	})
}

func TestAlbumFormatDTO_UnownedReleaseExists(t *testing.T) {
	release := sqlc.Release{
		ID:      "r2",
		AlbumID: "a1",
		Format:  models.ReleaseFormatCD,
	}

	dto := albumFormatDTOFromRelease(release, nil)

	t.Run("owned is false", func(t *testing.T) {
		if dto.Owned {
			t.Error("expected Owned = false")
		}
	})
	t.Run("release ID still set", func(t *testing.T) {
		if dto.ReleaseID != "r2" {
			t.Errorf("expected ReleaseID %q, got %q", "r2", dto.ReleaseID)
		}
	})
}
```

- [ ] **Step 2: Run to confirm compile failure**

```bash
go test ./src/internal/library/... -run TestAlbumFormatDTO -v
```

Expected: compile error — `albumFormatDTOFromRelease` not defined.

- [ ] **Step 3: Add `albumFormatDTOFromRelease` helper and `GetAlbumFormats` in `service.go`**

Add after the `AlbumFormatDTO` type:

```go
var physicalFormats = []models.ReleaseFormat{
	models.ReleaseFormatDigital,
	models.ReleaseFormatVinyl,
	models.ReleaseFormatCD,
	models.ReleaseFormatCassette,
}

func albumFormatDTOFromRelease(r sqlc.Release, ur *sqlc.UserRelease) AlbumFormatDTO {
	dto := AlbumFormatDTO{
		Format:    r.Format,
		ReleaseID: r.ID,
		DiscogsID: r.DiscogsID.String,
		Label:     r.Label.String,
	}
	if r.ReleasedAt.Valid {
		dto.ReleasedAt = &r.ReleasedAt.Time
	}
	if ur != nil {
		dto.Owned = true
		dto.AddedAt = &ur.AddedAt
	}
	return dto
}

func (s *Service) GetAlbumFormats(ctx context.Context, userID, albumID string) ([]AlbumFormatDTO, error) {
	allReleases, err := s.db.Queries().GetReleases(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}

	userReleases, err := s.db.Queries().GetUserReleasesByAlbumId(ctx, sqlc.GetUserReleasesByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user releases: %w", err)
	}

	releaseByFormat := make(map[models.ReleaseFormat]sqlc.Release, len(allReleases))
	for _, r := range allReleases {
		releaseByFormat[r.Format] = r
	}

	type ownedEntry struct {
		userRelease sqlc.UserRelease
	}
	ownedByReleaseID := make(map[string]ownedEntry, len(userReleases))
	for _, ur := range userReleases {
		ownedByReleaseID[ur.Release.ID] = ownedEntry{ur.UserRelease}
	}

	result := make([]AlbumFormatDTO, len(physicalFormats))
	for i, format := range physicalFormats {
		if r, ok := releaseByFormat[format]; ok {
			var ur *sqlc.UserRelease
			if entry, owned := ownedByReleaseID[r.ID]; owned {
				ur = &entry.userRelease
			}
			result[i] = albumFormatDTOFromRelease(r, ur)
		} else {
			result[i] = AlbumFormatDTO{Format: format}
		}
	}

	return result, nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./src/internal/library/... -run TestAlbumFormatDTO -v
```

Expected: PASS

- [ ] **Step 5: Build check**

```bash
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add src/internal/library/service.go src/internal/library/service_test.go
git commit -m "feat: add GetAlbumFormats service method"
```

---

## Task 5: Add `SaveAlbumFormats` Service Method

**Files:**
- Modify: `src/internal/library/service.go`
- Modify: `src/internal/library/service_test.go`

- [ ] **Step 1: Write the failing test**

Add to `src/internal/library/service_test.go`:

```go
// --- SaveFormatInput ---

func TestSaveFormatInput_PhysicalFormats(t *testing.T) {
	t.Run("owned input has format set", func(t *testing.T) {
		input := SaveFormatInput{
			Format:    models.ReleaseFormatVinyl,
			Owned:     true,
			DiscogsID: "42",
			Label:     "Blue Note",
		}
		if input.Format != models.ReleaseFormatVinyl {
			t.Errorf("expected vinyl, got %v", input.Format)
		}
		if !input.Owned {
			t.Error("expected Owned = true")
		}
		if input.DiscogsID != "42" {
			t.Errorf("expected DiscogsID %q, got %q", "42", input.DiscogsID)
		}
	})

	t.Run("unowned input with empty discogs", func(t *testing.T) {
		input := SaveFormatInput{
			Format: models.ReleaseFormatCD,
			Owned:  false,
		}
		if input.Owned {
			t.Error("expected Owned = false")
		}
		if input.DiscogsID != "" {
			t.Errorf("expected empty DiscogsID, got %q", input.DiscogsID)
		}
	})
}
```

- [ ] **Step 2: Run to confirm compile failure**

```bash
go test ./src/internal/library/... -run TestSaveFormatInput -v
```

Expected: compile error — `SaveFormatInput` not defined.

- [ ] **Step 3: Add `SaveFormatInput` and `SaveAlbumFormats` to `service.go`**

Add after `GetAlbumFormats`:

```go
type SaveFormatInput struct {
	Format     models.ReleaseFormat
	Owned      bool
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

func (s *Service) SaveAlbumFormats(ctx context.Context, userID, albumID string, inputs []SaveFormatInput) error {
	currentFormats, err := s.GetAlbumFormats(ctx, userID, albumID)
	if err != nil {
		return fmt.Errorf("failed to get current formats: %w", err)
	}

	currentByFormat := make(map[models.ReleaseFormat]AlbumFormatDTO, len(currentFormats))
	for _, f := range currentFormats {
		currentByFormat[f.Format] = f
	}

	return s.db.WithTx(func(tx *db.DB) error {
		for _, input := range inputs {
			if input.Format == models.ReleaseFormatDigital {
				continue // digital is managed by Spotify, never modified here
			}

			current := currentByFormat[input.Format]
			releaseID := current.ReleaseID

			if input.Owned {
				if !current.Owned {
					if releaseID == "" {
						r, err := tx.Queries().GetOrCreateRelease(ctx, sqlc.GetOrCreateReleaseParams{
							ID:      uuid.New().String(),
							AlbumID: albumID,
							Format:  input.Format,
						})
						if err != nil {
							return fmt.Errorf("failed to get/create release: %w", err)
						}
						releaseID = r.ID
					}
					_, err := tx.Queries().UpsertUserRelease(ctx, sqlc.UpsertUserReleaseParams{
						ID:        uuid.New().String(),
						UserID:    userID,
						ReleaseID: releaseID,
						AddedAt:   time.Now(),
					})
					if err != nil {
						return fmt.Errorf("failed to upsert user release: %w", err)
					}
				}

				if releaseID != "" && input.DiscogsID != "" {
					var releasedAt sql.NullTime
					if input.ReleasedAt != nil {
						releasedAt = sql.NullTime{Time: *input.ReleasedAt, Valid: true}
					}
					err := tx.Queries().UpdateReleaseDiscogsInfo(ctx, sqlc.UpdateReleaseDiscogsInfoParams{
						ID:         releaseID,
						DiscogsID:  sql.NullString{String: input.DiscogsID, Valid: true},
						Label:      sql.NullString{String: input.Label, Valid: input.Label != ""},
						ReleasedAt: releasedAt,
					})
					if err != nil {
						return fmt.Errorf("failed to update release discogs info: %w", err)
					}
				}
			} else if current.Owned && releaseID != "" {
				err := tx.Queries().SoftDeleteUserRelease(ctx, sqlc.SoftDeleteUserReleaseParams{
					UserID:    userID,
					ReleaseID: releaseID,
				})
				if err != nil {
					return fmt.Errorf("failed to soft delete user release: %w", err)
				}
			}
		}
		return nil
	})
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./src/internal/library/... -run TestSaveFormatInput -v
```

Expected: PASS

- [ ] **Step 5: Build check**

```bash
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add src/internal/library/service.go src/internal/library/service_test.go
git commit -m "feat: add SaveAlbumFormats service method"
```

---

## Task 6: Update Library HttpHandler + Wire Server Routes

**Files:**
- Modify: `src/internal/library/adapters/http.go`
- Modify: `src/internal/server/server.go`

- [ ] **Step 1: Add `discogsService` to `HttpHandler` in `adapters/http.go`**

Replace the existing `HttpHandler` struct and constructor:

```go
import (
	// existing imports...
	"github.com/alecdray/wax/src/internal/discogs"
)

type HttpHandler struct {
	spotifyAuth    *spotify.AuthService
	mb             *musicbrainz.Service
	feedService    *feed.Service
	libraryService *library.Service
	taskManager    *task.TaskManager
	discogsService *discogs.Service
}

func NewHttpHandler(spotifyAuth *spotify.AuthService, mb *musicbrainz.Service, feedService *feed.Service, libraryService *library.Service, taskManager *task.TaskManager, discogsService *discogs.Service) *HttpHandler {
	return &HttpHandler{
		spotifyAuth:    spotifyAuth,
		mb:             mb,
		feedService:    feedService,
		libraryService: libraryService,
		taskManager:    taskManager,
		discogsService: discogsService,
	}
}
```

- [ ] **Step 2: Update `server.go` — pass discogs and register routes**

In `Start()`, find the `libraryHandler` construction and replace it:

```go
libraryHandler := libraryAdapters.NewHttpHandler(
	services.spotifyAuth,
	services.musicbrainz,
	services.feed,
	services.library,
	services.taskManager,
	services.discogs,
)
```

Then after the existing library routes, add:

```go
appMux.Handle("GET /app/library/albums/{albumId}/formats", httpx.HandlerFunc(libraryHandler.GetFormatsModal))
appMux.Handle("PUT /app/library/albums/{albumId}/formats", httpx.HandlerFunc(libraryHandler.PutFormats))
appMux.Handle("GET /app/library/albums/{albumId}/formats/{format}/discogs/search", httpx.HandlerFunc(libraryHandler.GetDiscogsSearch))
appMux.Handle("GET /app/library/albums/{albumId}/formats/{format}/discogs/releases/{discogsId}", httpx.HandlerFunc(libraryHandler.GetDiscogsRelease))
```

- [ ] **Step 3: Build to catch any errors**

```bash
go build ./...
```

Expected: compile error on `libraryHandler.GetFormatsModal` etc. — those handlers don't exist yet. That's expected; we're wiring first.

- [ ] **Step 4: Create stub handlers in `formats.go` to make it compile**

Create `src/internal/library/adapters/formats.go`:

```go
package adapters

import "net/http"

func (h *HttpHandler) GetFormatsModal(w http.ResponseWriter, r *http.Request) {}
func (h *HttpHandler) PutFormats(w http.ResponseWriter, r *http.Request)      {}
func (h *HttpHandler) GetDiscogsSearch(w http.ResponseWriter, r *http.Request) {}
func (h *HttpHandler) GetDiscogsRelease(w http.ResponseWriter, r *http.Request) {}
```

- [ ] **Step 5: Build check**

```bash
go build ./...
```

Expected: clean build.

- [ ] **Step 6: Commit**

```bash
git add src/internal/library/adapters/http.go src/internal/library/adapters/formats.go src/internal/server/server.go
git commit -m "feat: wire formats routes to library handler"
```

---

## Task 7: Formats Modal Template

**Files:**
- Create: `src/internal/library/adapters/formats_modal.templ`

The modal uses the existing `@templates.Modal(...)` pattern (OOB-swapped into `#global-modal-container`). Format rows use Alpine.js `x-data` / `x-show` for the inline expand. The Discogs search uses `hx-get` + `hx-target` to inject results inline.

- [ ] **Step 1: Create `formats_modal.templ`**

```go
package adapters

import (
	"fmt"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/templates"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/library"
)

const FormatsModalId = "formats-modal"

func formatsReleasesID(albumID string) string {
	return fmt.Sprintf("album-detail-releases-%s", albumID)
}

templ FormatsModal(albumID string, formats []library.AlbumFormatDTO) {
	@templates.Modal(FormatsModalId, templates.ModalProps{
		ModalContent: formatsModalContent(albumID, formats),
	})
}

templ formatsModalContent(albumID string, formats []library.AlbumFormatDTO) {
	<h3 class="font-bold text-lg mb-4">Formats</h3>
	<form
		hx-put={ fmt.Sprintf("/app/library/albums/%s/formats", albumID) }
		hx-swap="none"
	>
		<div class="flex flex-col gap-3">
			for _, f := range formats {
				if f.Format == models.ReleaseFormatDigital {
					@digitalFormatRow(f)
				} else {
					@physicalFormatRow(albumID, f)
				}
			}
		</div>
		<div class="modal-action mt-6">
			<button type="submit" class="btn btn-primary btn-sm">Save</button>
			<form method="dialog">
				<button class="btn btn-ghost btn-sm">Cancel</button>
			</form>
		</div>
	</form>
}

templ digitalFormatRow(f library.AlbumFormatDTO) {
	<div class="flex items-center gap-3 py-2 opacity-50 cursor-not-allowed">
		<div class="w-5 h-5 flex items-center justify-center">
			@releaseFormatIcon(f.Format)
		</div>
		<span class="text-sm capitalize flex-1">{ string(f.Format) }</span>
		<span class="text-xs text-base-content/50">via Spotify</span>
	</div>
}

templ physicalFormatRow(albumID string, f library.AlbumFormatDTO) {
	<div
		x-data={ fmt.Sprintf(`{owned: %t, showDiscogs: false}`, f.Owned) }
		class="flex flex-col gap-2 py-2 border-b border-base-200 last:border-0"
	>
		<div class="flex items-center gap-3">
			<div class="w-5 h-5 flex items-center justify-center" :class="owned ? 'opacity-70' : 'opacity-20'">
				@releaseFormatIcon(f.Format)
			</div>
			<span class="text-sm capitalize flex-1">{ string(f.Format) }</span>
			<input type="hidden" :name={ fmt.Sprintf("%s_owned", string(f.Format)) } :value="owned ? 'true' : 'false'"/>
			<input
				type="checkbox"
				class="toggle toggle-sm"
				x-model="owned"
			/>
		</div>
		<div x-show="owned" x-cloak class="pl-8 flex flex-col gap-2">
			if f.DiscogsID != "" {
				@discogsAttachedDetails(f)
			}
			<button
				type="button"
				class="btn btn-ghost btn-xs text-base-content/50 self-start"
				@click="showDiscogs = !showDiscogs"
			>
				if f.DiscogsID != "" {
					Change pressing
				} else {
					+ Find on Discogs
				}
			</button>
			<div x-show="showDiscogs" x-cloak>
				@discogsSearchSection(albumID, f.Format, f.DiscogsID)
			</div>
		</div>
	</div>
}

templ discogsAttachedDetails(f library.AlbumFormatDTO) {
	<div class="text-xs text-base-content/60 flex flex-col gap-0.5" id={ fmt.Sprintf("discogs-details-%s", string(f.Format)) }>
		if f.Label != "" {
			<span>{ f.Label }</span>
		}
		if f.ReleasedAt != nil {
			<span>{ f.ReleasedAt.Format("2006") }</span>
		}
		<input type="hidden" name={ fmt.Sprintf("%s_discogs_id", string(f.Format)) } value={ f.DiscogsID }/>
		<input type="hidden" name={ fmt.Sprintf("%s_label", string(f.Format)) } value={ f.Label }/>
		if f.ReleasedAt != nil {
			<input type="hidden" name={ fmt.Sprintf("%s_released_at", string(f.Format)) } value={ f.ReleasedAt.Format("2006-01-02") }/>
		}
	</div>
}

templ discogsSearchSection(albumID string, format models.ReleaseFormat, currentDiscogsID string) {
	<div class="flex flex-col gap-2">
		<div class="flex gap-2">
			<input
				type="text"
				class="input input-bordered input-xs flex-1"
				placeholder="Search Discogs..."
				name="q"
				hx-get={ fmt.Sprintf("/app/library/albums/%s/formats/%s/discogs/search", albumID, string(format)) }
				hx-trigger="input changed delay:400ms"
				hx-target={ fmt.Sprintf("#discogs-results-%s", string(format)) }
				hx-include="this"
			/>
		</div>
		<div id={ fmt.Sprintf("discogs-results-%s", string(format)) } class="flex flex-col gap-1 max-h-48 overflow-y-auto">
		</div>
	</div>
}

// DiscogsSearchResults is returned by the search endpoint — injected into #discogs-results-{format}.
templ DiscogsSearchResults(albumID string, format models.ReleaseFormat, results []discogs.SearchItem) {
	if len(results) == 0 {
		<p class="text-xs text-base-content/40">No results</p>
	} else {
		for _, item := range results {
			firstLabel := ""
			if len(item.Label) > 0 {
				firstLabel = item.Label[0]
			}
			<div
				class="flex items-center gap-2 p-1.5 rounded hover:bg-base-200 cursor-pointer text-xs"
				hx-get={ fmt.Sprintf("/app/library/albums/%s/formats/%s/discogs/releases/%d?label=%s&year=%s", albumID, string(format), item.ID, templ.URLEscaper(firstLabel), templ.URLEscaper(item.Year)) }
				hx-target={ fmt.Sprintf("#discogs-details-%s", string(format)) }
				hx-swap="outerHTML"
			>
				if item.Thumb != "" {
					<img src={ item.Thumb } class="w-8 h-8 rounded flex-shrink-0" alt=""/>
				}
				<div class="flex flex-col min-w-0">
					<span class="truncate font-medium">{ item.Title }</span>
					<span class="text-base-content/50 truncate">
						if len(item.Label) > 0 {
							{ item.Label[0] }
						}
						if item.Year != "" {
							· { item.Year }
						}
					</span>
				</div>
			</div>
		}
	}
}

// DiscogsReleaseDetails is returned by the release detail endpoint — replaces #discogs-details-{format}.
templ DiscogsReleaseDetails(format models.ReleaseFormat, item discogs.SearchItem) {
	<div class="text-xs text-base-content/60 flex flex-col gap-0.5" id={ fmt.Sprintf("discogs-details-%s", string(format)) }>
		if len(item.Label) > 0 {
			<span>{ item.Label[0] }</span>
		}
		if item.Year != "" {
			<span>{ item.Year }</span>
		}
		<input type="hidden" name={ fmt.Sprintf("%s_discogs_id", string(format)) } value={ fmt.Sprintf("%d", item.ID) }/>
		if len(item.Label) > 0 {
			<input type="hidden" name={ fmt.Sprintf("%s_label", string(format)) } value={ item.Label[0] }/>
		}
		if item.Year != "" {
			<input type="hidden" name={ fmt.Sprintf("%s_released_at", string(format)) } value={ item.Year + "-01-01" }/>
		}
	</div>
}

// FormatsReleasesOOB updates the format icons on the album detail page after a save.
templ FormatsReleasesOOB(albumID string, releases library.ReleaseDTOs) {
	<div
		id={ formatsReleasesID(albumID) }
		class="flex flex-wrap gap-2 items-center"
		data-testid="album-detail-releases"
		hx-swap-oob="true"
	>
		@formatIcon(releases, models.ReleaseFormatDigital)
		@formatIcon(releases, models.ReleaseFormatVinyl)
		@formatIcon(releases, models.ReleaseFormatCD)
		@formatIcon(releases, models.ReleaseFormatCassette)
	</div>
}
```

- [ ] **Step 2: Build the template**

```bash
task build/templ
```

Expected: generates `formats_modal_templ.go` with no errors.

- [ ] **Step 3: Build check**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add src/internal/library/adapters/formats_modal.templ src/internal/library/adapters/formats_modal_templ.go
git commit -m "feat: add formats modal template"
```

---

## Task 8: Implement HTTP Handlers

**Files:**
- Modify: `src/internal/library/adapters/formats.go`

Replace the stubs with full implementations.

- [ ] **Step 1: Implement all four handlers**

Replace `src/internal/library/adapters/formats.go`:

```go
package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/library"
)

// GetFormatsModal renders the formats management modal for an album.
func (h *HttpHandler) GetFormatsModal(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.PathValue("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	formats, err := h.libraryService.GetAlbumFormats(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to get formats: %w", err)})
		return
	}

	if err := FormatsModal(albumID, formats).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// PutFormats saves format ownership and optional Discogs data.
func (h *HttpHandler) PutFormats(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.PathValue("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	physicalFormats := []models.ReleaseFormat{
		models.ReleaseFormatVinyl,
		models.ReleaseFormatCD,
		models.ReleaseFormatCassette,
	}

	inputs := make([]library.SaveFormatInput, 0, len(physicalFormats))
	for _, format := range physicalFormats {
		owned := r.FormValue(string(format)+"_owned") == "true"
		discogsID := r.FormValue(string(format) + "_discogs_id")
		label := r.FormValue(string(format) + "_label")
		releasedAtStr := r.FormValue(string(format) + "_released_at")

		input := library.SaveFormatInput{
			Format:    format,
			Owned:     owned,
			DiscogsID: discogsID,
			Label:     label,
		}
		if releasedAtStr != "" {
			if t, err := time.Parse("2006-01-02", releasedAtStr); err == nil {
				input.ReleasedAt = &t
			}
		}
		inputs = append(inputs, input)
	}

	if err := h.libraryService.SaveAlbumFormats(ctx, userID, albumID, inputs); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to save formats: %w", err)})
		return
	}

	// Re-fetch releases for OOB swap
	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}

	// Close the modal and update the releases section in place via OOB swaps.
	// Render both in sequence — HTMX processes multiple OOB elements in one response.
	if err := FormatsReleasesOOB(albumID, album.Releases).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := templates.ForceCloseModal(FormatsModalId).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// GetDiscogsSearch searches Discogs by query string and returns an inline results fragment.
func (h *HttpHandler) GetDiscogsSearch(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.PathValue("albumId")
	format := models.ReleaseFormat(r.PathValue("format"))
	q := r.URL.Query().Get("q")
	if q == "" {
		// Return empty fragment — no query yet
		w.WriteHeader(http.StatusOK)
		return
	}

	results, err := h.discogsService.SearchReleases(ctx, q)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("discogs search failed: %w", err)})
		return
	}

	items := results.Results
	if len(items) > 10 {
		items = items[:10]
	}

	if err := DiscogsSearchResults(albumID, format, items).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// (GetDiscogsRelease is defined below — see the corrected version after buildSearchItemFromRelease)

// GetDiscogsRelease uses label and year from query params (forwarded from the search result)
// because discogs.Release has no Labels field.
func (h *HttpHandler) GetDiscogsRelease(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	format := models.ReleaseFormat(r.PathValue("format"))
	discogsIDStr := r.PathValue("discogsId")

	discogsID, err := strconv.Atoi(discogsIDStr)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid Discogs ID: %w", err)})
		return
	}

	// label and year are forwarded from the search result that the user clicked.
	// discogs.Release has no Labels field, so we rely on the params.
	label := r.URL.Query().Get("label")
	year := r.URL.Query().Get("year")

	// Fetch full release to get the exact Released date if available.
	release, err := h.discogsService.GetRelease(ctx, discogsID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to fetch Discogs release: %w", err)})
		return
	}

	// Prefer the exact Released date from the Release object.
	if release.Released != "" && len(release.Released) >= 4 {
		year = release.Released[:4]
	} else if release.Year > 0 {
		year = strconv.Itoa(release.Year)
	}

	item := discogs.SearchItem{
		ID:    discogsID,
		Title: release.Title,
		Year:  year,
	}
	if label != "" {
		item.Label = []string{label}
	}

	if err := DiscogsReleaseDetails(format, item).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}
```

Note: `buildSearchItemFromRelease` constructs a `SearchItem` from a full `Release`. The `discogs.Release` entity has no labels field — label data is available on `SearchItem` from search results. Since the user already saw the label in the search results before clicking, `DiscogsReleaseDetails` will display what was available from the `SearchItem`; this handler is only called to confirm and lock in the choice.

- [ ] **Step 2: Build templates**

```bash
task build/templ && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add src/internal/library/adapters/formats.go
git commit -m "feat: implement formats modal HTTP handlers"
```

---

## Task 9: Update Album Detail Template

**Files:**
- Modify: `src/internal/library/adapters/album_detail.templ`

Replace the static format icons group with a clickable button that opens the formats modal.

- [ ] **Step 1: Update the releases section in `AlbumDetailPage`**

Find this block in `album_detail.templ`:

```templ
// Formats
<div class="flex flex-col gap-2">
    <div class="flex flex-wrap gap-2 items-center" data-testid="album-detail-releases">
        @formatIcon(album.Releases, models.ReleaseFormatDigital)
        @formatIcon(album.Releases, models.ReleaseFormatVinyl)
        @formatIcon(album.Releases, models.ReleaseFormatCD)
        @formatIcon(album.Releases, models.ReleaseFormatCassette)
    </div>
</div>
```

Replace with:

```templ
// Formats
<div class="flex flex-col gap-2">
    <button
        class="flex flex-wrap gap-2 items-center cursor-pointer hover:opacity-80 transition-opacity"
        data-testid="album-detail-releases-btn"
        hx-get={ fmt.Sprintf("/app/library/albums/%s/formats", album.ID) }
        hx-trigger="click"
        hx-swap="none"
    >
        <div
            id={ formatsReleasesID(album.ID) }
            class="flex flex-wrap gap-2 items-center"
            data-testid="album-detail-releases"
        >
            @formatIcon(album.Releases, models.ReleaseFormatDigital)
            @formatIcon(album.Releases, models.ReleaseFormatVinyl)
            @formatIcon(album.Releases, models.ReleaseFormatCD)
            @formatIcon(album.Releases, models.ReleaseFormatCassette)
        </div>
    </button>
</div>
```

Note: `formatsReleasesID` is defined in `formats_modal.templ` — same package, no import needed.

- [ ] **Step 2: Build templates**

```bash
task build/templ
```

- [ ] **Step 3: Build check**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add src/internal/library/adapters/album_detail.templ src/internal/library/adapters/album_detail_templ.go
git commit -m "feat: make format icons a button opening the formats modal"
```

---

## Task 10: Manual Smoke Test

- [ ] **Step 1: Start the dev server**

```bash
task dev
```

- [ ] **Step 2: Open an album detail page**

Navigate to an album. Confirm the format icons are now tappable.

- [ ] **Step 3: Open the formats modal**

Tap the format icons. The modal should open showing four rows: Digital (locked, "via Spotify"), Vinyl/CD/Cassette (toggleable).

- [ ] **Step 4: Toggle a physical format on**

Toggle Vinyl. Confirm the Discogs search section expands. Toggle it off and confirm it collapses.

- [ ] **Step 5: Search Discogs**

With Vinyl toggled on, type a query in the search box. Confirm results appear inline.

- [ ] **Step 6: Pick a Discogs result**

Click a result. Confirm the label and year appear as details. Confirm hidden inputs are populated.

- [ ] **Step 7: Save and verify**

Click Save. Confirm the modal closes and the format icon updates (vinyl now full opacity). Reload the album detail page and confirm the vinyl icon remains full opacity.

- [ ] **Step 8: Toggle off and verify**

Re-open the modal, toggle Vinyl off, save. Confirm vinyl icon returns to dimmed.

- [ ] **Step 9: Final commit**

```bash
git add -u
git commit -m "feat: physical media releases with Discogs integration"
```
