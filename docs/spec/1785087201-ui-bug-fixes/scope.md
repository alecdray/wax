# UI bug fixes

Fix four active UI bugs documented in `docs/backlog/bugs.md`.

---

## 1. Mobile: filter popovers can render off-screen

**Root cause.** Every filter/sort popover in `albums_list_frag.templ` is `absolute right-0` — right-anchored to its trigger button. On narrow viewports, buttons near the right edge of a horizontally-scrolling chip bar can push popovers past the viewport boundary.

**Shipped.** Changed `right-0` → `left-0` on all 5 popover divs (sort, rating, format, genre, artist). All popovers have fixed widths (w-56/w-64/w-72) that fit within any modern phone viewport from a left-anchor.

---

## 2. Library search bar: mid-request input loss

**Root cause.** The search `<input>` and the album table share the same `#album-list` swap target (`hx-swap="outerHTML"`). When a debounced request settles, HTMX replaces the entire `#album-list` region — including the input element — discarding characters buffered since the request was sent.

**Shipped.** Changed the input's `hx-target` from `#album-list` to `#album-list-results` and added `hx-select="#album-list-results"`. HTMX now extracts only the results subtree from the server response, leaving the input element untouched. Popover forms continue to target `#album-list` (they need the bar re-rendered with updated filter labels).

---

## 3. Review modal: back and exit buttons overlap

**Root cause.** The modal shell (`modal.templ`) renders `✕` as `absolute right-2 top-2`. The questionnaire form (`BaseQuestionsFormFrag`) renders its back arrow in a `flex justify-end` div at the very top of the form content — no clearance for the overlapping absolute button. On narrow viewports the two icons land on top of each other.

**Shipped.** Changed the header div class from `flex justify-end` to `flex items-center pr-8`. The back arrow moves to the left edge of the row; `pr-8` (2rem) provides clearance from the modal's absolute close button on the right.

---

## 4. Album detail: track list is out of order

**Root cause.** The `tracks` table has no ordering column and `GetAlbumTracksByAlbumId` has no `ORDER BY`, so SQLite returns tracks in insertion order.

**Shipped.**
1. Migration `20260726175307_add_track_position.sql` adds `disc_number INTEGER NOT NULL DEFAULT 1` and `track_number INTEGER NOT NULL DEFAULT 0` to `tracks`.
2. `GetAlbumTracksByAlbumId` and `GetAlbumTracksByAlbumIds` now `ORDER BY tracks.disc_number ASC, tracks.track_number ASC`.
3. `library/service.go` (`spotifyAlbumToDTO`) and `listeninghistory/service.go` both populate these fields from the Spotify `SimpleTrack`. The upsert updates positions on re-ingest so stale rows get corrected.
