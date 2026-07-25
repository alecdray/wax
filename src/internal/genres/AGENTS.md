# genres — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Owns `album_genres` (resolved leaf nodes) and `album_genre_enrichment` (per-album processed marker), both keyed by `album_id` only — genres are album-intrinsic and global, not per user.
- Primary derivation is a pure function of stored leaf genres + the `genregraph` allowlist, computed on read (`AlbumPrimaries`), so changing the allowlist needs no re-backfill.
- The enrichment task reads the album catalog through the `AlbumGenreSource` interface (defined here, satisfied by `library`), so this module never imports `library` — `library` imports it. The discogs client self-throttles; the task only bounds batch size.
