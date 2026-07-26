# genres — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Owns `album_genres` (resolved leaf nodes) and `album_genre_enrichment` (per-album processed marker), both keyed by `album_id` only — genres are album-intrinsic and global, not per user.
- Primary derivation is a pure function of stored leaf genres + the `genregraph` allowlist, computed on read (`AlbumPrimaries`), so changing the allowlist needs no re-backfill.
- The enrichment task reads the album catalog through the `AlbumGenreSource` interface (defined here, satisfied by `library`), so this module never imports `library` — `library` imports it. The discogs client self-throttles; the task only bounds batch size.

## Domain docs

| Doc | What | When |
|---|---|---|
| [genres](../../../docs/domain/genres.md) | primary genre taxonomy, auto-assignment pipeline, relationship to tags | working on enrichment logic, the allowlist, or the read-time derivation of primaries |

## Product docs

| Doc | What | When |
|---|---|---|
| [genres](../../../docs/product/genres.md) | curated, auto-assigned genre badges on albums | deciding enrichment behaviour or how genre assignments should work |

## Relevant ADRs

| ADR | Decision | When |
|---|---|---|
| [ADR 0009](../../../docs/adr/0009-primary-genres-curated-facet.md) | genres are a curated allowlist mapped via Wikidata ancestor graph, computed at read time — not derived from the graph's own top level | questioning the curated-bucket approach, most-specific-ancestor selection, or why primaries are not stored but recomputed |
