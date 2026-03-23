# Sleeve Notes — Implementation Plan

## Approach

A dedicated `notes` package under `src/internal/notes/` following the same structure as `tags`. This keeps sleeve notes cleanly separated from the rating workflow. While adding to `review` was considered (and noted as an open question in the research), the structural difference — a single mutable record per user/album vs. an append-only log — makes a dedicated package the right call. It also avoids muddying the `AlbumRatingDTO` naming space where `Note` already means something different.

**UI placement:** The detail page gets an inline editable section (no modal), not a modal. Sleeve notes are likely long-form text — a cramped modal textarea is a poor writing surface. The inline approach is also consistent with how tags are displayed on the detail page (button opens modal), while notes can go one step further and be edited in-place since the detail page has the space. The list row shows only a `NotesIcon` indicator (filled when notes exist, outline when not) that links directly to the album detail page — no modal from the list row at all.

**Rich text:** Markdown rendered to HTML using `goldmark`. The primary use case is clickable links (references to reviews, discogs pages, etc.). The rendered HTML is displayed in read-only mode; the editor textarea still shows raw markdown. `goldmark` is the standard Go markdown library and is already used widely in the Go ecosystem — no custom renderer needed for v1.

**Character limit:** 10 000 characters. The rating note field caps at 2 000; sleeve notes are intentionally longer-form. 10 000 is generous without being unbounded.

**Filtering by "has notes":** Not in scope for v1. Can be added to `FilterParams` later.

## Files to Change

| File | Change |
|------|--------|
| `db/migrations/<timestamp>_album_notes.sql` | New migration: create `album_notes` table |
| `db/queries/album_notes.sql` | New queries: upsert, get by album, get by album IDs |
| `src/internal/notes/service.go` | New `Service` with `UpsertAlbumNote`, `GetAlbumNote`, `GetAlbumNotesByAlbumIds` |
| `src/internal/notes/adapters/http.go` | New HTTP handler: `GetSleeveNotesEditor`, `SaveSleeveNote` |
| `src/internal/notes/adapters/notes.templ` | New templ: `SleeveNotesSection`, `SleeveNotesEditor`, `SleeveNotesDisplay` (renders markdown HTML) |
| `src/internal/library/service.go` | Add `SleeveNote *notes.AlbumNoteDTO` to `AlbumDTO`; augment `GetAlbumInLibrary` and `GetAlbumsInLibrary` to fetch notes |
| `src/internal/library/adapters/dashboard.templ` | Add `GetAlbumNotesSectionID` helper; add notes indicator icon to `AlbumRowTagsSection` |
| `src/internal/library/adapters/album_detail.templ` | Add sleeve notes section below tags section |
| `src/internal/server/server.go` | Instantiate `notes.Service`; register `GET /app/notes/album` and `PUT /app/notes/album`; add `notes` service to `library.Service` constructor |

## Implementation Steps

1. **Create migration and run it.**
   `task db/create -- album_notes` then edit the generated file to create the `album_notes` table. Run `task db/up`.

2. **Add SQL queries.**
   Create `db/queries/album_notes.sql` with `UpsertAlbumNote`, `GetAlbumNote`, and `GetAlbumNotesByAlbumIds`. Run `task build/sqlc`.

3. **Implement `notes.Service`.**
   Create `src/internal/notes/service.go` with the `AlbumNoteDTO` type and service methods wrapping the generated sqlc queries. Add a `RenderMarkdown(content string) template.HTML` helper (using `goldmark`) that converts the stored markdown to sanitized HTML for display.

4. **Augment `library.AlbumDTO`.**
   Add `SleeveNote *notes.AlbumNoteDTO` to `AlbumDTO`. Update `GetAlbumInLibrary` to call `notesService.GetAlbumNote` and `GetAlbumsInLibrary` to call `notesService.GetAlbumNotesByAlbumIds`. Update `library.Service` constructor and `server.go` to wire in `notes.Service`.

5. **Build templ components.**
   Create `src/internal/notes/adapters/notes.templ` with three components:
   - `SleeveNotesSection(album, isOobSwap)` — the stable wrapper with a stable ID used as an OOB target
   - `SleeveNotesDisplay(album)` — read-only view with an "Edit" button that triggers `hx-get`
   - `SleeveNotesEditor(album)` — inline form with textarea (`hx-put`, `hx-swap="outerHTML"`, targets the section ID)

   Run `task build/templ`.

6. **Build HTTP handler.**
   Create `src/internal/notes/adapters/http.go` with:
   - `GET /app/notes/album?albumId=X` — returns `SleeveNotesEditor` for inline swap
   - `PUT /app/notes/album?albumId=X` — validates, upserts, returns `SleeveNotesSection` (OOB) closing the editor

