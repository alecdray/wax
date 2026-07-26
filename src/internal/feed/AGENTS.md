# feed — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints — `adapters/` is intentionally absent. The radar inbox's enable control lives in `library/adapters` and calls `Service.EnableRadarInbox`.
- Owns the cron tasks `SyncStaleSpotifyFeedsTask` (saved albums) and `SyncStaleSpotifyRadarFeedsTask` (radar inbox), plus the on-demand `SyncSpotifyFeedTask` (see `task.go`).
- A feed is synced only when *due*: `next_sync_at` (nil = now) gates selection (`GetDueFeedsBatch` / `GetSyncableRadarFeeds`). On success `SetSyncSuccess` schedules `now + SyncInterval`; on failure `SetSyncFailed` increments `ConsecutiveFailures` and backs off exponentially up to `MaxSyncBackoff`. Both Spotify kinds share this cadence. `IsSyncStale`/`MinStaleDuration` are unrelated — they drive only the UI's freshness indicator, not scheduling.
- Depends on `spotify.Service` and `library.Service` for both feeds. The radar inbox sync (`radar.go`) reads a per-user playlist (handle in `feeds.source_ref`) and adds its albums to the radar; its ingest logic talks to those services through narrow interfaces so it can be faked in tests.

## Domain docs

| Doc | What | When |
|---|---|---|
| [feed](../../../docs/domain/feed.md) | feed sync state, inbox playlist handles | working on sync scheduling, feed state, or inbox ingest logic |

## Product docs

| Doc | What | When |
|---|---|---|
| [radar](../../../docs/product/radar.md) | watch albums; eligibility rules; Spotify-inbox entry | deciding sync behaviour or how the inbox should work |

## Relevant ADRs

| ADR | Decision | When |
|---|---|---|
| [ADR 0004](../../../docs/adr/0004-spotify-radar-playlist-entry.md) | a dedicated Wax-managed playlist is the Spotify-side radar entry point; ingestion rules, playlist lifecycle, and opt-in flow | working on the radar inbox sync in `radar.go` or the `EnableRadarInbox` flow |
