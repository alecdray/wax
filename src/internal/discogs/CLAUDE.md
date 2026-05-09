# discogs — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- Has `genres.go`/`genres_test.go` alongside the canonical `client.go`/`entities.go`/`service.go`. This is intentional: the genres logic here is a thin Discogs-specific adapter (compound-term splitting tuned for Discogs strings such as "Funk / Soul" and "Folk, World, & Country") over the `genres` utility. It stays here — see refactor backlog entry for rationale.
