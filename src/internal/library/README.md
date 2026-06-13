# library

The user's music library: albums, artists, tracks, releases, and the user's relationship to all of them.

## Responsibility

`library` owns the user's relationship to albums — what's in their collection, what physical formats they own, what's on their watch list, what's been recently played. The album is the central aggregate: an `AlbumDTO` composes data from peer modules (ratings, tags, sleeve notes, last-played) into one shape that the album view UI binds to.

`library` also owns the **album view UI** — every user-facing surface that centres on albums, including the dashboard, album-detail page, the discover flow, and the modals that mutate album state. When peer modules (review, tags) mutate album state, they broadcast an `album-changed` HTMX event; a hidden listener in the library header bar responds by calling `GET /app/library/album-surfaces`, which re-renders the affected album surfaces as OOB swaps.

## Album states

Two independent relationship dimensions exist between a user and an album:

- **Ownership** — per-format: wishlist, owned, or removed. An album appears in the library when at least one of its formats is owned or wishlisted.
- **Radar** — a "watching this" bookmark for albums *not* in the library. Independent of ownership; bringing the album into the library clears its radar entry.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
