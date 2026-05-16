# feed

Background sync of external sources (currently Spotify saved albums) into the user's library.

## Responsibility

`feed` owns the connection between an external source and the user's `library` collection. It tracks per-feed sync state (last run, success/failure, staleness) and runs both scheduled and on-demand syncs that pull data from the external source and hand it off to `library.Service` to persist.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
