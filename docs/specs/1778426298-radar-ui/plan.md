# Radar UI/UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the user-facing surface for the radar feature: a `/discover` page with a Spotify search, a radar carousel, and HTMX-driven add/remove/promote actions. Promote-to-library also pushes the album into the user's Spotify saved library.

**Architecture:** All UI lives inside `library/adapters/`. New service methods on `library.Service` and `spotify.Service`. One new repo SQL query (`GetUserAlbumStateBySpotifyIds`) and a small refactor of `Repo.AddAlbumToCollection` to extract the metadata-import portion so radar imports can reuse it without writing `user_releases` rows. HTMX swaps drive the in-page reactivity; a new body-level `radarUpdated` event keeps the radar carousel in sync with search-result-row state changes.

**Tech Stack:** Go 1.25, SQLite, SQLC, templ, HTMX + DaisyUI, `task` runner, `github.com/zmb3/spotify/v2`. All commands invoked via `task <name>`.

**Spec:** `docs/specs/1778426298-radar-ui/design.md` — read first.

**Working directory:** `/Users/shmoopy/workshop/workbench/wax/wt-4c45` (worktree on branch `feat/radar-discover-ui`). Run all `task` and `git` commands from there.

**Notes for the engineer:**
- Backend (`user_album_radar` table, `Service.AddAlbumToRadar`, `AcquireFromWishlist`, etc.) is already shipped from spec `1778379187-album-states`. This plan only adds the UI layer + the two Spotify methods (`SearchAlbums`, `AddAlbumToSavedLibrary`) it depends on.
- Worktree environment: copy `.env` from the main project (`cp /Users/shmoopy/workshop/projects/wax/.env .env`) and the dev DB (`cp /Users/shmoopy/workshop/projects/wax/tmp/db.sql ./tmp/db.sql`) before running `task dev`. Run `npm install` if `node_modules` is missing.
- The project tests pure logic only — there are **no DB-backed Go tests** today. We will write unit tests for `SearchAlbumsForDiscover`'s state-merging logic (pure once you stub the dependencies) but skip Playwright e2e this round to match the album-states plan's convention of "manual verification + build success serve as the gate."
- After modifying `db/queries/*.sql`, run `task build/sqlc` to regenerate `src/internal/core/db/sqlc/`.
- After modifying `*.templ` files, run `task build/templ` to regenerate `*_templ.go`. Never edit `*_templ.go` by hand.
- Before any commit: `task test && task build` — both must pass.
- Some tasks combine related code into one commit (the plan says so explicitly when applicable). Otherwise commit per task.

---

## File Structure

**Modify:**
- `src/internal/spotify/service.go` — add `SearchAlbums` and `AddAlbumToSavedLibrary` methods.
- `src/internal/library/library.go` — add `DiscoverAlbumState`, `UserAlbumStateRow`, `DiscoverResultDTO`.
- `src/internal/library/repo.go` — add `GetUserAlbumStateBySpotifyIDs`; extract `EnsureAlbumWithMetadata` helper from `AddAlbumToCollection`.
- `src/internal/library/service.go` — add `SearchAlbumsForDiscover`, `AddSpotifyAlbumToRadar`, `GetRadarAlbums`, `RemoveAlbumFromRadar`, `PromoteRadarToLibrary`.
- `src/internal/library/service_test.go` — add unit tests for `mergeDiscoverState` (pure helper).
- `src/internal/library/adapters/http.go` — add 6 new handler methods on `HttpHandler`.
- `src/internal/library/adapters/routes.go` — register the new routes.
- `src/internal/library/adapters/dashboard.templ` — replace `DashboardHeaderBar` with shared `LibraryHeaderBar`.
- `src/internal/library/adapters/album_detail.templ` — replace `AlbumDetailHeaderBar` with shared `LibraryHeaderBar`.
- `src/internal/core/templates/icons.templ` — add `CompassIcon`.
- `db/queries/user_releases.sql` — add `GetUserAlbumStateBySpotifyIds` query.

**Create:**
- `db/queries/user_releases.sql` (modify, see above) — query lives here because the state lookup joins `albums` and `user_releases`.
- `src/internal/library/adapters/discover.templ` — `DiscoverPage`, `LibraryHeaderBar`, `RadarCarousel`, `DiscoverSearchResults`, `discoverSearchResultRow`, `discoverRadarCardMenu`.

**Auto-generated (do not edit by hand, but verify after build):**
- `src/internal/core/db/sqlc/user_releases.sql.go` — gains `GetUserAlbumStateBySpotifyIds`.
- `src/internal/library/adapters/discover_templ.go`
- `src/internal/library/adapters/dashboard_templ.go`
- `src/internal/library/adapters/album_detail_templ.go`
- `src/internal/core/templates/icons_templ.go`

---

## Task 1: `spotify.Service.AddAlbumToSavedLibrary`

**Files:**
- Modify: `src/internal/spotify/service.go`

The vendor SDK exposes `Client.AddAlbumsToLibrary(ctx, ids...)`. We mirror the shape of `RemoveAlbumFromSavedLibrary` (line 138).

- [ ] **Step 1: Append the method**

After the existing `RemoveAlbumFromSavedLibrary` block (line 150), add:

```go
// AddAlbumToSavedLibrary saves an album to the user's Spotify saved library.
// Mirrors RemoveAlbumFromSavedLibrary; uses the SDK directly (no raw HTTP).
func (s *Service) AddAlbumToSavedLibrary(ctx contextx.ContextX, userId, spotifyId string) error {
    client, err := s.Client(ctx, userId)
    if err != nil {
        return fmt.Errorf("failed to get spotify client: %w", err)
    }
    if err := client.AddAlbumsToLibrary(ctx, spotify.ID(spotifyId)); err != nil {
        return fmt.Errorf("failed to add album to spotify saved library: %w", err)
    }
    return nil
}
```

- [ ] **Step 2: Build**

Run: `task build`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add src/internal/spotify/service.go
git commit -m "feat(spotify): AddAlbumToSavedLibrary"
```

---

## Task 2: `spotify.Service.SearchAlbums`

**Files:**
- Modify: `src/internal/spotify/service.go`

Wraps `client.Search(ctx, query, spotify.SearchTypeAlbum, opts...)`. Returns `[]spotify.SimpleAlbum`.

- [ ] **Step 1: Append the method**

```go
// SearchAlbums runs a Spotify catalog search restricted to albums.
// limit is clamped to the Spotify API max of 50.
func (s *Service) SearchAlbums(ctx contextx.ContextX, userId, query string, limit int) ([]spotify.SimpleAlbum, error) {
    if query == "" {
        return nil, nil
    }
    if limit <= 0 {
        limit = 20
    }
    if limit > 50 {
        limit = 50
    }
    client, err := s.Client(ctx, userId)
    if err != nil {
        return nil, fmt.Errorf("failed to get spotify client: %w", err)
    }
    result, err := client.Search(ctx, query, spotify.SearchTypeAlbum, spotify.Limit(limit))
    if err != nil {
        return nil, fmt.Errorf("spotify album search failed: %w", err)
    }
    if result == nil || result.Albums == nil {
        return nil, nil
    }
    return result.Albums.Albums, nil
}
```

- [ ] **Step 2: Build**

Run: `task build`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add src/internal/spotify/service.go
git commit -m "feat(spotify): SearchAlbums"
```

---

## Task 3: Discover domain types

**Files:**
- Modify: `src/internal/library/library.go`

Add the discover-page DTOs at the end of `library.go`. They live here because the file is the package's "topic file" for all DTOs (per the domain-module archetype) and these types share `ArtistDTO` from the same file.

- [ ] **Step 1: Append the types**