7. **Add notes indicator to album list row.**
   In `dashboard.templ`: add `GetAlbumNotesSectionID` helper and a `NotesIcon` button at the end of `AlbumRowTagsSection`. Filled style when `album.SleeveNote != nil && album.SleeveNote.Content != ""`, outline otherwise. The button is a plain link to the album detail page (no modal).

8. **Add sleeve notes section to album detail page.**
   In `album_detail.templ`: add `SleeveNotesSection` call below the tags section.

9. **Register routes in `server.go`.**
   Add `notes.Service` construction, `library.Service` constructor update, and route registrations.

## Database Changes

**New table: `album_notes`**

```sql
CREATE TABLE IF NOT EXISTS album_notes (
    id         TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    content    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL DEFAULT current_timestamp,
    UNIQUE(user_id, album_id)
);
```

**Queries in `db/queries/album_notes.sql`:**

- `UpsertAlbumNote :one` — `INSERT INTO album_notes ... ON CONFLICT(user_id, album_id) DO UPDATE SET content = excluded.content, updated_at = current_timestamp`
- `GetAlbumNote :one` — select by `(user_id, album_id)`
- `GetAlbumNotesByAlbumIds :many` — select where `user_id = ? AND album_id IN (sqlc.slice('album_ids'))`

## Feature Specs

```gherkin
Feature: Sleeve Notes

  A user can write free-text notes about any album in their library,
  separate from the per-rating note. Notes persist indefinitely and
  can be edited at any time from the album detail page.

  Scenario: Writing a sleeve note for the first time
    Given the user is on an album detail page
    And no sleeve note exists for that album
    When the user clicks the notes edit button
    Then an inline editor appears with an empty textarea
    When the user types a note and saves
    Then the note is displayed on the detail page
    And the notes indicator icon in the list row becomes filled

  Scenario: Editing an existing sleeve note
    Given the user is on an album detail page
    And a sleeve note already exists for that album
    When the user clicks the notes edit button
    Then the inline editor appears pre-filled with the existing note
    When the user changes the text and saves
    Then the updated note is displayed on the detail page

  Scenario: Clearing a sleeve note
    Given a sleeve note exists for an album
    When the user opens the editor, clears the textarea, and saves
    Then the note section shows an empty state
    And the notes indicator icon in the list row becomes outline (unfilled)

  Scenario: List row notes indicator
    Given the user is on the library dashboard
    When an album has a sleeve note
    Then a filled notes icon is visible in that album's row
    When an album has no sleeve note
    Then an outline notes icon is visible in that album's row

  Scenario: Note exceeds character limit
    Given the user is editing a sleeve note
    When the user submits a note longer than 10 000 characters
    Then an inline error is displayed
    And the note is not saved
```

## Testing

**Unit tests (`src/internal/notes/service_test.go`):**
- `TestUpsertAlbumNote`: creates a note, then updates it, verifies the latest content is returned
- `TestGetAlbumNote`: returns `sql.ErrNoRows` (or nil DTO) when no note exists
- `TestGetAlbumNotesByAlbumIds`: bulk fetch returns correct map keyed by album ID; albums with no note are absent from map

**Handler-level (manual / E2E):**
- Write a note from the detail page and verify it persists on reload
- Verify the list row indicator changes state after saving a note
- Verify the character limit is enforced

**E2E spec file:** `e2e/feat/sleeve-notes.feature` — use the Gherkin scenarios above as the basis.

## Risks & Mitigations

- **`note` naming collision:** The `AlbumRatingDTO.Note` field already uses `Note`. The new DTO must be named `AlbumNoteDTO` and the `AlbumDTO` field must be `SleeveNote`, not `Note`, to avoid confusion. Variable names in templates and handlers should use `sleeveNote` consistently.
- **Extra DB round-trip in list view:** `GetAlbumsInLibrary` already makes 5+ queries. Adding one more bulk note fetch is consistent with existing patterns. If performance becomes a problem, the indicator can be derived from a JOIN on the existing album query rather than a separate fetch — but this is premature for v1.
- **OOB swap target stability:** The sleeve notes section on the detail page needs a stable element ID (`sleeve-notes-<albumId>`) so the PUT response can OOB-swap it. The list row icon is simpler — it is part of `AlbumRowTagsSection` which already has a stable ID and is already OOB-swapped after tag saves. After a note save, the handler should also re-render `AlbumRowTagsSection` as an OOB swap so the indicator updates without a page reload.
- **SQLite upsert:** Use `INSERT ... ON CONFLICT(user_id, album_id) DO UPDATE SET ...` — explicit and safe.
- **Character limit enforcement:** Enforce in the HTTP handler with a named constant (`const MaxSleeveNoteLength = 10_000`). Return an error fragment for display inline above the textarea.

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-plan sleeve-notes -->
