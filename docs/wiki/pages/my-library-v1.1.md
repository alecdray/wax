---
description: >
  Planning page for My Library v1.1 — list view, faceted filtering and sorting,
  and dashboard UX cleanup.
links:
  - roadmap
  - features
  - data-model
---

[roadmap](./roadmap.md)

# My Library v1.1

Planning space for the next iteration of the core library experience.

## Current State (v1.0)

The library is a table-based view with:
- Sort by title, artist, rating, date added, last played (column header clicks)
- Infinite scroll in batches of 20
- Album title links to detail page; Spotify accessible via icon
- Rating control (badge if rated, "Rate" button if not) on each row
- Tags via ellipsis (⋯) dropdown
- Stats bar (artist, album, track counts) at the top
- Carousel above the table: Recently Spun / Unrated tabs

## Goals for v1.1

### 1. List View

Rework the table into a proper list view. The table format works but feels dense and spreadsheet-like — a list view leads with the visual experience of the collection, evoking physical media that had to speak through its cover art.

**Row layout:**
- **Left**: Album art thumbnail
- **Center**: Album name (prominent) on top, artist name underneath — merged into a single block
- **Right**: Rating number to one decimal place (e.g. `7.0`), large and in primary colors — the main feature of the app, treated as such; unrated albums show `--` as a placeholder
- **Footer**: Rating label shown as a tag badge alongside any other tags — decoupled from the number

**Interactions:**
- Clicking the rating number or `--` opens the rating modal
- Any click on a row or item within it navigates to the album detail page, unless the element has its own action (e.g. rating opens the modal)
- Date added, last played, and format are not shown as inline values — they exist as sortable facets only

**Scope**: One view, no mode-switching. Keep it simple.

### 2. Faceted Filtering & Sorting

Move beyond column-header sorting to a chip-based filter+sort UI. Users can slice their library by multiple attributes simultaneously.

**Facets in scope for v1.1:**
- Rating range (e.g., 7–10 only)
- Format (digital, vinyl, CD, cassette)
- Artist
- Unrated only / rated only

**Deferred facets (later):**
- Genre / tag — blocked on tag rework
- Date added
- Decade of release
- Recently spun / not recently spun

**Sorting in scope for v1.1:**
- Title (A–Z / Z–A)
- Artist (A–Z / Z–A)
- Rating (high to low / low to high)
- Date added (newest / oldest)
- Last played (most recent / oldest)

**Deferred sorts (later):**
- Release year

**UI approach:**
A row of chips, one per facet. Each chip can be activated or deactivated. Multiple chips can be active simultaneously for filtering; only one chip is active for sorting at a time. Active filters are visible and easily toggled off.

**State:**
- No session persistence — filters reset on page load
- Active filters and sort reflected in URL params, enabling future save/bookmark support
- Infinite scroll always applies; paginate the filtered set, never load all results at once

### 3. Dashboard UX Cleanup

- **Stats bar and carousel are retained** — no changes to the stats bar or the Recently Spun / Unrated carousel
- **Remove all Spotify outlinks from the dashboard** — Spotify is only accessible from the album detail page; no links from album name, artist name, or any dashboard column
- **Any row click → detail page** — the entire row is a link to the album detail page; specific elements with their own actions (rating number) are excepted

**Carousel edge case:** Recently Spun may surface albums not in the user's library, which would 404 on the detail page. These remain as Spotify outlinks for now. A dedicated roadmap item covers the album detail page enhancement and carousel fix.

---

## Brainstorm / Ideas (not this work)

Ideas worth capturing that are out of scope for v1.1:

- **Library Search** — a search box to find albums by title or artist; tracked on the roadmap as its own item
- **Hidden Albums / Soft Remove** — hide albums from the main view without removing them (e.g., podcasts, junk synced from Spotify); needs more thought, tracked on the roadmap
- **Spine view** — bookshelf-style layout showing record spines
- **Quick-rate** — hover or long-press a row to reveal a mini rating control without opening the modal
- **Bulk actions** — select multiple albums to bulk-tag, bulk-rate, or add to a shelf/ranklist
- **Sort by "last touched"** — last time the user interacted with an album in Wax, not just last played
- **Column customization** — let users choose which fields are visible

---

## Planning History

Key decisions and feedback that shaped this document during the initial planning review.

**List view:**
- The table format was too dense and spreadsheet-like; shifted to a visual list leading with album art to evoke the experience of physical media
- Rating is the core feature — it should be visually prominent with primary colors and a large number format
- Rating label decoupled from the number: shown as a tag badge in the row footer alongside other tags rather than tightly paired as in v1.0
- Rating always shown to one decimal place (`7.0`); unrated albums use `--` rather than a "Rate" button
- Date, last played, and format are facets only — not shown as inline row values
- Multiple view modes considered but deferred: build one view well rather than designing for mode-switching upfront

**Faceted filtering:**
- Genre/tag facet deferred — tags need a rework before they're a reliable filter surface
- Date added, decade of release, recently spun, and release year sort all deferred to a later pass
- Chip UI chosen over sidebar or dropdown bar: one chip per facet, multi-select for filtering, single-select for sorting
- Filters don't persist across sessions but are reflected in URL params as a foundation for future bookmarking

**Dashboard UX cleanup:**
- All Spotify outlinks removed from the dashboard; Spotify access moves exclusively to the album detail page
- Row click behavior made explicit: entire row navigates to detail page; only elements with their own actions (rating number) are excepted
- Carousel edge case noted: Recently Spun items without a library entry fall back to Spotify outlinks for now; a dedicated roadmap item covers the album detail page fix

**Scope reductions:**
- Physical Media moved to its own roadmap item — too complex to bundle into a library redesign
- Multiple View Modes removed from the roadmap queue; captured as an idea only
- Library Search and Hidden Albums / Soft Remove each split off as their own roadmap items
- Album Detail — Non-Library Albums added to the roadmap to cover the carousel edge case
