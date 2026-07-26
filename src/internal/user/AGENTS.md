# user — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints — `adapters/` is intentionally absent. The user-facing surface lives in `auth`.
- `UserDTO` carries the encrypted Spotify refresh token, plus a cached encrypted Spotify access token + expiry (`SetSpotifyAccessToken` / `CachedSpotifyAccessToken`); `core/cryptox` decrypts them on demand. The `spotify` module owns the refresh/caching policy — this module just persists the tokens.
