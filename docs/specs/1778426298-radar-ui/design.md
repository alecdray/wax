# Radar UI/UX — design

Ship the user-facing surface for the radar feature whose data model and service layer landed in `1778379187-album-states`. Users discover albums via Spotify search, queue them on radar, then triage radar entries into the library.

The previous spec deferred all UI changes; this is the follow-up.

## Goals

- A `/discover` page where the user can:
  - See their current radar entries as a horizontal carousel.
  - Search Spotify for albums by free-text query.
  - Add a result to radar in one tap.
  - Promote a radar entry to the library (digital format) and have that push to the user's Spotify saved library.
  - Remove a radar entry.
- A discoverable nav entry on the dashboard header bar that links to `/discover`.
- Search results that reflect wax state at a glance — already in library, already on radar, or new.

## Non-goals

- Wishlist UI (release-level pre-acquisition). The backend exists; the surface is a separate spec.
- Format picker on radar→library promotion. This round is digital-only; users refine format later from album detail.
- Filter/sort affordances on the radar list.
- The roadmap's "Hidden Albums" feature.
- Any cleanup of legacy `deleted_at` on `user_releases`.

## Page surface

```
+------------------------------------------------------------+
| [wax] | [📚 Library] [🧭 Discover (active)]                |
+------------------------------------------------------------+
| Discover                                                   |
|                                                            |
| ON RADAR                                                   |
| [carousel: cover · title · artists · ⋮]  →                 |
|                                                            |
| [🔍 Search Spotify for albums...........................]  |
|                                                            |
| Results                                                    |
| ┌────────┬────────────────────────────────┬──────────────┐ |
| │ cover  │ title · artists                │ + Add to     │ |
| │        │                                │   radar      │ |
| └────────┴────────────────────────────────┴──────────────┘ |
| ┌────────┬────────────────────────────────┬──────────────┐ |
| │ cover  │ title · artists                │ ✓ In library │ |
| └────────┴────────────────────────────────┴──────────────┘ |
+------------------------------------------------------------+
```

The radar carousel uses the same horizontal-strip styling as the dashboard's `Recently Spun` carousel so the visual language is consistent. The compass icon (Discover) sits next to the existing Library icon in the dashboard header bar.

## Module placement

Everything lives in `src/internal/library/adapters/`. The library module's `CLAUDE.md` already declares: *Library owns the album view UI.* Discover is another view of albums (specifically: not-yet-decided ones), so it belongs here. No new module.

A new module would also lock us into duplicating album/artist/release rendering, since those types are owned by `library`.

## Routes

Added to `library/adapters/routes.go`:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/app/library/discover` | The page (full document) |
| `GET` | `/app/library/discover/search` | HTMX fragment of search results, query in `?q=...` |
| `GET` | `/app/library/discover/radar` | HTMX fragment of the radar carousel (refreshed via `radarUpdated` body event) |
| `POST` | `/app/library/discover/radar` | Add a Spotify album to radar; body/query carries `spotifyId` |
| `DELETE` | `/app/library/albums/{albumId}/radar` | Remove an album from radar |
| `POST` | `/app/library/albums/{albumId}/library` | Promote radar entry to owned digital release |

The `albumId`-keyed routes use the existing wax album ID once the album has been imported. Adding to radar from a Spotify search result uses `spotifyId` because the album may not yet have a wax row.

### Existing routes that need to change

- `GET /app/library/albums/{albumId}` — when an album is *only* on radar (no `user_releases` row), the album detail page already requires an "in library" row to render. Two options:
  - (chosen) Don't link radar carousel cards to the album detail page; tap navigates to the Spotify URL instead. Mirrors the existing `Recently Spun` carousel behaviour for not-yet-in-library albums.
  - (rejected) Loosen the album detail handler to render radar-only albums. Bigger change, more invariants to revisit, not required for radar to work.

## Service layer

### `library.Service` — additions

```go
// SearchAlbumsForDiscover queries Spotify and enriches each result with the
// caller's current wax state (in library, on radar, or none). Used to render
// the discover search results.
SearchAlbumsForDiscover(ctx, userID, query string, limit int) ([]DiscoverResultDTO, error)

