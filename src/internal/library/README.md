# library

The user's music library: albums, artists, tracks, releases, and the user's relationship to all of them.

## Responsibility

`library` owns the user's relationship to albums — what's in their collection, what physical formats they own, what's on their watch list, what's been recently played. The album is the central aggregate: an `AlbumDTO` composes data from peer modules (ratings, tags, sleeve notes, last-played) into one shape that the album view UI binds to.

`library` also owns the **album view UI** — every user-facing surface that centres on albums, including the dashboard, album-detail page, the discover flow, and the modals that mutate album state. When peer modules (review, tags) mutate album state, they broadcast an `album-changed` HTMX event; a hidden listener in the library header bar responds by calling `GET /app/library/album-surfaces`, which re-renders the affected album surfaces as OOB swaps.

## Album states

Two independent relationship dimensions exist between a user and an album:

- **Ownership** — per-format: wishlist, owned, or removed. An album appears in the library when at least one of its formats is owned or wishlisted.
- **Radar** — a "watching this" bookmark. An album is radar-eligible unless it is currently owned or wishlisted; a `removed` album can be put (back) on the radar ([ADR 0005](../../../docs/adr/0005-radar-eligibility-excludes-only-owned-wishlisted.md)). Owning or wishlisting an album clears its radar entry. Albums reach the radar through in-app actions or from a Spotify-side inbox playlist the user opts into, which a periodic sync ingests; an inbox track whose album is already owned or wishlisted is dropped without a radar entry. See [ADR 0004](../../../docs/adr/0004-spotify-radar-playlist-entry.md).

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
