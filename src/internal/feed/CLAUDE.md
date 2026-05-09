# feed — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints currently — `adapters/` is intentionally absent.
- Owns the cron task `SyncStaleSpotifyFeedsTask` and the on-demand `SyncSpotifyFeedTask` (see `task.go`); registered by `server/`.
- Depends on `spotify.Service` and `library.Service` for syncing the saved-albums feed.
- NOT YET COMPLIANT: missing `repo.go` and `README.md`. See `docs/architecture/refactor-backlog.md`.
