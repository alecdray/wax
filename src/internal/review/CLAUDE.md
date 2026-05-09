# review — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Pure-logic types are split across `rating.go` (rating values, scoring questions, labels) and `state.go` (rating-state machine — snoozing, rerate timing). This is the canonical example of a *justified* multi-topic split per the archetype rules: distinct concepts, no shared types, no methods crossing them. Most domain modules should use a single `<package>.go` topic file by default.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported directly in `service.go`); `/app/review/...` routes still registered in `server/server.go`. See `docs/architecture/refactor-backlog.md`.
