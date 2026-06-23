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

## Rate limiting

Spotify rate-limits per app over a rolling 30-second window, so every call this module makes — whether through the vendor SDK or the direct-HTTP `client.go` paths — flows through one shared, process-wide guard. The guard paces normal traffic and, on a `429`, pauses all Spotify calls for the response's `Retry-After` duration; issuing requests during that window only lengthens the penalty. Spotify exposes no remaining-quota header (unlike Discogs), so the proactive pace is a fixed, conservative rate rather than an adaptive one — the `Retry-After` pause is the authoritative backstop. While paused, callers fail fast rather than block, so a long penalty never hangs a request — the `feed` cron defers to its next run and user-facing callers surface the rate-limited state. The guard's state is in-process and not persisted across restarts. See [ADR 0006](../../../docs/adr/0006-spotify-rate-limit-guard.md).

The access token is reused until it expires rather than re-exchanged before every call, keeping routine syncs from spending budget on token requests. The cached access token and its expiry are persisted encrypted by the `user` module, the same way the refresh token is, so the cache survives a restart; a refresh runs only once the cached token has expired.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/external-client.md`](../../../docs/architecture/archetypes/external-client.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
