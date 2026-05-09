# library — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- `service.go` is ~680 lines and still tangles three concerns (user-collection aggregate, formats/releases, sorting/filtering view logic). A topic-file split is open for review — see `docs/architecture/refactor-backlog.md`.
- Library owns the album view UI. Inline content from peer modules (e.g. sleeve notes from `notes`) is rendered by library's adapters using the peer module's `*Service`. Peer adapters never import `library/adapters` and library's adapters never import peer adapters.
- Repo wraps cross-module rating queries (`GetLatestUserAlbumRating`, `GetUserAlbumRatingLog`, etc.) that should eventually be replaced with calls to `reviewService` — flagged with `// TODO` in `repo.go`.
