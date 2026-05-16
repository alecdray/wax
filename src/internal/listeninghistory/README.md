# listeninghistory

Records what the user has been playing on Spotify and exposes per-album "last played" times to the rest of the app.

## Role

The source of "Recently Spun" and last-played times anywhere they appear in Wax. Populated by a cron task that polls Spotify's recently-played endpoint for every user with a stored Spotify token, persists the resulting track plays, and back-fills any new album/artist/track rows discovered along the way.

## Constraints

- Spotify exposes only the last 50 recently-played tracks. Sync must run frequently or gaps appear during heavy listening sessions.
- The polling cadence is the floor for accuracy. A user offline for the polling window will have correspondingly thinner history.
- Token failures for one user do not stop the job — that user is skipped with a warning and the loop continues.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
