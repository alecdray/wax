# library — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- `service.go` is 1009 lines and currently aggregates albums, artists, tracks, releases, formats, and user-collection state — structural reorg candidate (possible `artists`/`tracks` extraction). See `docs/architecture/refactor-backlog.md`.
- `adapters/sleeve_notes.templ` overlaps the `notes` module's templates — boundary needs clarifying.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported in `service.go`); `adapters/formats.go` contains business logic that belongs in the service; `/app/library/...` routes still registered in `server/server.go`.
