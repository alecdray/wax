# user — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints currently — `adapters/` is intentionally absent (the user-facing surface lives in `auth`).
- `UserDTO` carries the encrypted Spotify refresh token and uses `core/cryptox` to decrypt on demand — boundary with `auth` is fuzzy and may move once `auth` is properly shaped.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported directly in `service.go`). See `docs/architecture/refactor-backlog.md`.
