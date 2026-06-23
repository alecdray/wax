# user

The account identity behind every authenticated request.

## Responsibility

`user` owns the user record and its Spotify credentials. A user is created (or matched) on Spotify login, keyed by their Spotify id; there is no separate account system. Beyond identity, the module is the store for the user's Spotify OAuth tokens — both the long-lived refresh token and a cached short-lived access token — and exposes lookups by id and by Spotify id that other modules use to resolve "who is this".

Tokens are held encrypted at rest (`core/cryptox`) and only ever leave the module decrypted, on demand, through the `UserDTO`. The refresh token is the durable credential; the access token plus its expiry are cached so the [spotify module](../spotify/README.md) can reuse a valid token instead of exchanging the refresh token on every call. The refresh/caching *policy* lives in `spotify` — this module only persists what it is told.

## Boundaries

- No HTTP entrypoints — `adapters/` is intentionally absent. The user-facing surface (login, OAuth callback) lives in the [auth module](../auth).
- Owns no album, rating, or relationship data; those modules reference a user by id but their data is theirs.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Token caching rationale: [ADR 0006](../../../docs/adr/0006-spotify-rate-limit-guard.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