// AddSpotifyAlbumToRadar imports the Spotify album (creating album/artist/track
// rows if needed) and inserts a user_album_radar row. Returns ErrAlbumAlreadyDecided
// if the album already has any user_releases row for the caller.
AddSpotifyAlbumToRadar(ctx, userID, spotifyID string) error

// GetRadarAlbums returns the caller's radar entries as fully-populated
// AlbumDTOs (so the carousel can render covers, titles, artists).
// Mirrors the naming of GetRecentlyPlayedAlbums / GetUnratedAlbums.
GetRadarAlbums(ctx, userID string) ([]AlbumDTO, error)

// RemoveAlbumFromRadar deletes the radar row. No-op if absent.
RemoveAlbumFromRadar(ctx, userID, albumID string) error

// PromoteRadarToLibrary transitions an album from radar to owned digital and
// pushes it to the user's Spotify saved library. Spotify push is best-effort,
// matching RemoveAlbumFromLibrary.
PromoteRadarToLibrary(ctx, userID, albumID string) error
```

Implementation notes:

- `SearchAlbumsForDiscover` calls `spotify.Service.SearchAlbums`, then issues a single repo lookup per page of results — `Repo.GetUserAlbumStateBySpotifyIDs` (new) — to fetch `(albumID, state)` for every Spotify ID in one query. State is a string enum with three values: `in_library`, `on_radar`, `none`.
- `AddSpotifyAlbumToRadar` reuses the existing Spotify-album-fetch path that the feed sync uses (`spotify.Service.GetAlbum` or equivalent) to obtain a populated `AlbumDTO`, then in `db.WithTx`: get-or-create album, get-or-create artists, get-or-create tracks, then `Service.AddAlbumToRadar` (which keeps the `HasAnyUserReleaseForAlbum` guard). **No `releases` or `user_releases` rows are written** — radar is pre-decision; pre-creating a digital release would bias the format choice. Implementation hint: `AddAlbumToCollection` in `repo.go` already creates album/artists/tracks; extract its tail (the release-creating loop) into a separate method so both flows can share the metadata-import portion.
- `PromoteRadarToLibrary` runs in a transaction: get-or-create the digital release row, insert a `user_releases` row with `status='owned'`, delete the radar row. After commit, calls `spotify.Service.AddAlbumToSavedLibrary`. A failure to push is logged at WARN; the next Spotify feed sync will reconcile.

### `library.Repo` — additions

```go
// GetUserAlbumStateBySpotifyIDs returns the wax state for each Spotify ID the
// caller cares about. Returns a map keyed by Spotify ID; missing keys mean
// "user has no row for this album" (state = none, no wax album ID).
GetUserAlbumStateBySpotifyIDs(ctx, userID string, spotifyIDs []string) (map[string]UserAlbumStateRow, error)

type UserAlbumStateRow struct {
    AlbumID string             // wax album ID (always populated when present in the map)
    State   DiscoverAlbumState // in_library | on_radar | removed
}
```

`removed` surfaces explicitly so the UI can render a `Re-acquire` affordance (see *Edge cases*). `on_radar` rows have an `AlbumID` because the album row exists in wax even before the user owns any release of it.

### `spotify.Service` — additions

```go
// SearchAlbums runs a Spotify catalog search restricted to albums.
SearchAlbums(ctx, userID, query string, limit int) ([]spotify.SimpleAlbum, error)

