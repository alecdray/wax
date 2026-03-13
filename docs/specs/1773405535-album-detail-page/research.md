# Album Detail Page — Research

## Summary

A dedicated per-album page (`/app/library/albums/:albumId` or similar) that surfaces all data the system already holds for a single album: cover art, title, artists, release formats with dates, rating and review, tags, and last-played date. The roadmap entry frames it as the primary destination for interacting with an album and as a meaningful improvement to mobile usability — currently all interactions are surfaced through modals or dropdown menus on a table row, which works poorly on small screens. The page exists conceptually; almost all backend data is already gathered by `library.Service.GetAlbumInLibrary`, so this is primarily a routing + template addition.

## Relevant Code

| Path | Role |
|------|------|
| `src/internal/library/service.go` | `AlbumDTO` struct; `GetAlbumInLibrary()` — primary data fetch for a single album, already joins artists, tracks, releases, rating, and tags |
| `src/internal/library/adapters/http.go` | `HttpHandler` — add a new handler here; `NewHttpHandler` receives `libraryService`, `feedService`, `spotifyAuth`, `musicbrainz`, `taskManager` |
| `src/internal/library/adapters/dashboard.templ` | `DashboardHeaderBar`, `AlbumRating`, `AlbumNotesIcon`, `AlbumTagsCell`, `releaseFormatBadge` — reusable components to pull into the detail page |
| `src/internal/review/adapters/rating.templ` | `RatingModal` — modal triggered from the detail page's rate button |
| `src/internal/review/adapters/review_notes.templ` | `ReviewNotesModal` — modal triggered from the detail page's notes button |
| `src/internal/tags/adapters/tags.templ` | `TagsModal` — modal triggered from the detail page's tags button |
| `src/internal/server/server.go` | Route registration — new route wired here |
| `src/internal/core/templates/root.templ` | `RootComponent` — wraps every full page; includes `ModalContainer` in `<body>` |
| `src/internal/core/templates/layout.templ` | `PageLayoutComponent` — simple full-page layout with nav bar (may or may not be suitable for this page) |
| `db/queries/albums.sql` | `GetAlbum :one` — already exists; used by `GetAlbumInLibrary` |
| `src/internal/core/db/sqlc/` | Generated SQLC types (do not edit) |

## Architecture

### Data flow for the detail page (proposed)

```
User clicks album title / navigates to /app/library/albums/:albumId
  → GET /app/library/albums/:albumId
  → libraryAdapters.HttpHandler.GetAlbumDetailPage()
  → library.Service.GetAlbumInLibrary(ctx, userId, albumId)
    → DB queries: albums, artists, tracks, releases, rating, tags
  ← AlbumDTO
  → AlbumDetailPage(album) templ component
  ← Full HTML page (RootComponent wrapper)
```

### Modal interactions remain unchanged

Rating, notes, and tags modals already work against any page that includes the global `ModalContainer` (injected by `RootComponent`). Their HTMX flows (`hx-get /app/review/rating-recommender`, `hx-get /app/review/notes`, `hx-get /app/tags/album`) will work on the detail page without any changes to the review or tags handlers — the OOB swaps that update `AlbumRating` and `AlbumNotesIcon` use element IDs that must exist on the page.

### Existing single-album data fetch

`GetAlbumInLibrary` already fetches everything needed:
- Album core fields (`id`, `spotifyId`, `title`, `imageUrl`)
- `[]ArtistDTO` via `GetAlbumArtistByAlbumId`
- `[]TrackDTO` via `GetAlbumTracksByAlbumId`
- `[]ReleaseDTO` via `GetUserReleasesByAlbumId`
- `*review.AlbumRatingDTO` (rating + review text) via `GetUserAlbumRating`
- `[]tags.TagDTO` via `tagsService.GetAlbumTags`

`LastPlayedAt` is NOT currently fetched in `GetAlbumInLibrary` — it is only populated in the bulk `GetAlbumsInLibrary` path via `listeningHistoryService.GetLastPlayedAtByAlbumIds`. The detail page will need this field too.

## Existing Patterns

