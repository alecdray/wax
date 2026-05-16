# spotify

The primary data source for Wax. Wraps the Spotify Web API and is the only module that talks to it directly.

## Role

Spotify is used as a read-mostly backend for music metadata, library state, and listening activity. Users authenticate with Spotify (no separate account creation), and Wax stores album/artist/track data pulled from Spotify locally so the app can render and reason about it without a round-trip per view.

When a user mutates their wax library, the change is mirrored back to their Spotify saved library so the two stay aligned.

## Constraints

- The recently-played endpoint returns at most the last 50 tracks. Listening history is best-effort; gaps occur during long sessions or when polling falls behind.
- Saved-albums and saved-tracks page at 50 per request. Full backfills require batched pagination; steady-state sync should be incremental to avoid rate limits.
- The catalog search endpoint caps at 10 results per query.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/external-client.md`](../../../docs/architecture/archetypes/external-client.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
