# Album Detail Page — Implementation

## What Was Built

- `GET /app/library/albums/{albumId}` route serving a full album detail page
- `AlbumDetailPage` templ component with cover art, title, Spotify link, artists, rating, notes, tags, release formats, added/last-played dates, and track list
- `AlbumDetailHeaderBar` with brand link, back-to-library icon button (tooltip only, no label), and user dropdown (no feeds dropdown)
- `LastPlayedAt` now populated inside `GetAlbumInLibrary` via `listeningHistoryService.GetLastPlayedAtByAlbumIds`
- Album title in `albumRow` changed from Spotify external link to internal detail page link; Spotify link moved to a small external-link icon beside the title
- `CollectionIcon` (`bi-collection` / `bi-collection-fill`) added to the shared icons template; used in both nav bars as the library navigation icon — filled on the dashboard (selected state), outline on the album detail page
- E2E tests: `e2e/feat/album_detail.feature` (9 scenarios) and `e2e/spec/album_detail.spec.ts`; all 9 tests verified passing
- `data-testid` attributes added to detail page elements and the dashboard album title link to support e2e selectors
- `E2E_TEST_ALBUM_ID` and `E2E_TEST_ALBUM_WITH_HISTORY_ID` documented in `.env.template`

All three modal OOB swap contracts are satisfied: `AlbumRating(album, false)`, `AlbumNotesIcon(album, false)`, and `AlbumTagsCell(album, false)` are all rendered on the detail page with the correct IDs.

## Differences from the Plan

- **Nav icon**: the plan called for reusing `DashboardHeaderBar`'s home icon pattern in the detail header. Post-implementation, both headers were updated to use a new `CollectionIcon` instead of `HomeIcon`, with filled/outline variants indicating selected state. The detail header also dropped the "Library" text label in favour of a tooltip.
- **E2E tests**: the plan listed 3 e2e scenarios in the Testing section; 9 were implemented, covering all scenarios from the Feature Specs Gherkin block.

## Plan Inaccuracies

- The `task build/templ` run reported `updates=0` but still generated `album_detail_templ.go` and updated `dashboard_templ.go`; this appears to be a display artifact, not a real failure.
- The exact whitespace in `dashboard.templ`'s `albumRow` anchor block differed slightly from what the plan implied (tab indentation), but the edit was applied correctly.
