# genregraph — utility

Rules: ../../../docs/architecture/archetypes/utility.md

Module-specific notes:
- Builds a genre DAG from `data.json` (Wikidata-derived) using `//go:embed`. Provides fuzzy lookup via `lithammer/fuzzysearch`.
- Callers use `Load()` to construct the DAG and then operate on it.
- Owns the curated **primary genre** allowlist and maps a genre node to its most-specific primaries (`Primaries`); see [ADR 0009](../../../docs/adr/0009-primary-genres-curated-facet.md). The `genres` domain module persists per-album genres and calls this to derive primaries.
