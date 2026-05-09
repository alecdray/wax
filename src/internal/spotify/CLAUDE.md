# spotify — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- Exposes two services: `Service` (general user-scoped operations) and `AuthService` (OAuth flow + token refresh). `AuthService` lives in `auth.go` alongside the package's auth errors.
- Most operations flow through the vendor SDK's per-user `*spotify.Client`, built via `Service.Client(ctx, userId)`. `Client` in `client.go` covers endpoints the SDK does not expose, issued as direct HTTP requests.
