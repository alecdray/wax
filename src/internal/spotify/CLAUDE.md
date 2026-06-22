# spotify — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- Exposes two services: `Service` (general user-scoped operations) and `AuthService` (OAuth flow + token refresh). `AuthService` lives in `auth.go` alongside the package's auth errors.
- Most operations flow through the vendor SDK's per-user `*spotify.Client`, built via `Service.Client(ctx, userId)`. `Client` in `client.go` covers endpoints the SDK does not expose, issued as direct HTTP requests.
- The vendored SDK (v2.4.3) still targets pre-February-2026 API paths for several writes — `POST /users/{id}/playlists`, `/playlists/{id}/tracks`, `PUT/DELETE /me/albums` — which now return 403 for Development-mode apps. `client.go` issues the migrated endpoints directly: `POST /me/playlists`, `/playlists/{id}/items` (note: the response element field renamed `track`→`item`), and `/me/library`. Reach for `client.go` for any write the SDK still gets wrong, not the SDK method. A 403 on a write while reads succeed means the SDK is hitting a removed path — see the [February 2026 migration guide](https://developer.spotify.com/documentation/web-api/tutorials/february-2026-migration-guide).
