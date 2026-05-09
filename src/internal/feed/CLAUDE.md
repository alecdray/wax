# feed — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints — `adapters/` is intentionally absent.
- Owns the cron task `SyncStaleSpotifyFeedsTask` and the on-demand `SyncSpotifyFeedTask` (see `task.go`).
- Depends on `spotify.Service` and `library.Service` for syncing the saved-albums feed.
