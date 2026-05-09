# notes — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Service renders user notes via `goldmark` (Markdown → HTML); rendering is a service concern, not an adapter concern.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported directly in `service.go`); `/app/notes/...` routes still registered in `server/server.go`. See `docs/architecture/refactor-backlog.md`.
