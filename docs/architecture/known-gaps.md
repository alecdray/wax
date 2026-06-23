# Known Architectural Gaps

This doc tracks current architectural violations in the codebase — places where the rules in [`archetypes/`](archetypes/) (and module `CLAUDE.md` files) don't yet match reality. Each entry describes the gap, why it exists, and what closing it would require.

Gaps are tracked here (and not enumerated inside the archetype docs themselves) so the rule docs stay durable and conceptual, while the concrete list of divergences lives in one searchable place.

## Vendored Spotify SDK is behind the February 2026 API migration

The pinned `github.com/zmb3/spotify/v2@v2.4.3` SDK still calls API paths that Spotify removed/renamed in its [February 2026 migration](https://developer.spotify.com/documentation/web-api/tutorials/february-2026-migration-guide) (e.g. `POST /users/{id}/playlists`, `/playlists/{id}/tracks`, `PUT/DELETE /me/albums`). Those now return 403 for Development-mode apps, so the affected writes are hand-rolled against the migrated endpoints in `spotify/client.go` instead of going through the SDK. Closing the gap means upgrading (or replacing) the SDK to a version that targets the current endpoints, then folding the hand-rolled calls back behind it. Until then, `spotify/client.go` is the source of truth for those writes — see [`src/internal/spotify/CLAUDE.md`](../../src/internal/spotify/CLAUDE.md).

A 403 on a *write* with valid scopes (reads fine) is the tell-tale symptom: check the migration guide's removed/renamed endpoint list before assuming a scope or auth bug.
