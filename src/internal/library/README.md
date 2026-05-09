# library

The user's music library: albums, artists, tracks, releases, and the user's collection of all of these.

## Responsibility

`library` owns the user's relationship to albums — what's in their collection, what physical formats they own, what's been recently played, what's queued for re-rating. The album is the central aggregate: an `AlbumDTO` carries its artists, tracks, releases, and rating in one shape that other modules (`review`, `tags`, `notes`, `feed`) consume.

`library` also owns the **album view UI**: the dashboard, album-detail page, and inline overlays that compose data from peer modules into the user-facing presentation.

## Key types

- `AlbumDTO` — the central aggregate. Carries `[]ArtistDTO`, `[]TrackDTO`, `[]ReleaseDTO`, and an optional `*review.AlbumRatingDTO` and `*notes.AlbumNoteDTO`.
- `ReleaseDTO` / `AlbumFormatDTO` — physical-release information for an album (vinyl, CD, cassette).
- `AlbumDTOs` — slice of albums with sorting/filtering/pagination methods used by the dashboard.

## Boundaries

- **Inbound:** consumed by `feed` (during Spotify sync, which constructs `AlbumDTO` values) and by adapters in `tags`, `review`, `notes` that need to load an album to render their feature.
- **Outbound:** depends on `spotify`, `listeninghistory`, `tags`, `notes`, `review` services — all injected via `NewService`.
- **Adapters:** library renders inline content from peer modules using their `*Service`. Library's adapters never import a peer module's `adapters` package, and peer adapters never import `library/adapters`.

## Background tasks

None. Library is read/write through HTTP and the `feed` task; it does not run its own cron jobs.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Refactor backlog: [`../../../docs/architecture/refactor-backlog.md`](../../../docs/architecture/refactor-backlog.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
