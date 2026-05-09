# spotify — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- Exposes two services: `Service` (general API operations) and `AuthService` (OAuth flow + token refresh). Files: `spotify.go` and `auth.go`.
