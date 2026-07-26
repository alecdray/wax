# Genres — domain

A curated, app-owned genre taxonomy auto-assigned to every album. Genres are the deliberate exception to the "no global taxonomy" rule that governs tags.

## Primary genres

The taxonomy is a set of ~20 **primary genre buckets** curated from the Wikidata genre graph. Each bucket is a high-level, mutually-intelligible label (e.g. "Rock," "Jazz," "Electronic").

Primary genres are:
- **App-curated**, not user-authored — the taxonomy is fixed and shared across all users.
- **Derived**, not typed — an album's genres are assigned by auto-enrichment from Discogs data, not by user input.
- **Read-only** to users — displayed as badges, not editable. (Future: per-album user overrides that coexist with auto-assignment.)

## Auto-assignment

Album-genre assignments are derived through the genregraph:
1. Discogs data provides genre/style labels for each album.
2. The genregraph maps those labels to primary genre buckets.
3. Auto-enrichment assigns the derived primary genres to the album.

Assignments are deterministic given the same Discogs data and graph — re-running enrichment on the same album produces the same genres.

## Relationship to tags

Genres and tags serve different purposes:

| | Genres | Tags |
|---|---|---|
| Author | App-curated | User-defined |
| Scope | Shared across all users | Per-user |
| Assignment | Auto-derived from Discogs | Manually assigned |
| Taxonomy | Fixed primary buckets | Freeform, per user |