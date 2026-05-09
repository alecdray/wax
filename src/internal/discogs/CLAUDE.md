# discogs — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- `genres.go` is a Discogs-specific adapter over the `genres` utility — it handles compound-term splitting tuned for Discogs strings (e.g. "Funk / Soul", "Folk, World, & Country") and resolves them against the genre DAG.
