# Sleeve Notes — Implementation

## What was built

A `notes` package (`src/internal/notes/`) providing per-user, per-album free-text notes stored in a new `album_notes` DB table. Notes are displayed on the album detail page with inline editing (click "Add"/"Edit" to load a textarea, cancel to revert, save to persist). Markdown is rendered to HTML via `goldmark` with the linkify extension, so URLs and `[text](url)` syntax become clickable links. The album list row shows a filled/outline notes icon indicating whether a note exists, linking to the detail page.

**Routes registered:**
- `GET /app/notes/album?albumId=X` — returns the inline editor component
- `GET /app/notes/album/view?albumId=X` — returns the read-only section (for cancel)
- `PUT /app/notes/album?albumId=X` — saves the note, returns updated section + OOB row swap

**Deferred:** No v1 filtering by "has notes". No live markdown preview in the editor.

## Differences from the plan

| Plan said | What was done | Why |
|---|---|---|
| `SleeveNotesSection` in `notes/adapters` | Moved to `library/adapters/sleeve_notes.templ` | Import cycle: `library/adapters` (album_detail) → `notes/adapters` → `library/adapters` (AlbumRowTagsSection). Moving display components to `library/adapters` breaks the cycle while keeping the editor in `notes/adapters`. |
| GET handler returns editor for inline swap (innerHTML target) | Implemented as specified, with a separate GET /view endpoint for cancel | The cancel button needs to restore the section wrapper (outerHTML), not just its contents, so a dedicated view endpoint is simpler than reconstructing the outer div client-side. |
| Character limit error re-renders editor with preserved content | On validation failure, content is passed from form but existing DB note is not pre-fetched | Avoids a DB round-trip on the error path; the textarea content comes from the submitted form value anyway. |

## Plan inaccuracies

- **Import cycle not anticipated:** The plan assumed `SleeveNotesSection` could live in `notes/adapters` and be imported by `album_detail.templ` (in `library/adapters`), while `notes/adapters/http.go` also imports `library/adapters`. This creates a cycle. The plan did not account for this constraint.
- **`GetSleeveNotesSectionID` placement:** Plan put this helper in `notes/adapters`; it ended up in `library/adapters/sleeve_notes.templ` alongside the section component.
- **No `CloseModal` needed:** The plan referenced a `CloseNotesModal()` pattern from research (which described a modal), but the final design uses inline editing with `hx-swap="innerHTML"` / `outerHTML` — no modal close component required.
