# library — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- `service.go` is 1009 lines and tangles three concerns (user-collection aggregate, formats/releases, sorting/filtering view logic). The fix is topic files (`library.go`, `formats.go`, `view.go`) within the module — *not* a split into separate modules. See `docs/architecture/refactor-backlog.md`.
- Library owns the album view UI. Inline content from peer modules (e.g. sleeve notes from `notes`) is rendered by library's adapters using the peer module's `*Service`. Peer adapters never import `library/adapters` and library's adapters never import peer adapters.
- NOT YET COMPLIANT: `adapters/formats.go` contains business logic that belongs in the service; `/app/library/...` routes still registered in `server/server.go`.
