# Physical Media Releases — Design Spec

**Date:** 2026-04-17
**Status:** Approved

## Overview

Allow users to record ownership of physical media releases (vinyl, CD, cassette) for albums already in their Spotify/Wax library, with optional Discogs integration to attach specific pressing details.

## Scope

- Physical media can only be added to albums already in the user's library via Spotify. Adding albums that exist only physically (no Spotify connection) is out of scope.
- Supported formats: vinyl, CD, cassette. Digital is displayed read-only (managed by Spotify sync).

## Data Model

### `releases` table — new optional columns

```sql
discogs_id   TEXT     -- Discogs release ID (e.g. "12345")
label        TEXT     -- pressing label, e.g. "Warner Bros. Records"
released_at  DATETIME -- pressing/release date
```

All three columns are nullable. A release can be owned with no Discogs data attached. The `releases` table represents the release artifact itself (shared across users who own the same pressing). `user_releases` is unchanged — it continues to track per-user ownership.

## UI Flow

### Entry point

The format icons group on the album detail page becomes a single tappable button. Tapping it opens the formats modal via an HTMX `hx-get`.

### Formats modal

The modal shows four format rows:

| Format   | Behaviour |
|----------|-----------|
| Digital  | Read-only, marked "via Spotify" |
| Vinyl    | Toggle on/off |
| CD       | Toggle on/off |
| Cassette | Toggle on/off |

When a physical format is toggled **on**, an inline section expands beneath it offering an optional Discogs search:

1. User types a query (album title, artist, catalog number, etc.)
2. Results appear inline as an HTMX fragment
3. User selects a pressing — `label` and `released_at` populate automatically
4. User can clear the Discogs selection without removing ownership

### Save behaviour

On modal save (`PUT /app/library/albums/{albumId}/formats`):

- Format toggled **ON**, not previously owned → create `release` (if not exists for format) + `user_release`
- Format toggled **OFF**, previously owned → soft-delete `user_release`
- Discogs details present → update `release` row with `discogs_id`, `label`, `released_at`
- No Discogs details → release saved with those fields null

The album detail format icons update in-place after save via an existing OOB HTMX swap.

## Routes

```
GET  /app/library/albums/{albumId}/formats
     Renders the formats modal

PUT  /app/library/albums/{albumId}/formats
     Saves ownership toggles and any Discogs data

GET  /app/library/albums/{albumId}/formats/{format}/discogs/search?q=...
     Searches Discogs, returns inline results fragment

GET  /app/library/albums/{albumId}/formats/{format}/discogs/releases/{discogsId}
     Fetches a specific Discogs release, returns prefilled details fragment
```

The last two routes are only called if the user chooses to search Discogs. The modal is fully functional without them.

## Module Structure

No new module. All changes are within the existing `library` module.

### Database
- One migration: add `discogs_id`, `label`, `released_at` to `releases`
- New SQLC queries:
  - Get releases + ownership state for an album + user
  - Upsert `user_release`
  - Soft-delete `user_release`
  - Update `release` Discogs fields

### Go
- **`library/adapters/formats.go`** — HTTP handlers for the 4 new routes
- **`library/adapters/formats_modal.templ`** — modal template and inline Discogs search/result fragments
- **`library/service.go`** — two new methods: `GetAlbumFormats`, `SaveAlbumFormats`

The `discogs` module is used as-is. The new library handlers call the existing `discogs.Service` for search and release fetch — no changes needed in the discogs module.

## Out of Scope

- Adding albums that don't exist in Spotify library
- Genre/tag suggestions from Discogs (existing CLI tool handles this separately)
- Editing Discogs details after initial save (can be addressed in a follow-up)