```go
// DiscoverAlbumState describes whether the caller already has a relationship
// with an album (used to render Spotify search results in /discover).
type DiscoverAlbumState string

const (
    DiscoverAlbumStateNone      DiscoverAlbumState = "none"
    DiscoverAlbumStateInLibrary DiscoverAlbumState = "in_library"
    DiscoverAlbumStateOnRadar   DiscoverAlbumState = "on_radar"
    DiscoverAlbumStateRemoved   DiscoverAlbumState = "removed"
)

// UserAlbumStateRow is the per-album result of GetUserAlbumStateBySpotifyIDs.
// AlbumID is the wax album row's primary key (always populated when present
// in the map).
type UserAlbumStateRow struct {
    AlbumID string
    State   DiscoverAlbumState
}

// DiscoverResultDTO is one row in the /discover page's search results.
// AlbumID is empty when State == "none" (the album has no wax row yet).
type DiscoverResultDTO struct {
    SpotifyID string
    Title     string
    Artists   []ArtistDTO
    ImageURL  string
    State     DiscoverAlbumState
    AlbumID   string
}
```

- [ ] **Step 2: Build**

Run: `go build ./src/internal/library`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add src/internal/library/library.go
git commit -m "feat(library): discover DTOs and DiscoverAlbumState"
```

---

## Task 4: SQL query — bulk Spotify-ID state lookup

**Files:**
- Modify: `db/queries/user_releases.sql`

For each input Spotify ID, we want one row per `(album, state)` so the service can build the map. Three sources of state, joined together:

1. `user_releases.status = 'owned'` → `in_library`
2. `user_album_radar` → `on_radar`
3. `user_releases.status = 'removed'` (only if no `'owned'` row for the same album) → `removed`

We let the service collapse state for albums that have multiple matches (e.g., owned + radar shouldn't co-exist by invariant, but the query stays correct anyway).

- [ ] **Step 1: Append the query**

Append to `db/queries/user_releases.sql`:

```sql
-- name: GetUserAlbumStateBySpotifyIds :many
-- For each Spotify ID, returns the album's wax ID plus the user's state for
-- that album: 'owned', 'removed', or 'on_radar'. An album appears at most
-- twice (e.g., 'removed' + 'on_radar' won't co-exist by invariant, but the
-- query stays correct if invariants drift). The service collapses to one
-- DiscoverAlbumState per album.
SELECT albums.id        AS album_id,
       albums.spotify_id AS spotify_id,
       'owned'          AS state
FROM albums
JOIN user_releases ON user_releases.release_id IN (
    SELECT id FROM releases WHERE album_id = albums.id
)
WHERE user_releases.user_id = ?
  AND user_releases.status = 'owned'
  AND albums.spotify_id IN (sqlc.slice('spotify_ids'))

UNION ALL

SELECT albums.id        AS album_id,
       albums.spotify_id AS spotify_id,
       'removed'        AS state
FROM albums
JOIN user_releases ON user_releases.release_id IN (
    SELECT id FROM releases WHERE album_id = albums.id
)
WHERE user_releases.user_id = ?
  AND user_releases.status = 'removed'
  AND albums.spotify_id IN (sqlc.slice('spotify_ids'))
  AND NOT EXISTS (
      SELECT 1 FROM user_releases ur2
      JOIN releases r2 ON r2.id = ur2.release_id
      WHERE ur2.user_id = user_releases.user_id
        AND r2.album_id = albums.id
        AND ur2.status = 'owned'
  )

UNION ALL

SELECT albums.id        AS album_id,
       albums.spotify_id AS spotify_id,
       'on_radar'       AS state
FROM albums
JOIN user_album_radar ON user_album_radar.album_id = albums.id
WHERE user_album_radar.user_id = ?
  AND albums.spotify_id IN (sqlc.slice('spotify_ids'));
```

- [ ] **Step 2: Regenerate sqlc**

Run: `task build/sqlc`
Expected: no errors. `src/internal/core/db/sqlc/user_releases.sql.go` is regenerated with a new `GetUserAlbumStateBySpotifyIds` function.

- [ ] **Step 3: Verify generated code compiles**

Run: `go build ./src/internal/core/db/sqlc/...`
Expected: success.

- [ ] **Step 4: Commit**

```bash
git add db/queries/user_releases.sql src/internal/core/db/sqlc/user_releases.sql.go
git commit -m "feat(library): GetUserAlbumStateBySpotifyIds query"
```

---

## Task 5: Repo — `GetUserAlbumStateBySpotifyIDs` method

**Files:**
- Modify: `src/internal/library/repo.go`

Wraps the sqlc query and collapses multi-row results to one `UserAlbumStateRow` per Spotify ID using state precedence: `in_library` > `on_radar` > `removed`.

- [ ] **Step 1: Append the method**

Append to `repo.go`:

```go
// GetUserAlbumStateBySpotifyIDs returns the caller's wax state for each
// Spotify ID. Missing keys mean the user has no row for that album. When an
// album would qualify for multiple states (defensive — invariants forbid it),
// in_library wins, then on_radar, then removed.
func (r *Repo) GetUserAlbumStateBySpotifyIDs(ctx context.Context, userID string, spotifyIDs []string) (map[string]UserAlbumStateRow, error) {
    if len(spotifyIDs) == 0 {
        return map[string]UserAlbumStateRow{}, nil
    }
    rows, err := r.q.GetUserAlbumStateBySpotifyIds(ctx, sqlc.GetUserAlbumStateBySpotifyIdsParams{
        UserID:     userID,
        UserID_2:   userID,
        UserID_3:   userID,
        SpotifyIds: spotifyIDs,
    })
    if err != nil {
        return nil, err
    }
    out := make(map[string]UserAlbumStateRow, len(rows))
    rank := func(s DiscoverAlbumState) int {
        switch s {
        case DiscoverAlbumStateInLibrary:
            return 3
        case DiscoverAlbumStateOnRadar:
            return 2
        case DiscoverAlbumStateRemoved:
            return 1
        default:
            return 0
        }
    }
    for _, row := range rows {
        next := UserAlbumStateRow{
            AlbumID: row.AlbumID,
            State:   DiscoverAlbumState(row.State),
        }
        if cur, ok := out[row.SpotifyID]; !ok || rank(next.State) > rank(cur.State) {
            out[row.SpotifyID] = next
        }
    }
    return out, nil
}
```

> **Note on `UserID_2` / `UserID_3`:** sqlc names repeated `?` placeholders with numeric suffixes. After running `task build/sqlc` in Task 4, open `user_releases.sql.go` and confirm the actual parameter struct field names. If sqlc named them differently (e.g. `UserID`, `UserID_1`, `UserID_2`), adjust the call here. The compile in Step 2 below catches any mismatch immediately.

- [ ] **Step 2: Build**

Run: `go build ./src/internal/library`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add src/internal/library/repo.go
git commit -m "feat(library): GetUserAlbumStateBySpotifyIDs repo method"
```

---

## Task 6: Repo — extract `EnsureAlbumWithMetadata`

**Files:**
- Modify: `src/internal/library/repo.go`

`AddAlbumToCollection` does two things: (a) imports album/artists/tracks/releases metadata, (b) writes a `user_releases` row for each release. We extract (a) so the radar-add flow can reuse it without writing `user_releases`.

- [ ] **Step 1: Refactor — split the function**

Locate the existing `AddAlbumToCollection` (around line 366). Replace the entire function with two functions:

