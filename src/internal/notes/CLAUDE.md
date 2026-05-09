# notes — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Pure data + markdown-rendering service. No HTTP entrypoints — `adapters/` is intentionally absent in the target shape. `library/adapters/` renders sleeve notes inline on the album view and depends on `notes.Service` for upsert/fetch and `notes.RenderMarkdown` for HTML.
- Persistence type is `AlbumNote`. "Sleeve note" is a UI label only — don't introduce parallel naming.
- NOT YET COMPLIANT: missing `repo.go` (sqlc imported directly in `service.go`); `adapters/` still exists today and forbidden-imports `library/adapters` — to be removed during refactor (UI moves to library, routes re-prefix). See `docs/architecture/refactor-backlog.md`.