- **Full page routes** use `RootComponent` as the outer wrapper. `DashboardPage` is the only current app page — it calls `templates.RootComponent` directly via `DashboardPage.templ`. A second full page follows the same pattern.
- **Header bar** — `DashboardHeaderBar` is a sticky top bar with brand link and user dropdown. A detail page should share the same structure or reuse the component.
- **Handler structure**: extract `userId` via `contextx.NewContextX(r.Context())` + `ctx.UserId()`, call service, render templ component. Errors via `http.Error` (dashboard pattern) or `httpx.HandleErrorResponse` (review/tags pattern). The simpler `http.Error` approach is fine for a page handler.
- **Route registration**: `appMux.Handle("GET /app/library/albums/{albumId}", ...)` follows Go 1.22 path parameter syntax. `albumId` extracted via `r.PathValue("albumId")`.
- **OOB swap IDs on the detail page**: `AlbumRating(album, false)` sets `id="album-rating-{albumId}"` and `AlbumNotesIcon(album, false)` sets `id="album-notes-{albumId}"`. These IDs must be present on the detail page for OOB swaps from rating/notes modals to work.
- **Spotify links**: album titles and artist names link to `https://open.spotify.com/album/{spotifyId}` and `https://open.spotify.com/artist/{spotifyId}` — established pattern in `albumRow`.
- **Navigation**: dashboard currently uses `hx-boost="true"` on its header bar for smooth SPA-style navigation. The detail page can link back to the dashboard the same way.

## Constraints & Risks

1. **`LastPlayedAt` gap in `GetAlbumInLibrary`**: the field is present on `AlbumDTO` but only set in the bulk library fetch. Fix: call `listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, []string{albumId})` inside `GetAlbumInLibrary` (or in the new handler) and populate `albumDto.LastPlayedAt`. This is a small change but requires adding `listeningHistoryService` access to the call path.
2. **OOB swap contract**: the detail page must render `AlbumRating` and `AlbumNotesIcon` with the correct IDs so the post-modal OOB swaps land correctly. These are already parameterized components; the constraint is just remembering to include them.
3. **No `AlbumTagsCell` OOB swap**: the tags modal's `SubmitAlbumTags` handler OOB-swaps `AlbumTagsCell` by ID. If the detail page shows tags in a different component, it won't be updated automatically after saving. Either reuse `AlbumTagsCell` on the detail page (with the same ID contract) or update the tags handler to also return an OOB swap for a detail-page component.
4. **Access control**: `GetAlbumInLibrary` returns `errors.New("album not in library")` if the user has no releases for the album. This naturally gates access — albums not in the user's library 404. This is the right behavior.
5. **Mobile**: the roadmap names this feature explicitly as an improvement to mobile usability. The template should be designed mobile-first, avoiding the table layout used on the dashboard.
6. **Track list display**: `AlbumDTO` carries `[]TrackDTO` (title only, no track number or duration). The detail page could display a plain track list, but track numbers are not stored in the current schema — this is display-only and low risk.

## Open Questions

1. **URL pattern**: `/app/library/albums/{albumId}` is the natural fit. Should this be a full page load (link in `<a>`) or an HTMX fragment swap? A full page (with `hx-boost`) is simpler and more shareable/bookmarkable.
2. **Entry point**: where does the user navigate from? The most natural is clicking the album title (currently a Spotify external link). Should the album title in the table become an internal link, and the Spotify link move to an icon/button? Or is there a separate "Detail" action in the row dropdown menu?
3. **Track list**: display all tracks? Just count? Collapsible section? Track numbers and durations are not in the current schema.
4. **`LastPlayedAt` fix scope**: fix it inside `GetAlbumInLibrary` (making it consistent everywhere) or handle it locally in the new handler? Fixing it in the service is cleaner.
5. **Header**: reuse `DashboardHeaderBar` or create a simpler header for the detail page? The dashboard header includes the feeds dropdown, which is dashboard-specific.
6. **Edit actions on the page**: rating, notes, and tags are currently modal-triggered. On a detail page there's room for inline editing — is that in scope for v1?
7. **Back navigation**: how does the user get back to the dashboard? Browser back button, or an explicit back link?

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research album-detail-page -->
