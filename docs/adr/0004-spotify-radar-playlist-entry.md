# A dedicated Spotify playlist is the radar's Spotify-side entry point

Users can put albums on their radar from inside Spotify by dropping tracks into a dedicated, Wax-managed playlist. A periodic sync reads the playlist, derives each track's album, adds those albums to the radar, and removes the tracks it ingested.

Spotify offers no webhooks and no album-level "save for later" that is distinct from the saved library — which Wax already treats as ownership. A playlist is the only writable, user-curated container the API exposes that is not already claimed by other semantics. Because playlists hold tracks rather than albums, every track maps to its album and the set is deduplicated: adding one track or a whole tracklist yields a single radar entry, consistent with the album-as-anchor model.

The playlist is created on demand when the user opts in, not automatically — there is no value in provisioning a playlist for users who never use the feature. Playlist access itself is part of the standard set of permissions requested when connecting Spotify, so a freshly-connected user can opt in without a further prompt; users who connected before this feature existed are re-authenticated the first time they opt in. If the user later deletes the playlist on Spotify, the next sync detects its absence, forgets the stored handle, and stops until the user opts in again — Wax never silently recreates a playlist the user removed.

Two rules govern ingestion:

- **Only successfully-ingested tracks are removed.** A track whose album fails to resolve or import stays in the playlist and is retried next cycle, so a transient error never silently drops an album.
- **An album the user already owns or wishlists is not re-added to the radar.** It is in the library, so the track is removed without a radar entry — treated as "already handled", not an error. A previously `removed` album, by contrast, is radar-eligible and *is* re-added (see [ADR 0005](0005-radar-eligibility-excludes-only-owned-wishlisted.md)).

Overloading the saved library or liked songs was rejected: saved albums already mean ownership, and liked songs are noisy listening behaviour rather than a deliberate "consider this" signal.
