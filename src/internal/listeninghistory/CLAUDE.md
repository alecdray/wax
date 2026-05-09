# listeninghistory — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints currently — `adapters/` is intentionally absent.
- Owns the hourly cron task `SyncListeningHistoryTask` (see `task.go`); registered by `server/`. Depends on `spotify.Service` to pull recently-played items per user.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported directly in `service.go`). See `docs/architecture/refactor-backlog.md`.
