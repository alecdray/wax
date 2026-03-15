---
description: >
  External services Wax depends on or communicates with. Belongs here: what each service provides,
  how authentication works, API constraints, and data flow between Wax and third parties. Does not
  belong here: internal system structure (→ architecture), what users see as a result
  (→ features), or how external data maps to domain entities (→ data-model).
links:
  - architecture
  - data-model
  - roadmap
---

[Parent: wiki](../wiki.md)

# Integrations

External services Wax connects to and what they're used for.

## Spotify

The primary data source. Wax uses Spotify as a read-only backend for music data.

| Purpose | Detail |
|---|---|
| **Authentication** | Users log in via Spotify OAuth2. No separate account creation |
| **Library sync** | Pulls user's saved albums on a recurring schedule |
| **Listening history** | Polls recently played tracks (limited to last 50 by Spotify's API) |
| **Open in Spotify** | Deep links back to Spotify for playback |

**Auth model:** OAuth2 authorization code flow. The Spotify refresh token is stored encrypted in the database and used to issue new access tokens on demand.

**Constraints:**
- Spotify's recently played API only returns the last 50 tracks — listening history is best-effort and requires frequent polling to avoid gaps
- Library data (album metadata, artwork, track listings) comes from Spotify and is stored locally

## MusicBrainz

A secondary metadata source used for enrichment beyond what Spotify provides.

- Open music encyclopedia (no auth required)
- Used to fill gaps in album/artist metadata