// AddAlbumToSavedLibrary adds the album to the user's saved library on Spotify.
AddAlbumToSavedLibrary(ctx, userID, spotifyID string) error
```

Both go through the per-user `*spotify.Client` built by `Service.Client(...)`. `AddAlbumToSavedLibrary` uses the SDK's `AddAlbumsToLibrary`. `SearchAlbums` uses `client.Search(ctx, query, spotify.SearchTypeAlbum)`. No raw HTTP needed — neither endpoint requires the workaround `RemoveAlbum` uses.

## Domain types

Added in `library/library.go`:

```go
type DiscoverAlbumState string

const (
    DiscoverAlbumStateNone      DiscoverAlbumState = "none"
    DiscoverAlbumStateInLibrary DiscoverAlbumState = "in_library"
    DiscoverAlbumStateOnRadar   DiscoverAlbumState = "on_radar"
    DiscoverAlbumStateRemoved   DiscoverAlbumState = "removed"
)

type DiscoverResultDTO struct {
    SpotifyID string
    Title     string
    Artists   []ArtistDTO
    ImageURL  string
    State     DiscoverAlbumState
    AlbumID   string // empty when State == "none"; populated for in_library / on_radar / removed
}
```

`DiscoverResultDTO` is intentionally smaller than `AlbumDTO` — search results don't need releases, ratings, tags, sleeve notes.

## Templates

New `library/adapters/discover.templ` with:

- `DiscoverPage(props DiscoverPageProps)` — full page wrapped by the shared header (`LibraryHeaderBar`, see *Navbar refactor*).
- `RadarCarousel(albums []library.AlbumDTO, isOobSwap bool)` — horizontal strip; reuses the visual style of `carouselStrip` in `dashboard.templ`. Each card has a 3-dot menu (DaisyUI dropdown) with `Remove from radar` and `Add to library`. Tapping the cover opens Spotify (mirrors how Recently Spun handles not-yet-in-library albums; radar entries by definition aren't yet in the library, so the existing album detail page can't render them).
- `DiscoverSearchResults(results []library.DiscoverResultDTO, query string)` — renders the result list. Each row is `discoverSearchResultRow`.
- `discoverSearchResultRow(result library.DiscoverResultDTO)` — cover, title, artists, and one of three trailing affordances based on `State`.

The header bar refactor is small but improves both pages — same shape as the existing `AlbumDetailHeaderBar` borrowing from `DashboardHeaderBar`. Captured as a separate plan task.

### HTMX wiring

- Search input: `hx-get="/app/library/discover/search"` `hx-trigger="keyup changed delay:300ms, search"` `hx-target="#discover-results"` `hx-swap="innerHTML"`. `name="q"` on the input means the query rides naturally.
- Add-to-radar button on a result row: `hx-post="/app/library/discover/radar?spotifyId=…"` `hx-target="closest [data-result]"` `hx-swap="outerHTML"`. The handler returns the same row re-rendered with `State=on_radar`.
- Remove-from-radar from a result row: `hx-delete="/app/library/albums/{albumId}/radar"` `hx-target="closest [data-result]"` `hx-swap="outerHTML"`. Handler returns the row re-rendered with `State=none`.
- Radar carousel actions: same endpoints, but the responses also fire a `radarUpdated` body event via `HX-Trigger`. The carousel listens with `hx-trigger="radarUpdated from:body"` and re-fetches `/app/library/discover/radar`.
- Add-to-library from radar: `hx-post="/app/library/albums/{albumId}/library"` `hx-target="closest [data-result]"` `hx-swap="outerHTML"`. Handler fires both `radarUpdated` (so the carousel refreshes) and `libraryUpdated` (so the dashboard refreshes when the user navigates back).

The `radarUpdated` body-event pattern matches how `libraryUpdated` already works for the dashboard's `AlbumsList` and `LibraryStats`.

## Navbar refactor

Today the dashboard header bar (`DashboardHeaderBar`) renders an inert "Library" tooltip-icon and the album detail page has its own `AlbumDetailHeaderBar` with an active back-link. We consolidate to one `LibraryHeaderBar(props HeaderBarProps)` used by all three pages (dashboard, album detail, discover):

```go
type HeaderBarProps struct {
    Active    string         // "library" | "discover" — which icon highlights
    ShowFeeds bool           // dashboard-only; album detail and discover hide it
    Feeds     []feed.FeedDTO // ignored when ShowFeeds is false
}
```

Behaviour:
- Library icon: active when `Active == "library"`; otherwise links to `/app/library/dashboard`.
- Compass icon (new): active when `Active == "discover"`; otherwise links to `/app/library/discover`.
- Feeds dropdown rendered only when `ShowFeeds`.
- User menu unchanged.

Net: one component replaces two, one place to add new top-level destinations, no behaviour change for existing pages.

## Edge cases

| Case | Behaviour |
|---|---|
| User searches for an album they previously **removed** (`status='removed'`) | Result card shows `Removed` state with a `Re-acquire` button that calls `POST /app/library/albums/{albumId}/library`. Same effect as adding from radar, just from a different starting state. (Reuses the same handler.) |
| User clicks `+ Add to radar` for an album that already has a `user_releases` row (any status) | Backend returns `ErrAlbumAlreadyDecided`. Handler returns the row re-rendered to reflect the actual current state plus an inline error toast. UX never silently lies to the user. |
| Spotify push fails on `Add to library` | Local DB is already consistent (radar gone, owned release present). Log at WARN, return success to the client, refresh radar carousel. The next Spotify feed sync reconciles the saved-library push. Mirrors `RemoveAlbumFromLibrary`. |
| Spotify search rate-limited / errors | Render an error fragment in `#discover-results` with the message; preserve the input value. |
| Search query is empty / whitespace | `#discover-results` renders an empty hint ("Start typing to search Spotify"). No API call. |
| User on `/discover` with zero radar entries | Carousel shows a placeholder strip ("No albums on your radar yet — search below to find some"). Same shape as the existing `carouselStrip` empty-state handling. |

