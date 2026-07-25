# spotify — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- Exposes two services: `Service` (general user-scoped operations) and `AuthService` (OAuth flow + token refresh). `AuthService` lives in `auth.go` alongside the package's auth errors.
- Most operations flow through the vendor SDK's per-user `*spotify.Client`, built via `Service.Client(ctx, userId)`. `Client` in `client.go` covers endpoints the SDK does not expose, issued as direct HTTP requests.
- All Spotify HTTP traffic — SDK and raw `client.go` — routes through one shared `guard` (`ratelimit.go`), the process-wide rate-limit transport. It paces with a token bucket and, on a 429, pauses every call for `Retry-After`, returning `*ErrRateLimited`. The guard is injected into the SDK path via the `oauth2.HTTPClient` context key in `auth.go` and into the raw client via `NewClient`.
- `Service.Client` reuses a cached access token (persisted encrypted by the `user` module) until it nears expiry, refreshing and re-persisting only then — instead of exchanging the refresh token on every call.
- The vendored SDK (v2.4.3) still targets pre-February-2026 API paths for several writes — `POST /users/{id}/playlists`, `/playlists/{id}/tracks`, `PUT/DELETE /me/albums` — which now return 403 for Development-mode apps. `client.go` issues the migrated endpoints directly: `POST /me/playlists`, `/playlists/{id}/items` (note: the response element field renamed `track`→`item`), and `/me/library`. Reach for `client.go` for any write the SDK still gets wrong, not the SDK method. A 403 on a write while reads succeed means the SDK is hitting a removed path — see the [February 2026 migration guide](https://developer.spotify.com/documentation/web-api/tutorials/february-2026-migration-guide).
