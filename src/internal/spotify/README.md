# spotify

The primary data source for Wax. Wraps the Spotify Web API and is the only module that talks to it directly.

## Role

Spotify is used as a read-mostly backend for music metadata, library state, and listening activity. Users authenticate with Spotify (no separate account creation), and Wax stores album/artist/track data pulled from Spotify locally so the app can render and reason about it without a round-trip per view.

When a user mutates their wax library, the change is mirrored back to their Spotify saved library so the two stay aligned.

The module also exposes the operations for a dedicated, per-user playlist that acts as a Spotify-side inbox for the radar: creating the playlist when the user opts in, reading its contents, and removing tracks from it. Scheduling the reads and clears is the `feed` module's job; persisting the resulting albums is `library`'s. Reading and modifying that playlist requires playlist read/modify access, requested as part of the standard Spotify connection; users who connected before the feature existed grant it through a re-authentication the first opt-in triggers. The album-level meaning of this inbox lives with the [library module](../library/README.md) and [ADR 0004](../../../docs/adr/0004-spotify-radar-playlist-entry.md).

## Constraints

- The recently-played endpoint returns at most the last 50 tracks. Listening history is best-effort; gaps occur during long sessions or when polling falls behind.
- Saved-albums and saved-tracks page at 50 per request. Full backfills require batched pagination; steady-state sync should be incremental to avoid rate limits.
- The catalog search endpoint caps at 10 results per query.
- Playlists hold tracks, not albums. The radar inbox therefore resolves each track to its album and deduplicates, so one track and a full tracklist produce the same single radar entry.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/external-client.md`](../../../docs/architecture/archetypes/external-client.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