```go
// EnsureAlbumWithMetadata creates or updates the album row, its artists, its
// tracks, and its releases — but does NOT touch user_releases or
// user_album_radar. Used by both the collection-add flow (which then writes
// user_releases) and the radar-add flow (which then writes user_album_radar).
//
// Returns the canonical AlbumDTO with sqlc-derived IDs filled in.
func (r *Repo) EnsureAlbumWithMetadata(ctx context.Context, album AlbumDTO) (AlbumDTO, error) {
    albumModel, err := r.q.GetOrCreateAlbum(ctx, sqlc.GetOrCreateAlbumParams{
        ID:        album.ID,
        SpotifyID: album.SpotifyID,
        Title:     album.Title,
        ImageUrl:  sql.NullString{String: album.ImageURL, Valid: album.ImageURL != ""},
    })
    if err != nil {
        return album, err
    }
    album = albumDTOFromModel(albumModel, album.Artists, album.Tracks, album.Releases, album.Rating)

    for i, track := range album.Tracks {
        trackModel, err := r.q.GetOrCreateTrack(ctx, sqlc.GetOrCreateTrackParams{
            ID:        track.ID,
            SpotifyID: track.SpotifyID,
            Title:     track.Title,
        })
        if err != nil {
            return album, err
        }
        if _, err := r.q.GetOrCreateAlbumTrack(ctx, sqlc.GetOrCreateAlbumTrackParams{
            AlbumID: albumModel.ID,
            TrackID: trackModel.ID,
        }); err != nil {
            return album, err
        }
        album.Tracks[i] = trackDTOFromModel(trackModel)
    }

    for i, artist := range album.Artists {
        artistModel, err := r.q.GetOrCreateArtist(ctx, sqlc.GetOrCreateArtistParams{
            ID:        artist.ID,
            SpotifyID: artist.SpotifyID,
            Name:      artist.Name,
        })
        if err != nil {
            return album, err
        }
        if _, err := r.q.GetOrCreateAlbumArtist(ctx, sqlc.GetOrCreateAlbumArtistParams{
            AlbumID:  albumModel.ID,
            ArtistID: artistModel.ID,
        }); err != nil {
            return album, err
        }
        album.Artists[i] = artistDTOFromModel(artistModel)
    }

    for i, release := range album.Releases {
        releaseModel, err := r.q.GetOrCreateRelease(ctx, sqlc.GetOrCreateReleaseParams{
            ID:      release.ID,
            AlbumID: albumModel.ID,
            Format:  release.Format,
        })
        if err != nil {
            return album, err
        }
        album.Releases[i] = releaseDTOFromModel(releaseModel, nil)
    }

    return album, nil
}

// AddAlbumToCollection imports the album's metadata and writes an owned
// user_releases row for every release on the AlbumDTO. The caller is
// responsible for clearing the album's radar entry (the cross-cutting rule
// lives at the service layer).
func (r *Repo) AddAlbumToCollection(ctx context.Context, userID string, album AlbumDTO) (AlbumDTO, error) {
    album, err := r.EnsureAlbumWithMetadata(ctx, album)
    if err != nil {
        return album, err
    }
    for i, release := range album.Releases {
        now := time.Now()
        userRelease, err := r.q.UpsertOwnedRelease(ctx, sqlc.UpsertOwnedReleaseParams{
            ID:              uuid.New().String(),
            UserID:          userID,
            ReleaseID:       release.ID,
            CreatedAt:       now,
            StatusUpdatedAt: now,
        })
        if err != nil {
            return album, err
        }
        album.Releases[i] = releaseDTOFromModel(sqlc.Release{
            ID:         release.ID,
            AlbumID:    album.ID,
            Format:     release.Format,
            DiscogsID:  sql.NullString{String: release.DiscogsID, Valid: release.DiscogsID != ""},
            Label:      sql.NullString{String: release.Label, Valid: release.Label != ""},
        }, &userRelease)
    }
    return album, nil
}
```

- [ ] **Step 2: Build**

Run: `task build`
Expected: success.

- [ ] **Step 3: Run existing unit tests**

Run: `task test`
Expected: all pass. The Spotify feed sync continues to work because `AddAlbumToCollection` retains its existing behaviour — only the inner factoring changed.

- [ ] **Step 4: Commit**

```bash
git add src/internal/library/repo.go
git commit -m "refactor(library): extract EnsureAlbumWithMetadata from AddAlbumToCollection"
```

---

## Task 7: Service — `GetRadarAlbums` and `RemoveAlbumFromRadar`

**Files:**
- Modify: `src/internal/library/service.go`

Two thin wrappers. `GetRadarAlbums` parallels existing `GetRecentlyPlayedAlbums` / `GetUnratedAlbums`. `RemoveAlbumFromRadar` exists at the repo layer (created in spec 1778379187) but isn't yet exposed at the service layer.

- [ ] **Step 1: Append the methods**

Append to `service.go` (after `AddAlbumToRadar`, around line 378):

```go
// GetRadarAlbums returns the caller's radar entries as fully-populated
// AlbumDTOs (artists set; tracks/releases left empty — radar entries have no
// release rows). Used by the discover page's radar carousel.
func (s *Service) GetRadarAlbums(ctx context.Context, userID string) ([]AlbumDTO, error) {
    _, albums, err := s.repo.GetRadarAlbums(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get radar albums: %w", err)
    }
    if len(albums) == 0 {
        return nil, nil
    }
    albumIDs := make([]string, len(albums))
    for i, a := range albums {
        albumIDs[i] = a.ID
    }
    artistsByAlbumID, err := s.repo.GetArtistsByAlbumIDs(ctx, albumIDs)
    if err != nil {
        return nil, fmt.Errorf("failed to get artists for radar albums: %w", err)
    }
    for i := range albums {
        albums[i].Artists = artistsByAlbumID[albums[i].ID]
    }
    return albums, nil
}

// RemoveAlbumFromRadar deletes the radar row. No-op if the user has no radar
// row for the album.
func (s *Service) RemoveAlbumFromRadar(ctx context.Context, userID, albumID string) error {
    return s.repo.RemoveAlbumFromRadar(ctx, userID, albumID)
}
```

- [ ] **Step 2: Verify `GetArtistsByAlbumIDs` exists**

Run: `grep -n "GetArtistsByAlbumIDs\b" src/internal/library/repo.go`
Expected: at least one match. If absent, replace the per-album loop in `GetRadarAlbums` with a per-album call to `s.repo.GetArtistsByAlbumID(ctx, albumID)` instead — that single-album method exists today (search confirms it at `repo.go` near `GetArtistsByAlbumID`).

If the bulk method exists, no change needed; keep the `GetArtistsByAlbumIDs` call as written.

- [ ] **Step 3: Build**

Run: `task build`
Expected: success.

- [ ] **Step 4: Commit**

```bash
git add src/internal/library/service.go
git commit -m "feat(library): GetRadarAlbums and RemoveAlbumFromRadar service methods"
```

---

## Task 8: Service — `AddSpotifyAlbumToRadar`

**Files:**
- Modify: `src/internal/library/service.go`
- Modify: `src/internal/spotify/service.go` (small helper)

`AddSpotifyAlbumToRadar` fetches the full album from Spotify, converts it into an `AlbumDTO`, ensures album+artists+tracks rows in a transaction, then calls the existing `Service.AddAlbumToRadar` (which has the `HasAnyUserReleaseForAlbum` guard). The release rows are intentionally **not** created — radar is pre-decision.

- [ ] **Step 1: Add a `GetFullAlbum` helper to spotify service**

Append to `src/internal/spotify/service.go`:

```go
// GetFullAlbum returns one Spotify album by ID, including artists and tracks.
func (s *Service) GetFullAlbum(ctx contextx.ContextX, userId, spotifyId string) (*spotify.FullAlbum, error) {
    client, err := s.Client(ctx, userId)
    if err != nil {
        return nil, fmt.Errorf("failed to get spotify client: %w", err)
    }
    album, err := client.GetAlbum(ctx, spotify.ID(spotifyId))
    if err != nil {
        return nil, fmt.Errorf("failed to get spotify album: %w", err)
    }
    return album, nil
}
```

- [ ] **Step 2: Add the conversion helper**

In `src/internal/library/service.go`, just under the existing imports add (or extend an existing block if convenient):

