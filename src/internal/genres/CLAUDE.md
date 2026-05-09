# genres — utility

Rules: ../../../docs/architecture/archetypes/utility.md

Module-specific notes:
- Builds a genre DAG from `data.json` (Wikidata-derived) using `//go:embed`. Provides fuzzy lookup via `lithammer/fuzzysearch`.
- Callers use `Load()` to construct the DAG and then operate on it.
