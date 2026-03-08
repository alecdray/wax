# Carousel Multi-View — Research

## Summary

The carousel multi-view feature enhances the existing `RecentlyPlayedCarousel` component on the dashboard to support multiple distinct views or sections — for example, different time-based groupings of recently played albums ("Today", "This Week", "All Time"), or different content categories (e.g. "Recently Played", "Recently Added", "Highest Rated"). Currently the carousel is a single horizontal scroll strip showing up to 20 recently played albums. This feature would allow users to switch between multiple named views within the carousel area, each backed by potentially different data queries.

## Relevant Code

### Current carousel implementation
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/library/adapters/dashboard.templ` — `RecentlyPlayedCarousel(albums []listeninghistory.RecentAlbumDTO)` renders a single horizontal `carousel carousel-center` using DaisyUI's carousel classes. It is rendered inside `DashboardPage` between `LibraryStats` and `AlbumsTable`. The carousel item is an album art thumbnail (96x96) with title and artist name below.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/library/adapters/http.go` — `GetDashboardPage` calls `listeningHistoryService.GetRecentlyPlayedAlbums(ctx, userId)` and passes the result to `DashboardPage`. No carousel-specific endpoint exists yet.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/server/server.go` — Route `GET /app/library/dashboard` maps to `libraryHandler.GetDashboardPage`. No carousel-specific route exists.

### Listening history data layer
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/listeninghistory/service.go` — `GetRecentlyPlayedAlbums(ctx, userID)` returns `[]RecentAlbumDTO` (fields: ID, SpotifyID, Title, Artists, ImageURL, LastPlayedAt). The `RecentAlbumDTO.Artists` field is a single comma-joined string (not a slice).
- `/Users/shmoopy/Documents/code/repos/shmoopicks/db/queries/track_plays.sql` — `GetRecentlyPlayedAlbums` query groups by album, MAX(played_at) as `last_played_at`, includes `artist_names` via correlated subquery, limits to 20 rows, ordered by `last_played_at DESC`. To support multiple views, this query would need to be parameterized (e.g. by date range, limit, or order) or new sibling queries added.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/db/schema.sql` — `track_plays(id, user_id, track_id, album_id, played_at)` is the source of truth for listening history. `albums(id, spotify_id, title, image_url)` and `album_artists`/`artists` provide display data.

### Library data layer (for non-history views)
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/library/service.go` — `GetAlbumsInLibrary` and `GetLibrary` are the access points for the user's full library. `AlbumDTO` has `Rating *review.AlbumRatingDTO`, `Releases ReleaseDTOs`, `LastPlayedAt *time.Time`. These could back a "Recently Added" or "Highest Rated" view.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/db/queries/user_releases.sql` — `GetUserReleases` returns all releases with user metadata (including `added_at`). Could support "Recently Added" carousel.

### UI framework
- **DaisyUI carousel**: The project uses DaisyUI 5.5.18. The carousel component is `display: inline-flex; overflow-x: scroll; scroll-snap-type: x mandatory`. It has no built-in tab/multi-view capability — tabs would be a separate DaisyUI component (`tabs`, `tab`, `tab-content`). DaisyUI tabs can be purely CSS-driven using radio inputs or links with `#` anchors. Alpine.js (already loaded via CDN in `root.templ`) could also drive tab switching client-side.
- **AlpineJS**: `defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"` is loaded globally. `x-data`, `x-show`, `x-bind:class` etc. can drive client-side view switching with no HTMX round-trip needed if all data is rendered server-side.
- **HTMX**: `hx-get` + `hx-swap` is the established pattern for fetching partial HTML. A lazy-loaded carousel view could fetch its content on tab click, matching the `FeedsDropdownContent` pattern.

### Existing patterns for multi-section UI
- `FeedsDropdownContent` uses `hx-trigger="every 5s"` and `hx-get` for polling. This is the closest example of a component that fetches its own data.
- `AlbumsTable` uses `hx-get` with query params to re-sort, swapping `outerHTML`. A "switch carousel view" action could similarly use `hx-get` to swap the carousel content with a different data set.
- No existing tab or multi-view component exists in the templates directory — `effects.templ`, `layout.templ`, `modal.templ`, `navbar.templ`, `ticker.templ`, `tooltip.templ`, `root.templ`, `icons.templ`, `utils.templ` are the full set.

## Architecture

### Option A — Server-rendered tabs with HTMX (lazy loading)
Each view is a named endpoint. Clicking a tab fires `hx-get="/app/library/dashboard/carousel?view=recently-played"` (or `?view=recently-added`, etc.), swapping the carousel content div. The active tab indicator is managed either by OOB swap or by returning a tab+content fragment together.

```
GET /app/library/dashboard/carousel?view=<name>
  → new HTTP handler: GetCarousel(w, r)
    → switch view name:
        case "recently-played": listeningHistoryService.GetRecentlyPlayedAlbums()
        case "recently-added":  libraryService.GetAlbumsInLibrary() → sort by date desc → take N
        case "top-rated":       libraryService.GetAlbumsInLibrary() → sort by rating desc → take N
    → render CarouselContent(albums, view) fragment
```

