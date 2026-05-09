# library — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Topic files: `service.go` (Service + methods only), `album.go` (album aggregate types), `release.go` (release/format types and helpers), `view.go` (`AlbumDTOs` sort/filter/page + `Dashboard`).
- The dashboard aggregate type is named `Dashboard` (not `Library`, to avoid the package-name clash) and is constructed via `NewDashboard(...)`. The `*Service.GetLibrary` method name is preserved since callers expect it.
- Library owns the album view UI. Inline content from peer modules (e.g. sleeve notes from `notes`) is rendered by library's adapters using the peer module's `*Service`. Peer adapters never import `library/adapters` and library's adapters never import peer adapters.
- Repo wraps cross-module rating queries (`GetLatestUserAlbumRating`, `GetUserAlbumRatingLog`, etc.) that should eventually be replaced with calls to `reviewService` — flagged with `// TODO` in `repo.go`.
