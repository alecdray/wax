# tags — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Tag normalization (lowercase, trim, strip non-letter/digit/`-&`) lives in `service.go` as a private helper — keep it there; it's domain logic, not a utility.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported directly in `service.go`); `/app/tags/...` routes still registered in `server/server.go`; no `README.md`. See `docs/architecture/refactor-backlog.md`.
