# Carousel Multi-View — Implementation Notes

## What Was Built

Only one new view was implemented in v1: **Unrated** (albums in the user's library with no rating, ordered by most recently added). The plan's "Recently Added" and "Top Rated" views were deferred — the user scoped this iteration to Unrated only.

## Differences from the Plan

### Views
The plan specified three views: Recently Spun, Recently Added, Top Rated. The actual implementation delivered two: Recently Spun (existing) and Unrated (new). The Unrated view is conceptually different from either planned view — it filters by absence of a rating rather than sorting by one.

### SQL query location
The plan placed both new queries in `db/queries/track_plays.sql`. The `GetUnratedAlbums` query was instead added to `db/queries/album_ratings.sql`, which is the more appropriate file given the query's primary concern is rating status.

### DTO type name and package
The plan introduced `CarouselAlbumDTO` in `listeninghistory/service.go`. During implementation this was reconsidered twice:

1. The name `CarouselAlbumDTO` was rejected as view-specific — renamed to `AlbumSummaryDTO`.
2. The `listeninghistory` package was rejected as the wrong home — `AlbumSummaryDTO` describes library data, not play history. It was moved to `library/service.go`.

### Service method ownership
Because `AlbumSummaryDTO` moved to `library`, `GetRecentlyPlayedAlbums` and `GetUnratedAlbums` were also moved to `library.Service`. This meant `HttpHandler` no longer needed a reference to `listeninghistory.Service` at all, so it was removed from the handler struct and `NewHttpHandler` signature.

### No separate `CarouselTabBar` component
The plan called for a standalone exported `CarouselTabBar` component. The tab bar was instead inlined directly inside `CarouselSection`, keeping the component surface smaller. There was no identified need for the tab bar to be rendered independently.

### `RecentAlbumDTO` removed
The plan didn't explicitly address `RecentAlbumDTO`. In practice it became dead code once `GetRecentlyPlayedAlbums` switched to returning `[]AlbumSummaryDTO`, and was deleted.

### `parseInterfaceTime` check dropped in carousel path
The original `GetRecentlyPlayedAlbums` in `listeninghistory` used `parseInterfaceTime` to validate `LastPlayedAt` and skip rows with unparseable timestamps. Since `AlbumSummaryDTO` has no `LastPlayedAt` field and the carousel doesn't use that value, the check was dropped in the new `library.Service` implementation.

## Plan Inaccuracies

- The Files table listed `listeninghistory/service.go` as the home for the shared DTO and new service methods. This was incorrect — that package has no business owning library view queries.
- The plan noted `CarouselAlbumDTO` as the type name without flagging it as a candidate for review. The name was too view-specific and needed to be changed.
- The plan didn't account for the fact that `library.Service` already imported `listeninghistory`, making the original placement a potential circular import risk if the dependency direction had been reversed.