```go
// spotifyAlbumToDTO converts a Spotify FullAlbum into an AlbumDTO ready for
// EnsureAlbumWithMetadata. Mirrors the inline conversion used by the feed
// sync (see feed/service.go), but lives here so the radar-add flow can share it.
func spotifyAlbumToDTO(album *spotify.FullAlbum) AlbumDTO {
    var imageURL string
    if len(album.Images) > 0 {
        imageURL = album.Images[0].URL
    }
    dto := AlbumDTO{
        ID:        uuid.NewString(),
        SpotifyID: album.ID.String(),
        Title:     album.Name,
        ImageURL:  imageURL,
    }
    dto.Artists = make([]ArtistDTO, len(album.Artists))
    for i, a := range album.Artists {
        dto.Artists[i] = ArtistDTO{
            ID:        uuid.NewString(),
            SpotifyID: a.ID.String(),
            Name:      a.Name,
        }
    }
    dto.Tracks = make([]TrackDTO, 0, len(album.Tracks.Tracks))
    for _, t := range album.Tracks.Tracks {
        dto.Tracks = append(dto.Tracks, TrackDTO{
            ID:        uuid.NewString(),
            SpotifyID: t.ID.String(),
            Title:     t.Name,
        })
    }
    return dto
}
```

You will need to add the imports `"github.com/google/uuid"` and `spotify "github.com/zmb3/spotify/v2"` if not already present. Check the existing imports first.

- [ ] **Step 3: Add the service method**

Append to `service.go`:

```go
// AddSpotifyAlbumToRadar imports a Spotify album's metadata (album, artists,
// tracks) into wax and adds the album to the user's radar. Refuses with
// ErrAlbumAlreadyDecided if the album already has any user_releases row.
func (s *Service) AddSpotifyAlbumToRadar(ctx contextx.ContextX, userID, spotifyID string) error {
    spotifyAlbum, err := s.spotifyService.GetFullAlbum(ctx, userID, spotifyID)
    if err != nil {
        return fmt.Errorf("failed to fetch spotify album: %w", err)
    }
    dto := spotifyAlbumToDTO(spotifyAlbum)

    return s.db.WithTx(func(tx *db.DB) error {
        txRepo := NewRepo(tx.Queries())
        imported, err := txRepo.EnsureAlbumWithMetadata(ctx, dto)
        if err != nil {
            return fmt.Errorf("failed to import album metadata: %w", err)
        }
        hasRelease, err := txRepo.HasAnyUserReleaseForAlbum(ctx, userID, imported.ID)
        if err != nil {
            return fmt.Errorf("failed to check user releases: %w", err)
        }
        if hasRelease {
            return ErrAlbumAlreadyDecided
        }
        if err := txRepo.AddAlbumToRadar(ctx, userID, imported.ID); err != nil {
            return fmt.Errorf("failed to add album to radar: %w", err)
        }
        return nil
    })
}
```

- [ ] **Step 4: Build**

Run: `task build`
Expected: success.

- [ ] **Step 5: Commit**

```bash
git add src/internal/spotify/service.go src/internal/library/service.go
git commit -m "feat(library): AddSpotifyAlbumToRadar"
```

---

## Task 9: Service — `SearchAlbumsForDiscover` (with unit test)

**Files:**
- Modify: `src/internal/library/service.go`
- Modify: `src/internal/library/service_test.go`

`SearchAlbumsForDiscover` calls `spotify.Service.SearchAlbums`, then enriches each result with the user's wax state via `Repo.GetUserAlbumStateBySpotifyIDs`. The merge logic is pure once the dependencies return their data — we extract it as `mergeDiscoverState` and unit-test it directly.

- [ ] **Step 1: Add the pure merge helper**

Append to `service.go`:

```go
// mergeDiscoverState combines a slice of Spotify search results with the
// caller's per-album state lookup, producing one DiscoverResultDTO per
// Spotify result.
func mergeDiscoverState(results []spotify.SimpleAlbum, states map[string]UserAlbumStateRow) []DiscoverResultDTO {
    out := make([]DiscoverResultDTO, len(results))
    for i, a := range results {
        var imageURL string
        if len(a.Images) > 0 {
            imageURL = a.Images[0].URL
        }
        artists := make([]ArtistDTO, len(a.Artists))
        for j, ar := range a.Artists {
            artists[j] = ArtistDTO{
                SpotifyID: ar.ID.String(),
                Name:      ar.Name,
            }
        }
        dto := DiscoverResultDTO{
            SpotifyID: a.ID.String(),
            Title:     a.Name,
            Artists:   artists,
            ImageURL:  imageURL,
            State:     DiscoverAlbumStateNone,
        }
        if row, ok := states[a.ID.String()]; ok {
            dto.State = row.State
            dto.AlbumID = row.AlbumID
        }
        out[i] = dto
    }
    return out
}
```

- [ ] **Step 2: Add the service method**

Append to `service.go`:

```go
// SearchAlbumsForDiscover queries Spotify and enriches each result with the
// caller's wax state (in_library, on_radar, removed, or none). Returns an
// empty slice (not nil) when the query is empty or yields no hits.
func (s *Service) SearchAlbumsForDiscover(ctx contextx.ContextX, userID, query string, limit int) ([]DiscoverResultDTO, error) {
    results, err := s.spotifyService.SearchAlbums(ctx, userID, query, limit)
    if err != nil {
        return nil, fmt.Errorf("spotify search failed: %w", err)
    }
    if len(results) == 0 {
        return []DiscoverResultDTO{}, nil
    }
    spotifyIDs := make([]string, len(results))
    for i, a := range results {
        spotifyIDs[i] = a.ID.String()
    }
    states, err := s.repo.GetUserAlbumStateBySpotifyIDs(ctx, userID, spotifyIDs)
    if err != nil {
        return nil, fmt.Errorf("failed to look up album states: %w", err)
    }
    return mergeDiscoverState(results, states), nil
}
```

- [ ] **Step 3: Write the failing test**

Append to `src/internal/library/service_test.go`:

```go
func TestMergeDiscoverState(t *testing.T) {
    t.Run("marks unknown albums as 'none' with empty AlbumID", func(t *testing.T) {
        results := []spotify.SimpleAlbum{
            simpleAlbumStub("sp-1", "Unknown One", "art-a", "Artist A"),
        }
        out := mergeDiscoverState(results, map[string]UserAlbumStateRow{})
        if len(out) != 1 {
            t.Fatalf("got %d results, want 1", len(out))
        }
        if out[0].State != DiscoverAlbumStateNone {
            t.Errorf("state = %q, want %q", out[0].State, DiscoverAlbumStateNone)
        }
        if out[0].AlbumID != "" {
            t.Errorf("AlbumID = %q, want empty", out[0].AlbumID)
        }
    })

    t.Run("attaches in_library state and AlbumID when known", func(t *testing.T) {
        results := []spotify.SimpleAlbum{
            simpleAlbumStub("sp-1", "Known", "art-a", "Artist A"),
        }
        states := map[string]UserAlbumStateRow{
            "sp-1": {AlbumID: "wax-1", State: DiscoverAlbumStateInLibrary},
        }
        out := mergeDiscoverState(results, states)
        if out[0].State != DiscoverAlbumStateInLibrary {
            t.Errorf("state = %q, want in_library", out[0].State)
        }
        if out[0].AlbumID != "wax-1" {
            t.Errorf("AlbumID = %q, want wax-1", out[0].AlbumID)
        }
    })

    t.Run("preserves order and per-album metadata", func(t *testing.T) {
        results := []spotify.SimpleAlbum{
            simpleAlbumStub("sp-1", "First", "art-a", "Artist A"),
            simpleAlbumStub("sp-2", "Second", "art-b", "Artist B"),
        }
        states := map[string]UserAlbumStateRow{
            "sp-2": {AlbumID: "wax-2", State: DiscoverAlbumStateOnRadar},
        }
        out := mergeDiscoverState(results, states)
        if out[0].SpotifyID != "sp-1" || out[1].SpotifyID != "sp-2" {
            t.Fatalf("order broken: %v", out)
        }
        if out[0].State != DiscoverAlbumStateNone {
            t.Errorf("first state = %q, want none", out[0].State)
        }
        if out[1].State != DiscoverAlbumStateOnRadar || out[1].AlbumID != "wax-2" {
            t.Errorf("second result = %+v, want on_radar/wax-2", out[1])
        }
        if len(out[0].Artists) != 1 || out[0].Artists[0].Name != "Artist A" {
            t.Errorf("artist not propagated: %+v", out[0].Artists)
        }
    })
}

// simpleAlbumStub builds a minimal spotify.SimpleAlbum for tests.
func simpleAlbumStub(id, name, artistID, artistName string) spotify.SimpleAlbum {
    return spotify.SimpleAlbum{
        ID:   spotify.ID(id),
        Name: name,
        Artists: []spotify.SimpleArtist{
            {ID: spotify.ID(artistID), Name: artistName},
        },
    }
}
```

