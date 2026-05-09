# genres — utility

Rules: ../../../docs/architecture/archetypes/utility.md

Module-specific notes:
- Currently the only utility module in the codebase.
- Builds a genre DAG from `data.json` (Wikidata-derived) using `//go:embed`. Provides fuzzy lookup via `lithammer/fuzzysearch`.
- Stateless; callers use `Load()` to construct the DAG and then operate on it. No `Service` struct.
