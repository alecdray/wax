# spotify — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- Diverges from the canonical layout: has `auth.go` + `spotify.go` instead of `client.go`/`entities.go`/`service.go`. Exposes both `Service` and `AuthService`. See refactor backlog for the planned reorganization (split into `client.go`, `entities.go`, `service.go`; keep `auth.go` for `AuthService` as a separate concern).