You will need to import `spotify "github.com/zmb3/spotify/v2"` in the test file if not already imported.

- [ ] **Step 4: Run the tests, verify pass**

Run: `task test/unit`
Expected: all pass, including the three `TestMergeDiscoverState` subtests.

- [ ] **Step 5: Commit**

```bash
git add src/internal/library/service.go src/internal/library/service_test.go
git commit -m "feat(library): SearchAlbumsForDiscover with state merge"
```

---

## Task 10: Service — `PromoteRadarToLibrary`

**Files:**
- Modify: `src/internal/library/service.go`

The album is on radar (no `user_releases` row, no `releases` row). Promotion creates a digital release row and a `user_releases` row in `'owned'` state, deletes the radar row, and pushes to Spotify saved library.

- [ ] **Step 1: Append the method**

Append to `service.go`:

```go
// PromoteRadarToLibrary transitions a radar album to an owned digital release
// and pushes the album to the user's Spotify saved library. Spotify push is
// best-effort; a failure is logged but does not roll back the local DB
// (mirrors RemoveAlbumFromLibrary).
func (s *Service) PromoteRadarToLibrary(ctx contextx.ContextX, userID, albumID string) error {
    spotifyID, err := s.repo.GetAlbumSpotifyID(ctx, albumID)
    if err != nil {
        return fmt.Errorf("failed to get album spotify id: %w", err)
    }

    err = s.db.WithTx(func(tx *db.DB) error {
        txRepo := NewRepo(tx.Queries())
        if _, err := txRepo.AddOwnedRelease(ctx, userID, albumID, models.ReleaseFormatDigital, "", time.Now()); err != nil {
            return fmt.Errorf("failed to add owned digital release: %w", err)
        }
        if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
            return fmt.Errorf("failed to clear radar: %w", err)
        }
        return nil
    })
    if err != nil {
        return err
    }

    if err := s.spotifyService.AddAlbumToSavedLibrary(ctx, userID, spotifyID); err != nil {
        slog.WarnContext(ctx, "failed to push album to spotify saved library after radar promotion", "error", err, "album_id", albumID, "spotify_id", spotifyID)
    }
    return nil
}
```

You will need `"log/slog"` and `"github.com/alecdray/wax/src/internal/core/db/models"` in the imports if not already present.

- [ ] **Step 2: Build**

Run: `task build`
Expected: success.

- [ ] **Step 3: Run unit tests**

Run: `task test`
Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add src/internal/library/service.go
git commit -m "feat(library): PromoteRadarToLibrary"
```

---

## Task 11: Compass icon

**Files:**
- Modify: `src/internal/core/templates/icons.templ`

Adds a `CompassIcon` matching the existing icon pattern (Bootstrap icons SVG, fill or outline by `IconStyle`).

- [ ] **Step 1: Append the icon**

Append to `icons.templ`:

```templ
templ CompassIcon(props IconProps) {
  if props.Style == IconStyleFill {
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-compass-fill" viewBox="0 0 16 16">
      <path d="M15.5 8.516a7.5 7.5 0 1 1-9.462-7.24A1 1 0 0 1 7 0h2a1 1 0 0 1 .962.276 7.5 7.5 0 0 1 5.538 8.24m-3.61-3.905L6.94 7.439 4.11 12.39l4.95-2.828 2.828-4.95z"></path>
    </svg>
  } else {
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-compass" viewBox="0 0 16 16">
      <path d="M8 16.016a7.5 7.5 0 0 0 1.962-14.74A1 1 0 0 0 9 0H7a1 1 0 0 0-.962 1.276A7.5 7.5 0 0 0 8 16.016m6.5-7.5a6.5 6.5 0 1 1-13 0 6.5 6.5 0 0 1 13 0"></path>
      <path d="m6.94 7.44 4.95-2.83-2.83 4.95-4.949 2.83z"></path>
    </svg>
  }
}
```

- [ ] **Step 2: Generate templ**

Run: `task build/templ`
Expected: success. `src/internal/core/templates/icons_templ.go` regenerated.

- [ ] **Step 3: Build**

Run: `task build`
Expected: success.

- [ ] **Step 4: Commit**

```bash
git add src/internal/core/templates/icons.templ src/internal/core/templates/icons_templ.go
git commit -m "feat(templates): CompassIcon"
```

---

## Task 12: Shared `LibraryHeaderBar`

**Files:**
- Modify: `src/internal/library/adapters/dashboard.templ`
- Modify: `src/internal/library/adapters/album_detail.templ`

Replace `DashboardHeaderBar` and `AlbumDetailHeaderBar` with one `LibraryHeaderBar`. The new component takes a small props struct that controls which icon highlights and whether the feeds dropdown is shown.

- [ ] **Step 1: Add `LibraryHeaderBar` to `dashboard.templ`**

In `dashboard.templ`, replace the existing `DashboardHeaderBar` (around line 917) with:

```templ
type HeaderBarProps struct {
    Active    string // "library" | "discover"
    ShowFeeds bool
    Feeds     []feed.FeedDTO
}

templ LibraryHeaderBar(props HeaderBarProps) {
    <div class="bg-base-100 border-b border-base-300 h-11 w-full flex-shrink-0 sticky top-0 z-10">
        <div class="h-full flex items-center justify-between px-6">
            <div class="flex items-center gap-4">
                <a href="/" class="text-lg text-primary font-brand">wax</a>
                <div class="h-4 w-px bg-base-300"></div>
                <div class="flex items-center gap-1">
                    <div class="tooltip tooltip-bottom" data-tip="Library">
                        if props.Active == "library" {
                            <div class="btn btn-ghost btn-xs btn-square opacity-30 pointer-events-none" data-testid="header-library-icon-active">
                                @templates.CollectionIcon(templates.IconProps{Style: templates.IconStyleFill})
                            </div>
                        } else {
                            <a href="/app/library/dashboard" class="btn btn-ghost btn-xs btn-square" data-testid="header-library-icon">
                                @templates.CollectionIcon(templates.IconProps{Style: templates.IconStyleOutline})
                            </a>
                        }
                    </div>
                    <div class="tooltip tooltip-bottom" data-tip="Discover">
                        if props.Active == "discover" {
                            <div class="btn btn-ghost btn-xs btn-square opacity-30 pointer-events-none" data-testid="header-discover-icon-active">
                                @templates.CompassIcon(templates.IconProps{Style: templates.IconStyleFill})
                            </div>
                        } else {
                            <a href="/app/library/discover" class="btn btn-ghost btn-xs btn-square" data-testid="header-discover-icon">
                                @templates.CompassIcon(templates.IconProps{Style: templates.IconStyleOutline})
                            </a>
                        }
                    </div>
                </div>
            </div>
            <div class="flex items-center gap-2">
                if props.ShowFeeds {
                    @feedsDropdown(props.Feeds)
                    <div class="h-4 w-px bg-base-300"></div>
                }
                <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs btn-circle">
                        @templates.UserIcon(templates.IconProps{Style: templates.IconStyleOutline})
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu menu-compact bg-base-100 rounded-box w-32 shadow-xl border border-base-300 mt-1">
                        <li><a href="/logout" class="text-xs">Logout</a></li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
}
```

Then delete the old `DashboardHeaderBar` templ block from the same file.

- [ ] **Step 2: Update `DashboardPage` to use the new component**

In `dashboard.templ`, find the `DashboardPage` templ (around line 1071) and change its body's first call from:

```templ
@DashboardHeaderBar(props.Feeds)
```

to:

```templ
@LibraryHeaderBar(HeaderBarProps{Active: "library", ShowFeeds: true, Feeds: props.Feeds})
```

- [ ] **Step 3: Update `album_detail.templ`**

In `album_detail.templ`, delete the `AlbumDetailHeaderBar` templ block and replace its single call site (in `AlbumDetailPage`, around line 42) from:

```templ
@AlbumDetailHeaderBar()
```

to:

```templ
@LibraryHeaderBar(HeaderBarProps{Active: "library"})
```

(`ShowFeeds` defaults to `false`; the album detail header doesn't show feeds today.)

- [ ] **Step 4: Generate templ and build**

Run: `task build/templ && task build`
Expected: success.

- [ ] **Step 5: Run unit tests**

Run: `task test`
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add src/internal/library/adapters/dashboard.templ src/internal/library/adapters/dashboard_templ.go src/internal/library/adapters/album_detail.templ src/internal/library/adapters/album_detail_templ.go
git commit -m "refactor(library): one LibraryHeaderBar replaces dashboard and album-detail variants"
```

