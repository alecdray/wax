# genregraph

A Wikidata-derived genre Directed Acyclic Graph (DAG) that provides genre lookup and primary-genre derivation.

## Responsibility

`genregraph` builds a genre DAG from an embedded `data.json` (Wikidata-derived taxonomy) using `//go:embed`. It is a stateless utility — no database, no HTTP, no service layer.

## Primary genre derivation

The DAG maps any genre node to its most-specific curated **primary genre** buckets. A curated allowlist defines which nodes count as primary genres; the `Primaries` method walks up the DAG from a given genre node to find the most-specific ancestor(s) in the allowlist.

## Fuzzy lookup

Provides fuzzy-search over genre names via `lithammer/fuzzysearch` — callers use `Load()` to construct the DAG, then call lookup methods to resolve user-typed genre names to canonical nodes.

## Relationship to the genres module

`genregraph` is a utility — it provides the taxonomy and derivation logic. The [genres module](../genres/README.md) is the domain module that persists per-album genres and calls `genregraph.Primaries` to derive display badges. genregraph does not import genres.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/utility.md`](../../../docs/architecture/archetypes/utility.md)
- Module-specific notes: [`./AGENTS.md`](./AGENTS.md)
- ADR: [ADR 0009 — Primary Genres](../../../docs/adr/0009-primary-genres-curated-facet.md)