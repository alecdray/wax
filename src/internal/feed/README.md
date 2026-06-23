# feed

Background sync of external sources into the user's album relationships.

## Responsibility

`feed` owns the connection between an external source and the user's album data. It tracks per-feed sync state (last run, success/failure, staleness) — and, where a kind needs one, the external source handle for that feed (such as the radar inbox playlist's id). It runs both scheduled and on-demand syncs that pull from the external source and hand off to `library.Service` to persist. Each feed kind fixes its source and which album relationship it feeds — Spotify saved albums sync into the owned library; the Spotify radar inbox playlist syncs into the radar. See [ADR 0004](../../../docs/adr/0004-spotify-radar-playlist-entry.md).

## Sync cadence and backoff

Scheduled syncs run on a periodic tick, but a feed is only synced when it is *due*. Both Spotify feed kinds — saved albums and the radar inbox — share one recurring incremental cadence; the prior split (saved albums hourly, the radar inbox re-read every tick) was incidental, not designed. Steady-state polls are cheap for both: a saved-albums sync fetches only the window since its last successful run, and a radar poll of an already-cleared inbox is a couple of calls. The expensive full saved-library backfill is reserved for a feed's first sync — and reconnect or an explicit resync — not the recurring tick.

A feed that just failed is not retried on the next tick: failures back off (growing with consecutive failures, up to a cap) and a successful sync resets the feed to the normal cadence. This keeps a persistently-failing feed — a revoked token, a rate-limit penalty — from hammering the external source every tick, which is what previously let one failure sustain a Spotify rate-limit outage ([ADR 0006](../../../docs/adr/0006-spotify-rate-limit-guard.md)). Because the incremental window spans the actual time since the last *successful* sync, a feed that has been backing off still catches up everything it missed once it recovers.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