---

## Task 13: Discover templates

**Files:**
- Create: `src/internal/library/adapters/discover.templ`

Creates the page, the radar carousel, the search results list, and the per-row component. All reuse existing card/strip patterns from `dashboard.templ` so the visual language stays consistent.

- [ ] **Step 1: Create the file**

Write `src/internal/library/adapters/discover.templ`:

```templ
package adapters

import (
    "fmt"
    "github.com/alecdray/wax/src/internal/core/templates"
    "github.com/alecdray/wax/src/internal/library"
)

type DiscoverPageProps struct {
    RadarAlbums  []library.AlbumDTO
    Query        string
    SearchResults []library.DiscoverResultDTO
}

templ DiscoverPage(props DiscoverPageProps) {
    @templates.RootComponent(templates.RootProps{
        Title: templates.CreatePageTitle("Discover"),
    }) {
        <div class="w-full flex flex-col" hx-history="false">
            @LibraryHeaderBar(HeaderBarProps{Active: "discover"})
            <div class="flex flex-col py-4 gap-6 items-center max-w-3xl mx-auto w-full px-4">
                <div class="w-full flex flex-col gap-1">
                    <span class="text-xs font-semibold uppercase tracking-widest text-base-content/40 px-1">On Radar</span>
                    @RadarCarousel(props.RadarAlbums, false)
                </div>
                <div class="w-full flex flex-col gap-3">
                    <input
                        id="discover-search-input"
                        name="q"
                        type="search"
                        class="input input-bordered w-full"
                        placeholder="Search Spotify for albums..."
                        value={ props.Query }
                        autocomplete="off"
                        hx-get="/app/library/discover/search"
                        hx-trigger="keyup changed delay:300ms, search"
                        hx-target="#discover-results"
                        hx-swap="innerHTML"
                        data-testid="discover-search-input"
                    />
                    <div id="discover-results" data-testid="discover-results">
                        @DiscoverSearchResults(props.SearchResults, props.Query)
                    </div>
                </div>
            </div>
        </div>
    }
}

templ RadarCarousel(albums []library.AlbumDTO, isOobSwap bool) {
    <div
        id="radar-carousel"
        class="w-full flex-shrink-0"
        hx-get="/app/library/discover/radar"
        hx-trigger="radarUpdated from:body"
        hx-swap="outerHTML"
        if isOobSwap {
            hx-swap-oob="true"
        }
        data-testid="radar-carousel"
    >
        if len(albums) == 0 {
            <div class="px-4 py-4 text-xs text-base-content/40" data-testid="radar-empty">
                No albums on your radar yet — search below to find some.
            </div>
        } else {
            <div class="carousel carousel-end gap-3 py-2 w-full overscroll-x-none">
                for _, album := range albums {
                    <div class="carousel-item" data-testid="radar-carousel-item" data-album-id={ album.ID }>
                        @radarCarouselCard(album)
                    </div>
                }
            </div>
        }
    </div>
}

templ radarCarouselCard(album library.AlbumDTO) {
    <div class="flex flex-col items-center gap-1 w-26 relative">
        <a
            href={ templ.URL(fmt.Sprintf("https://open.spotify.com/album/%s", album.SpotifyID)) }
            target="_blank"
            rel="noopener noreferrer"
            class="hover:opacity-80 transition-opacity"
            data-testid="radar-card-cover"
        >
            if album.ImageURL != "" {
                <div class="avatar">
                    <div class="mask mask-squircle h-24 w-24 flex-shrink-0">
                        <img src={ album.ImageURL } alt={ album.Title }/>
                    </div>
                </div>
            } else {
                <div class="avatar avatar-placeholder">
                    <div class="mask mask-squircle h-24 w-24 flex-shrink-0 bg-base-300"></div>
                </div>
            }
        </a>
        <span class="text-xs text-nowrap truncate w-full text-left">{ album.Title }</span>
        if len(album.Artists) > 0 {
            <span class="text-xs text-nowrap truncate w-full text-left text-base-content/40">
                for i, artist := range album.Artists {
                    if i > 0 {
                        { ", " }
                    }
                    { artist.Name }
                }
            </span>
        }
        <div class="dropdown dropdown-end absolute top-1 right-1">
            <div tabindex="0" role="button" class="btn btn-ghost btn-xs btn-circle bg-base-100/80" data-testid="radar-card-menu">
                @templates.EllipsisVerticalIcon(templates.IconProps{})
            </div>
            <ul tabindex="0" class="dropdown-content z-[1] menu menu-compact bg-base-100 rounded-box w-44 shadow-xl border border-base-300 mt-1">
                <li>
                    <button
                        class="text-xs"
                        hx-post={ fmt.Sprintf("/app/library/albums/%s/library", album.ID) }
                        hx-swap="none"
                        data-testid="radar-card-add-to-library"
                    >Add to library</button>
                </li>
                <li>
                    <button
                        class="text-xs text-error"
                        hx-delete={ fmt.Sprintf("/app/library/albums/%s/radar", album.ID) }
                        hx-swap="none"
                        data-testid="radar-card-remove"
                    >Remove from radar</button>
                </li>
            </ul>
        </div>
    </div>
}

templ DiscoverSearchResults(results []library.DiscoverResultDTO, query string) {
    if query == "" {
        <div class="text-sm text-base-content/40 px-1 py-2" data-testid="discover-results-empty">
            Start typing to search Spotify.
        </div>
    } else if len(results) == 0 {
        <div class="text-sm text-base-content/40 px-1 py-2" data-testid="discover-results-no-hits">
            No results.
        </div>
    } else {
        <ul class="list flex flex-col gap-1">
            for _, result := range results {
                @discoverSearchResultRow(result)
            }
        </ul>
    }
}

templ discoverSearchResultRow(result library.DiscoverResultDTO) {
    <li
        class="list-row items-center gap-3 px-2 py-2 rounded"
        data-result
        data-spotify-id={ result.SpotifyID }
        data-testid="discover-result-row"
    >
        if result.ImageURL != "" {
            <div class="avatar flex-shrink-0">
                <div class="mask mask-squircle h-12 w-12">
                    <img src={ result.ImageURL } alt={ result.Title }/>
                </div>
            </div>
        } else {
            <div class="avatar avatar-placeholder flex-shrink-0">
                <div class="mask mask-squircle h-12 w-12 bg-base-300"></div>
            </div>
        }
        <div class="flex-1 min-w-0">
            <div class="text-sm font-medium truncate" data-testid="discover-result-title">{ result.Title }</div>
            if len(result.Artists) > 0 {
                <div class="text-xs text-base-content/60 truncate">
                    for i, a := range result.Artists {
                        if i > 0 {
                            { ", " }
                        }
                        { a.Name }
                    }
                </div>
            }
        </div>
        <div class="flex-shrink-0">
            @discoverResultAffordance(result)
        </div>
    </li>
}

templ discoverResultAffordance(result library.DiscoverResultDTO) {
    switch result.State {
        case library.DiscoverAlbumStateInLibrary:
            <a
                href={ templ.URL(fmt.Sprintf("/app/library/albums/%s", result.AlbumID)) }
                class="badge badge-soft badge-primary"
                data-testid="discover-result-in-library"
            >In library</a>
        case library.DiscoverAlbumStateOnRadar:
            <button
                class="btn btn-xs btn-ghost text-error"
                hx-delete={ fmt.Sprintf("/app/library/albums/%s/radar", result.AlbumID) }
                hx-target="closest [data-result]"
                hx-swap="outerHTML"
                data-testid="discover-result-remove-radar"
            >On radar — remove</button>
        case library.DiscoverAlbumStateRemoved:
            <button
                class="btn btn-xs btn-primary"
                hx-post={ fmt.Sprintf("/app/library/albums/%s/library", result.AlbumID) }
                hx-target="closest [data-result]"
                hx-swap="outerHTML"
                data-testid="discover-result-reacquire"
            >Re-acquire</button>
        default:
            <button
                class="btn btn-xs btn-primary"
                hx-post={ fmt.Sprintf("/app/library/discover/radar?spotifyId=%s", result.SpotifyID) }
                hx-target="closest [data-result]"
                hx-swap="outerHTML"
                data-testid="discover-result-add-radar"
            >+ Add to radar</button>
    }
}
```

