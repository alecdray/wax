# user — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints — `adapters/` is intentionally absent. The user-facing surface lives in `auth`.
- `UserDTO` carries the encrypted Spotify refresh token; `core/cryptox` decrypts it on demand.