**Pros**: works without JS beyond HTMX, content is fetched on demand, server controls what data each view shows.
**Cons**: each view switch costs a round-trip; need a new route and handler.

### Option B — All views rendered server-side, Alpine.js tab switching
All views are rendered into the page on initial load. Alpine.js hides/shows the active view via `x-show`. No additional routes needed.

```
DashboardPage renders:
  CarouselMultiView(views: [recently-played, recently-added, top-rated])
    → tabs row (Alpine x-data activeView)
    → for each view:
        <div x-show="activeView === 'recently-played'">
          CarouselStrip(albums)
        </div>
```

**Pros**: instant tab switching, no new routes or handlers.
**Cons**: all view data must be fetched on initial page load (potentially expensive); `DashboardPageProps` needs new fields; carousel data for non-recently-played views requires additional service calls in `GetDashboardPage`.

### Option C — Hybrid: initially render active view, lazy-load others on first switch
Default view rendered server-side. Switching to an unloaded view fires an HTMX request that fetches and caches the content in-DOM.

### Recommended: Option A (HTMX lazy loading)
Consistent with existing HTMX patterns (albums table sort, feeds dropdown). Avoids loading expensive library data on page load. Keeps `GetDashboardPage` fast. A single new route is clean.

## Existing Patterns

- **Component architecture**: large dashboard components (`AlbumsTable`, `RecentlyPlayedCarousel`) are top-level `templ` functions with props structs. Sub-components are unexported helpers.
- **HTMX partial swaps**: `hx-target`, `hx-swap="outerHTML"` is the standard pattern. The carousel wrapper div would need a stable `id` attribute for targeting.
- **Query parameters for data variant**: `AlbumsTable` uses `?sortBy=&dir=` query params. The carousel could use `?view=recently-played` etc.
- **Data fetching in handlers**: `GetDashboardPage` calls multiple services. The carousel handler would call 1-2 services based on the `view` param.
- **DashboardPageProps struct**: in `dashboard.templ`, this struct aggregates all data needed by the page. Adding a default carousel view to it is straightforward; additional views stay on-demand.
- **No client-side routing**: the app uses HTMX for partial updates, not a SPA framework. Alpine.js is used for purely local UI state (dropdowns, modals, etc.).

## Constraints & Risks

- **`RecentAlbumDTO` vs `AlbumDTO`**: the currently rendered carousel uses `listeninghistory.RecentAlbumDTO` (with a denormalized `Artists string`). A "Recently Added" or "Top Rated" view would use `library.AlbumDTO` (with `Artists []ArtistDTO`). The carousel item template would need to handle both or be refactored to a common interface/adapter. This is the primary data-model friction point.
- **Album image quality**: both `RecentAlbumDTO.ImageURL` and `AlbumDTO.ImageURL` come from the same `albums.image_url` column. Image resolution is consistent across views.
- **Performance of `GetAlbumsInLibrary`**: this function does 5-6 separate DB queries (releases, albums, artists, tracks, ratings, last-played-at). For a "Recently Added" carousel showing only the top 10 albums, loading the full library is wasteful. A dedicated lightweight query (e.g. `GetRecentlyAddedAlbums`) would be better.
- **"Recently Added" data availability**: `user_releases.added_at` tracks when a release was added to the library. For carousel purposes this is a good proxy. However, albums can have multiple releases (vinyl, digital, CD) — "added date" might mean the earliest or most recent release date.
- **View names and routing**: if views are added/renamed, the URL query param values become a quasi-API. Should use a constrained set of valid values with a default fallback.
- **No existing multi-panel/tab component**: this is the first tab-style UI in the app. A reusable tab component in `core/templates` would be useful but adds scope.
- **Carousel width**: current carousel uses `w-full overflow-x-auto` on the outer wrapper and `w-26` on each item. With tab controls, the total height of the carousel section increases — could affect the fixed-height dashboard scroll area.

## Open Questions

1. What views should be included in v1? Candidates: "Recently Played" (existing), "Recently Added" (library), "Top Rated" (library by rating). Are all three needed, or just two?
2. Should the carousel use HTMX lazy loading per view (Option A) or Alpine.js in-page switching (Option B)?
3. Should a new lightweight DB query be written for "Recently Added" albums, or is reusing `GetAlbumsInLibrary` (full library load) acceptable given library sizes are small for now?
4. How should the carousel item template handle the two different DTO types (`RecentAlbumDTO` vs `AlbumDTO`)? Options: (a) define a common `CarouselAlbumItem` struct and adapt both; (b) overload the carousel template; (c) only use `RecentAlbumDTO` and adapt library albums into it.
5. Should the active view be remembered across page loads (e.g. stored in `localStorage` via Alpine.js, or as a URL query param)?
6. Is there a desired visual design for the tab controls (DaisyUI `tabs` component, custom button group, etc.)?
7. Should the "Recently Added" view filter to only albums in the user's library, or include all albums the user has played (i.e., from `track_plays`)?

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research carousel-multi-view -->
