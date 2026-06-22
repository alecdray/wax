# feed

Background sync of external sources into the user's album relationships.

## Responsibility

`feed` owns the connection between an external source and the user's album data. It tracks per-feed sync state (last run, success/failure, staleness) — and, where a kind needs one, the external source handle for that feed (such as the radar inbox playlist's id). It runs both scheduled and on-demand syncs that pull from the external source and hand off to `library.Service` to persist. Each feed kind fixes its source and which album relationship it feeds — Spotify saved albums sync into the owned library; the Spotify radar inbox playlist syncs into the radar. See [ADR 0004](../../../docs/adr/0004-spotify-radar-playlist-entry.md).

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