- [ ] **Step 2: Generate templ**

Run: `task build/templ`
Expected: success. `src/internal/library/adapters/discover_templ.go` is created.

- [ ] **Step 3: Build**

Run: `task build`
Expected: success — but it WILL fail until the handlers in Task 14 are wired. If it fails only because the handler symbols don't exist yet, that's expected; proceed.

- [ ] **Step 4: Commit**

```bash
git add src/internal/library/adapters/discover.templ src/internal/library/adapters/discover_templ.go
git commit -m "feat(library): discover.templ with radar carousel and search results"
```

---

## Task 14: HTTP handlers + routes

**Files:**
- Modify: `src/internal/library/adapters/http.go`
- Modify: `src/internal/library/adapters/routes.go`

Six new handlers. Each fires HTMX events through the `HX-Trigger` header where appropriate to keep the radar carousel and search rows in sync.

- [ ] **Step 1: Add handlers to `http.go`**

Append to `http.go`:

```go
// --- Discover page ---

func (h *HttpHandler) GetDiscoverPage(w http.ResponseWriter, r *http.Request) {
    ctx := contextx.NewContextX(r.Context())
    userId, err := ctx.UserId()
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusBadRequest,
            Err:    fmt.Errorf("failed to get user ID: %w", err),
        })
        return
    }
    radar, err := h.libraryService.GetRadarAlbums(ctx, userId)
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to get radar albums: %w", err),
        })
        return
    }
    DiscoverPage(DiscoverPageProps{
        RadarAlbums:   radar,
        Query:         "",
        SearchResults: nil,
    }).Render(r.Context(), w)
}

func (h *HttpHandler) GetDiscoverRadar(w http.ResponseWriter, r *http.Request) {
    ctx := contextx.NewContextX(r.Context())
    userId, err := ctx.UserId()
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusBadRequest,
            Err:    fmt.Errorf("failed to get user ID: %w", err),
        })
        return
    }
    radar, err := h.libraryService.GetRadarAlbums(ctx, userId)
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to get radar albums: %w", err),
        })
        return
    }
    RadarCarousel(radar, false).Render(r.Context(), w)
}

func (h *HttpHandler) GetDiscoverSearch(w http.ResponseWriter, r *http.Request) {
    ctx := contextx.NewContextX(r.Context())
    userId, err := ctx.UserId()
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusBadRequest,
            Err:    fmt.Errorf("failed to get user ID: %w", err),
        })
        return
    }
    query := strings.TrimSpace(r.URL.Query().Get("q"))
    if query == "" {
        DiscoverSearchResults(nil, "").Render(r.Context(), w)
        return
    }
    results, err := h.libraryService.SearchAlbumsForDiscover(ctx, userId, query, 20)
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("spotify search failed: %w", err),
        })
        return
    }
    DiscoverSearchResults(results, query).Render(r.Context(), w)
}

func (h *HttpHandler) PostDiscoverRadar(w http.ResponseWriter, r *http.Request) {
    ctx := contextx.NewContextX(r.Context())
    userId, err := ctx.UserId()
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusBadRequest,
            Err:    fmt.Errorf("failed to get user ID: %w", err),
        })
        return
    }
    spotifyID := r.URL.Query().Get("spotifyId")
    if spotifyID == "" {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusBadRequest,
            Err:    fmt.Errorf("missing spotifyId"),
        })
        return
    }
    if err := h.libraryService.AddSpotifyAlbumToRadar(ctx, userId, spotifyID); err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to add album to radar: %w", err),
        })
        return
    }
    // Look up the row's new state so we can re-render the row in place.
    states, err := h.libraryService.LookupDiscoverState(ctx, userId, []string{spotifyID})
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to look up new state: %w", err),
        })
        return
    }
    row := library.DiscoverResultDTO{SpotifyID: spotifyID, State: library.DiscoverAlbumStateOnRadar}
    if s, ok := states[spotifyID]; ok {
        row.AlbumID = s.AlbumID
        row.State = s.State
    }
    w.Header().Set("HX-Trigger", "radarUpdated")
    discoverSearchResultRow(row).Render(r.Context(), w)
}

func (h *HttpHandler) DeleteAlbumRadar(w http.ResponseWriter, r *http.Request) {
    ctx := contextx.NewContextX(r.Context())
    userId, err := ctx.UserId()
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusBadRequest,
            Err:    fmt.Errorf("failed to get user ID: %w", err),
        })
        return
    }
    albumID := r.PathValue("albumId")
    // We need the Spotify ID before we delete the row (so the search-results
    // re-render below can carry it). The album row itself is preserved by the
    // delete; only the user_album_radar row goes away.
    spotifyID, err := h.libraryService.GetAlbumSpotifyID(ctx, albumID)
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to look up album: %w", err),
        })
        return
    }
    if err := h.libraryService.RemoveAlbumFromRadar(ctx, userId, albumID); err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to remove from radar: %w", err),
        })
        return
    }
    w.Header().Set("HX-Trigger", "radarUpdated")
    // Re-render the row in the "none" state. The carousel-card caller uses
    // hx-swap="none" so it ignores this body; the search-results-row caller
    // uses hx-swap="outerHTML" so the row updates in place.
    discoverSearchResultRow(library.DiscoverResultDTO{
        SpotifyID: spotifyID,
        State:     library.DiscoverAlbumStateNone,
    }).Render(r.Context(), w)
}

func (h *HttpHandler) PostAlbumLibrary(w http.ResponseWriter, r *http.Request) {
    ctx := contextx.NewContextX(r.Context())
    userId, err := ctx.UserId()
    if err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusBadRequest,
            Err:    fmt.Errorf("failed to get user ID: %w", err),
        })
        return
    }
    albumID := r.PathValue("albumId")
    if err := h.libraryService.PromoteRadarToLibrary(ctx, userId, albumID); err != nil {
        httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to promote to library: %w", err),
        })
        return
    }
    w.Header().Set("HX-Trigger", "radarUpdated, libraryUpdated")
    w.WriteHeader(http.StatusOK)
}
```

