# genres

An album's structured **genre facet**: app-curated primary genres for coarse library filtering.

## Responsibility

`genres` owns the relationship between albums and the genre graph. It persists each album's resolved
**leaf genres** (Discogs-derived Wikidata Q-ids) and derives each album's **primary genres** from them —
the small, curated set of broad buckets (rock, pop, electronic, …) an album is filtered by. Genres are
app-curated and album-intrinsic, so storage is **global per album**, not per user — unlike free-form
[tags](../tags/README.md), which are a user vocabulary. See [ADR 0009](../../../docs/adr/0009-primary-genres-curated-facet.md)
and the [data model](../../../docs/architecture/data-model.md).

## Primaries

A leaf genre maps to its **most-specific** allowlisted primaries by walking the genre graph: the deepest
matching bucket on each ancestor path wins (a death-metal album lands in *metal*, not *rock*), while an
album spanning unrelated branches keeps one primary per branch (hyperpop → *pop* + *electronic*). The
allowlist and the mapping live in the [`genregraph`](../genregraph/) utility; this module unions the
results across an album's leaf genres. An album with no primary is *uncategorized*.

## Enrichment

A background task resolves genres from Discogs for albums not yet processed (a one-time backfill plus
ongoing coverage for newly-synced albums), bounded per run so the Discogs client's own throttling paces it.
Each album records that it was processed, so an album that resolved to nothing is distinguishable from one
not yet seen. The album catalog is supplied through `AlbumGenreSource`, which the owner of album metadata
satisfies — this module never imports it.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
