# Album Detail Page — Implementation Plan

## Approach

Add a full-page route at `GET /app/library/albums/{albumId}` that renders all available album data using the existing `GetAlbumInLibrary` service method. The page wraps in `RootComponent` (same pattern as `DashboardPage`) so the global `ModalContainer` is included, meaning all three modal flows — rating, notes, and tags — work without any changes to those handlers.

The `AlbumTagsCell` OOB swap will work on the detail page because `SubmitAlbumTags` renders `AlbumTagsCell` with `isOobSwap=true` using the shared `GetAlbumTagsID` function. The detail page just needs to include `AlbumTagsCell(album, false)` with the matching ID.

`LastPlayedAt` is not populated in `GetAlbumInLibrary`. Rather than patching the handler locally, fix it inside the service so both call paths (single and bulk) are consistent. This requires calling `listeningHistoryService.GetLastPlayedAtByAlbumIds` with the single album ID.

The album title in the dashboard `albumRow` currently links externally to Spotify. The primary entry point for the detail page will be changing this to an internal link, with the Spotify link moved to a separate icon/button on the detail page itself.

A simple page-level header (brand link + back-to-dashboard link + user dropdown) is created for the detail page rather than reusing `DashboardHeaderBar`, which carries feeds-dropdown logic specific to the dashboard.

Inline editing (rating, notes, tags) stays modal-based in v1 — the same buttons/triggers as on the dashboard row are used on the detail page.

## Files to Change

| File | Change |
|------|--------|
| `src/internal/library/service.go` | Populate `LastPlayedAt` inside `GetAlbumInLibrary` via `listeningHistoryService.GetLastPlayedAtByAlbumIds` |
| `src/internal/library/adapters/http.go` | Add `GetAlbumDetailPage` handler |
| `src/internal/library/adapters/album_detail.templ` | New file: `AlbumDetailPage` component and a minimal `AlbumDetailHeaderBar` |
| `src/internal/library/adapters/dashboard.templ` | Change `albumRow` album-title link from Spotify external to internal detail page URL |
| `src/internal/server/server.go` | Register `GET /app/library/albums/{albumId}` route |

## Implementation Steps

1. **Fix `LastPlayedAt` in `GetAlbumInLibrary`** — call `listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, []string{albumId})` at the end of `GetAlbumInLibrary` and set `albumDto.LastPlayedAt` from the result. No DB migration required.

2. **Add `GetAlbumDetailPage` handler** — in `src/internal/library/adapters/http.go`, extract `albumId` from `r.PathValue("albumId")`, call `GetAlbumInLibrary`, 404 on "album not in library", otherwise render `AlbumDetailPage`.

3. **Create `album_detail.templ`** — implement `AlbumDetailHeaderBar` (brand link, back link, user dropdown) and `AlbumDetailPage`. The page renders: cover art, title with Spotify link icon, artist list with Spotify links, release format badges with added-at dates, `AlbumRating`, `AlbumNotesIcon`, `AlbumTagsCell`, last-played date, and track list. All three components must use their respective ID-returning helpers (`GetAlbumRatingID`, `GetAlbumNotesID`, `GetAlbumTagsID`) so OOB swaps land correctly.

4. **Update `albumRow` entry point** — change the album title anchor in `dashboard.templ` from `https://open.spotify.com/album/{spotifyId}` to the internal `/app/library/albums/{albumId}`. Add a small external-link icon next to the title as the Spotify link.

5. **Register route** — add `appMux.Handle("GET /app/library/albums/{albumId}", httpx.HandlerFunc(libraryHandler.GetAlbumDetailPage))` in `server.go`.

6. **Run `task build/templ`** after creating/modifying `.templ` files.

## Database Changes

None. All data is already fetched by `GetAlbumInLibrary` once `LastPlayedAt` is wired in (step 1).

## Feature Specs

```gherkin
Feature: Album Detail Page

  A dedicated page for a single album in the user's library, surfacing all
  album metadata, ratings, notes, tags, and release formats in a mobile-friendly layout.

  Scenario: Viewing an album in the library
    Given a logged-in user with at least one album in their library
    When they navigate to /app/library/albums/{albumId}
    Then they see the album cover art, title, and artist names
    And they see the release format badges with added-at dates
    And they see the album rating (or a Rate button if unrated)
    And they see the album tags (or an empty state if none)

  Scenario: Navigating to the detail page from the dashboard
    Given a logged-in user on the dashboard
    When they click an album title in the library table
    Then they are taken to that album's detail page

  Scenario: Rating an album from the detail page
    Given a logged-in user on an album detail page
    When they click the Rate button
    Then the rating modal opens
    And after submitting, the rating display on the detail page updates

  Scenario: Editing notes from the detail page
    Given a logged-in user on an album detail page
    When they click the notes button
    Then the notes modal opens
    And after saving, the notes icon on the detail page updates

  Scenario: Editing tags from the detail page
    Given a logged-in user on an album detail page
    When they click the tags button
    Then the tags modal opens
    And after saving, the tags display on the detail page updates

  Scenario: Accessing an album not in the library
    Given a logged-in user
    When they navigate to /app/library/albums/{albumId} for an album not in their library
    Then they receive a 404 error

  Scenario: Last played date is shown when available
    Given a logged-in user on an album detail page for an album with listening history
    Then the last played date is displayed

  Scenario: Last played date is absent when not available
    Given a logged-in user on an album detail page for an album with no listening history
    Then no last played date is shown

  Scenario: Back navigation to the dashboard
    Given a logged-in user on an album detail page
    When they click the back / dashboard link in the header
    Then they are returned to the library dashboard
```

## Testing

**Manual verification:**
- Navigate to the dashboard, click an album title, confirm the detail page loads with correct data.
- Open the rating modal, submit, confirm the rating badge updates on the detail page via OOB swap.
- Open the notes modal, save a note, confirm the notes icon updates.
- Open the tags modal, save tags, confirm the tags cell updates.
- Check that `LastPlayedAt` is populated for albums with listening history.
- Test on a narrow viewport to confirm the mobile-first layout.
- Navigate to an albumId not in the library and confirm a 404 response.

**E2E feature file (`e2e/feat/album_detail.feature`):**
- Scenario: detail page loads for an album in the library.
- Scenario: album title link on the dashboard navigates to the detail page.
- Scenario: 404 for an album not in the user's library.

**Unit tests:**
- `GetAlbumInLibrary` now populates `LastPlayedAt` — test that the field is set when listening history exists and nil when it does not.

## Risks & Mitigations

1. **OOB swap contract** — if any of the three required element IDs (`album-rating-{id}`, `album-notes-{id}`, `album-tags-{id}`) are absent from the detail page, post-modal updates will silently fail. Mitigate by including all three components and writing a smoke test that verifies the IDs are present in the rendered HTML.

2. **`LastPlayedAt` service change** — touching `GetAlbumInLibrary` adds an extra DB call for every single-album fetch. The call is a keyed lookup by a single albumId so it is cheap, but it is a new dependency. If `listeningHistoryService` is ever slow or unavailable, it now affects the detail page. Acceptable for v1.

3. **`albumRow` link change** — converting the album title from an external Spotify link to an internal link is a UX change to the dashboard. Users who relied on clicking the title to open Spotify will need to use the new Spotify icon instead. The Spotify link must remain prominently accessible on both the row and the detail page to avoid regression.

4. **Mobile layout** — the roadmap names this as a mobile-usability improvement. The template should avoid `table` or `hidden md:` patterns used on the dashboard. Using a card-style or stacked layout for releases, rating, and tags is the correct approach for v1.

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-plan album-detail-page -->