You will need `"strings"` in the imports.

- [ ] **Step 2: Add two service helpers**

Append to `src/internal/library/service.go`:

```go
// LookupDiscoverState exposes the per-Spotify-ID state lookup for adapters
// that need it after a write (e.g., to re-render a row in its new state).
func (s *Service) LookupDiscoverState(ctx contextx.ContextX, userID string, spotifyIDs []string) (map[string]UserAlbumStateRow, error) {
    return s.repo.GetUserAlbumStateBySpotifyIDs(ctx, userID, spotifyIDs)
}

// GetAlbumSpotifyID returns the Spotify ID for a wax album. Thin wrapper over
// the repo method (already used internally by RemoveAlbumFromLibrary); now
// exposed so adapters can re-render search-result rows after a radar delete.
func (s *Service) GetAlbumSpotifyID(ctx contextx.ContextX, albumID string) (string, error) {
    return s.repo.GetAlbumSpotifyID(ctx, albumID)
}
```

- [ ] **Step 3: Register the routes**

Replace the body of `RegisterRoutes` in `routes.go` to append the new routes:

```go
package adapters

import (
    "github.com/alecdray/wax/src/internal/core/httpx"
)

func RegisterRoutes(mux *httpx.Mux, h *HttpHandler) {
    mux.Handle("/app/library/dashboard", httpx.HandlerFunc(h.GetDashboardPage))
    mux.Handle("/app/library/dashboard/feeds-dropdown-content", httpx.HandlerFunc(h.GetFeedsDropdown))
    mux.Handle("GET /app/library/dashboard/stats", httpx.HandlerFunc(h.GetLibraryStats))
    mux.Handle("POST /app/library/dashboard/feeds/sync", httpx.HandlerFunc(h.TriggerFeedSync))
    mux.Handle("/app/library/dashboard/albums-table", httpx.HandlerFunc(h.GetAlbumsTable))
    mux.Handle("GET /app/library/dashboard/albums-page", httpx.HandlerFunc(h.GetAlbumsPage))
    mux.Handle("GET /app/library/dashboard/carousel", httpx.HandlerFunc(h.GetCarousel))

    mux.Handle("GET /app/library/discover", httpx.HandlerFunc(h.GetDiscoverPage))
    mux.Handle("GET /app/library/discover/search", httpx.HandlerFunc(h.GetDiscoverSearch))
    mux.Handle("GET /app/library/discover/radar", httpx.HandlerFunc(h.GetDiscoverRadar))
    mux.Handle("POST /app/library/discover/radar", httpx.HandlerFunc(h.PostDiscoverRadar))

    mux.Handle("GET /app/library/albums/{albumId}", httpx.HandlerFunc(h.GetAlbumDetailPage))
    mux.Handle("DELETE /app/library/albums/{albumId}", httpx.HandlerFunc(h.DeleteAlbum))
    mux.Handle("DELETE /app/library/albums/{albumId}/radar", httpx.HandlerFunc(h.DeleteAlbumRadar))
    mux.Handle("POST /app/library/albums/{albumId}/library", httpx.HandlerFunc(h.PostAlbumLibrary))
    mux.Handle("GET /app/library/albums/{albumId}/formats", httpx.HandlerFunc(h.GetFormatsModal))
    mux.Handle("PUT /app/library/albums/{albumId}/formats", httpx.HandlerFunc(h.PutFormats))
    mux.Handle("GET /app/library/albums/{albumId}/formats/{format}/discogs/search", httpx.HandlerFunc(h.GetDiscogsSearch))
    mux.Handle("GET /app/library/albums/{albumId}/formats/{format}/discogs/releases/{discogsId}", httpx.HandlerFunc(h.GetDiscogsRelease))
    mux.Handle("GET /app/library/albums/{albumId}/sleeve-notes/editor", httpx.HandlerFunc(h.GetSleeveNotesEditor))
    mux.Handle("GET /app/library/albums/{albumId}/sleeve-notes/view", httpx.HandlerFunc(h.GetSleeveNotesView))
    mux.Handle("PUT /app/library/albums/{albumId}/sleeve-notes", httpx.HandlerFunc(h.SaveSleeveNote))
}
```

- [ ] **Step 4: Build**

Run: `task build`
Expected: success.

- [ ] **Step 5: Run tests**

Run: `task test`
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add src/internal/library/adapters/http.go src/internal/library/adapters/routes.go src/internal/library/service.go
git commit -m "feat(library): discover handlers and routes"
```

---

## Task 15: Manual verification

End-to-end smoke check. The plan deliberately defers Playwright e2e (matching the album-states plan's convention). Skip this step if the worktree's `.env`/`db.sql` have already been provisioned.

- [ ] **Step 1: Provision the worktree**

```bash
cp /Users/shmoopy/workshop/projects/wax/.env .env
cp /Users/shmoopy/workshop/projects/wax/tmp/db.sql ./tmp/db.sql 2>/dev/null || mkdir -p tmp && cp /Users/shmoopy/workshop/projects/wax/tmp/db.sql ./tmp/db.sql
npm install
```

- [ ] **Step 2: Boot the dev server**

Run: `task dev`
Expected: server up at localhost (whatever port `.env` sets).

- [ ] **Step 3: Walk the golden path in a browser**

Log in, then:

1. From the dashboard, click the new Compass icon in the header bar. Confirm `/app/library/discover` loads.
2. Confirm the radar carousel shows the empty-state hint when there are no entries.
3. Type a query (e.g. `boards of canada`). Confirm:
   - Results appear after the 300ms debounce.
   - At least one result has `+ Add to radar`.
   - Albums you already own show `In library` linking to the album page.
4. Click `+ Add to radar` on a result. Confirm:
   - The result row updates in place to `On radar — remove`.
   - The radar carousel above refreshes and now shows the album.
5. Open the radar carousel card's 3-dot menu and click `Add to library`. Confirm:
   - The album leaves the radar carousel (or the empty-state returns).
   - Navigating to `/app/library/dashboard` shows the album in the list view.
   - Spotify (web/desktop) shows the album in your saved library (best-effort — if Spotify is unreachable, check the server logs for the WARN line; the local DB should still reflect the change).
6. Find a different album, add it to radar, then click `Remove from radar` from the carousel menu. Confirm the carousel returns to the empty state.

- [ ] **Step 4: Tail logs for unexpected errors**

Run: `tail -n 200 tmp/dev.log` (or wherever the dev server's stderr is captured).
Expected: no panics, no 500s. WARN lines from Spotify push failures are acceptable; ERROR lines are not.

(No commit — manual check only.)

---

## Self-review notes for the implementer

After all tasks: re-open `docs/specs/1778426298-radar-ui/design.md` and confirm:

1. Every section in the design's *Routes*, *Service layer*, and *Templates* lists is implemented.
2. The HTMX `radarUpdated` body event fires on every mutation (POST radar, DELETE radar, POST library) and the radar carousel refreshes accordingly.
3. The Spotify-push call in `PromoteRadarToLibrary` is wrapped in a `slog.Warn` and does NOT roll back the local DB on failure (matches `RemoveAlbumFromLibrary`).
4. The `ErrAlbumAlreadyDecided` path returns a structured error toast on `POST /app/library/discover/radar` rather than silently succeeding (the handler currently returns 500; that's acceptable for the v1 since the UI prevents this for already-known albums — file as follow-up if you want a cleaner toast).
5. `task build` and `task test` pass with zero warnings.

If anything is missing, file it as follow-up rather than retrofitting into this plan.
