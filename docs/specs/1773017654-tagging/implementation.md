# Tagging — Implementation

## What Was Built

Full v1 tagging feature as described in the plan:

- Three new DB tables: `tag_groups`, `tags`, `album_tags` (migration `20260309010044_add_tagging.sql`)
- SQL queries in `db/queries/tags.sql` with full sqlc generation
- `src/internal/tags/service.go` — `Service`, `TagGroupDTO`, `TagDTO`, `TagInput`, and methods: `GetOrCreateDefaultGroups`, `GetUserTagGroups`, `GetUserTags`, `GetAlbumTagsByAlbumIds`, `GetAlbumTags`, `SetAlbumTags`
- `src/internal/tags/adapters/tags.templ` — `TagsModal`, `TagsFormWrapper`, `CloseTagsModal`, `tagsAlpineData` helper
- `src/internal/tags/adapters/http.go` — `HttpHandler` with `GetTagsModal` (GET) and `SubmitAlbumTags` (POST)
- `library.AlbumDTO` extended with `Tags []tags.TagDTO`; both `GetAlbumsInLibrary` (bulk) and `GetAlbumInLibrary` (single) populate tags
- `library.Service` constructor updated to accept `*tags.Service`
- `dashboard.templ` updated: Tags column header, `AlbumTagsCell` component, `GetAlbumTagsID` helper, Tags menu item in ellipsis dropdown
- `TagIcon` added to `src/internal/core/templates/icons.templ`
- `server.go` wired: `tags.NewService`, `library.NewService` updated, two new routes registered

Tag filtering of the albums table is deferred (out of scope for v1, per plan).

## Differences from the Plan

**`SetAlbumTags` implementation (simplified diff approach):** The plan described computing a diff (current vs. new, insert new rows, delete removed rows). The implementation instead deletes all `album_tags` for the album and re-inserts. This is simpler, safe under a transaction, and avoids a diff query. The end state is identical.

**`GetAlbumInLibrary` also populates tags:** The plan focused on `GetAlbumsInLibrary` (bulk). Since the tags modal handler uses `GetAlbumInLibrary` to fetch the album, that path also needed to populate `Tags`. This was added.

**Group assignment UX (per plan's risk note):** The modal shows the group selector as tab buttons that set `activeGroupId` in Alpine state. New tags typed while a group is active get assigned that group. Existing tags from autocomplete carry their stored group. A chip shows the group name in parentheses. This is simpler than per-chip dropdowns.

**`GetOrCreateDefaultGroups` called on every GET:** Each `GetTagsModal` call calls `GetOrCreateDefaultGroups`, which is two `INSERT OR IGNORE` + `SELECT` operations. This is safe and idempotent, matching the lazy-bootstrap plan decision.

## Plan Inaccuracies

- The plan placed `AlbumTagsCell` and `GetAlbumTagsID` in `tags/adapters`. They were moved to `library/adapters/dashboard.templ` — they are library view components, not tags module concerns. `tags/adapters/http.go` imports `library/adapters` to call `AlbumTagsCell` for the OOB swap after save.
- The plan described `AlbumTagsCell` taking `(album, isOobSwap bool)` — implemented exactly as described.
- The plan mentioned group assignment could be deferred: it was not deferred; the Alpine combobox supports group assignment via the group-selector tabs.
- `sqlc.embed(tag_groups)` in queries generated non-nullable `string` fields for `TagGroup` that could not scan SQL `NULL` from the LEFT JOIN. Fixed by replacing the embed with `COALESCE(tag_groups.id, '')` and `COALESCE(tag_groups.name, '')`, producing flat string fields that safely return `""` for ungrouped tags.
- Tag normalization (lowercase, strip disallowed characters) was added post-implementation in `service.go` via `normalizeTag`, applied in `SetAlbumTags` before any DB operation.
