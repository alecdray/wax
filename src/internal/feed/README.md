# feed

Background sync of external sources (currently Spotify saved albums) into the user's library.

## Responsibility

`feed` owns the user's connection between an external source (Spotify) and their `library` collection. A `FeedDTO` represents one such connection plus its sync state — when it last ran, whether it succeeded, and whether it is currently due to run again. The module coordinates pulling saved-album data from `spotify.Service`, shaping it into `library.AlbumDTO` values, and handing them off to `library.Service` to persist.

## Key types

- `FeedDTO` — one user/source connection with `LastSyncStatus`, `LastSyncStartedAt`, `LastSyncCompletedAt`. Methods (`IsSyncStale`, `SetSyncing`, `SetSyncSuccess`, `SetSyncFailed`) drive the state machine.
- `MinStaleDuration` — how long after a successful sync the feed is considered fresh.

## Boundaries

- **Inbound:** `auth` and `library/adapters` consume `*feed.Service` to upsert and trigger feeds.
- **Outbound:** depends on `spotify.Service` and `library.Service`, both injected via `NewService`.
- **Adapters:** none — the module has no HTTP entrypoints. Triggers happen via the cron task and via library/auth handlers calling into `*Service`.

## Background tasks

- `SyncStaleSpotifyFeedsTask` — cron task (every minute) that finds stale Spotify feeds and syncs each.
- `SyncSpotifyFeedTask` — ad-hoc task scheduled by handlers for an explicit feed sync.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
