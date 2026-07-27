# crates — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Crates owns `crates` and `crate_albums` tables. Album data is owned by `library` — read through `library.Service.GetAlbumsByIDs` and `library.Service.GetAlbumsInLibrary`; never query `albums` or related tables directly from this module's repo.
- `SearchNonMembers` is a service-layer composition: calls `library.Service.GetAlbumsInLibrary`, excludes current member IDs, filters in Go by query string. No cross-module SQL.
- Add/remove responses dispatch a `crateUpdated` HTMX event (via `httpx.SetHXTrigger`) so the crate detail member list re-fetches itself without a full page reload.

## Domain docs

| Doc | What | When |
|---|---|---|
| [crates](../../../docs/domain/crates.md) | crate ownership rules, membership model | working on crate reads/writes or membership logic |

## Product docs

| Doc | What | When |
|---|---|---|
| [crates](../../../docs/product/crates.md) | user-facing feature description | deciding crate behaviour or how a flow should work |