## Telemetry / observability

- `slog.Info` on each radar transition with `userID`, `albumID`, action.
- `slog.Warn` on Spotify push failures inside `PromoteRadarToLibrary` (non-fatal).
- `slog.Warn` on Spotify search failures (returned to the client as an inline error).

No new metrics this round; existing request-level logging is sufficient.

## Testing

Following the wax convention: pure logic gets unit tests, integration is covered by Playwright e2e.

### Unit tests

- `service_test.go`: `SearchAlbumsForDiscover` state-merge — given a fixed list of Spotify search results and a fixed `GetUserAlbumStateBySpotifyIDs` map, the resulting DTOs carry the right `State` and `AlbumID`. Mock both repo and spotify dependencies via small consumer-defined interfaces (per the domain-module archetype rule).

### Playwright e2e

A new spec under `e2e/` covering the golden path:

1. Visit `/app/library/discover` from the dashboard via the compass icon.
2. Empty radar carousel renders the placeholder text.
3. Search for a query, see results.
4. Click `Add to radar` on a result; result row updates to `On radar`; radar carousel now contains the album.
5. Click `Add to library` from the radar carousel; row leaves the carousel; placeholder returns.
6. Navigate to dashboard; album appears in the list view.

Spotify is mocked in e2e via the existing test infrastructure (we'll reuse whatever pattern the album-detail e2e uses for Spotify).

## Migration / rollout

No DB migration. No backwards-incompatible API changes. Feature is additive — the existing dashboard and album detail pages are unaffected aside from the header bar refactor.

## Open follow-ups (captured, not in scope)

- Wishlist UI (release-level): a similar row affordance on result cards (`+ Add to wishlist (vinyl)`), plus a wishlist section on `/discover` or its own page.
- Format picker on radar→library promotion: replace the digital-only flow with the existing formats modal.
- Filter/sort the radar list (by added date, artist, etc.) once it grows past one screen.
- "Hidden albums" — a separate album-level concern flagged in the album-states spec.
