# listeninghistory — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints — `adapters/` is intentionally absent.
- Owns the hourly cron task `SyncListeningHistoryTask` (see `task.go`).
- Depends on `spotify.Service` to pull recently-played items per user.

## Domain docs

| Doc | What | When |
|---|---|---|
| [listeninghistory](../../../docs/domain/listeninghistory.md) | play record model | working on listening history reads/writes or sync logic |

## Product docs

| Doc | What | When |
|---|---|---|
| [listening-history](../../../docs/product/listening-history.md) | play records over time | deciding sync behaviour or how listening history should work |
