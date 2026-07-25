# feed — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints — `adapters/` is intentionally absent. The radar inbox's enable control lives in `library/adapters` and calls `Service.EnableRadarInbox`.
- Owns the cron tasks `SyncStaleSpotifyFeedsTask` (saved albums) and `SyncStaleSpotifyRadarFeedsTask` (radar inbox), plus the on-demand `SyncSpotifyFeedTask` (see `task.go`).
- A feed is synced only when *due*: `next_sync_at` (nil = now) gates selection (`GetDueFeedsBatch` / `GetSyncableRadarFeeds`). On success `SetSyncSuccess` schedules `now + SyncInterval`; on failure `SetSyncFailed` increments `ConsecutiveFailures` and backs off exponentially up to `MaxSyncBackoff`. Both Spotify kinds share this cadence. `IsSyncStale`/`MinStaleDuration` are unrelated — they drive only the UI's freshness indicator, not scheduling.
- Depends on `spotify.Service` and `library.Service` for both feeds. The radar inbox sync (`radar.go`) reads a per-user playlist (handle in `feeds.source_ref`) and adds its albums to the radar; its ingest logic talks to those services through narrow interfaces so it can be faked in tests.
