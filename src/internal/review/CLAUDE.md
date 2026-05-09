# review — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Closest existing module to the canonical domain-module shape. Pure-logic types live in `rating.go` and `state.go` with their own tests — good example of "split by topic" within a module root.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported directly in `service.go`); `/app/review/...` routes still registered in `server/server.go`. See `docs/architecture/refactor-backlog.md`.
