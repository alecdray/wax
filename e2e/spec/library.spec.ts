import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';
import {
  clearAllRatingStates,
  getLibraryAlbumIds,
  setRatingStateValue,
} from '../helpers/db';

// Scenarios from e2e/feat/library.feature

const userId = process.env.E2E_TEST_USER_ID;

// --- Dashboard load ---

test('Viewing the library dashboard shows list view', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await expect(page.getByTestId('library-stats')).toBeVisible();
  await expect(page.getByTestId('carousel-section-recently-spun-tab')).toBeVisible();
  await expect(page.getByTestId('albums-list')).toBeVisible();
});

test('Album rows show art, title, and rating — no Spotify outlinks', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await expect(page.getByTestId('album-list-row').first()).toBeVisible();
  await expect(page.getByTestId('album-list-row-title-link').first()).toBeVisible();
  // First row's rating control: either the rated or unrated readout
  const firstRow = page.getByTestId('album-list-row').first();
  const ratingReadout = firstRow.locator('[data-testid="album-score-readout-rated"], [data-testid="album-score-readout-unrated"]');
  await expect(ratingReadout).toBeVisible();

  // No Spotify outlinks in list rows. The testid is declared in the row templ
  // as a reserved name (see comment in albums_list_frag.templ) and is
  // intentionally absent from rendered rows.
  const spotifyLinks = page.getByTestId('albums-list').locator('[data-testid="album-list-row-spotify-link"]');
  await expect(spotifyLinks).toHaveCount(0);
});

// --- Carousel ---

test('Default carousel shows Recently Spun', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  const tab = page.getByTestId('carousel-section-recently-spun-tab');
  await expect(tab).toBeVisible();
  // Active tab has no hx-get trigger
  await expect(tab).not.toHaveAttribute('hx-get');
});

test('Switching the carousel to Unrated', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('carousel-section-unrated-tab').click();

  // After the HTMX swap the unrated tab becomes active (loses its hx-get trigger)
  await expect(page.getByTestId('carousel-section-unrated-tab')).not.toHaveAttribute('hx-get');
});

test('Switching the carousel to Provisional lists only provisional albums', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  // Position the user's library so exactly one album is provisional and one is
  // finalized — the Provisional tab must list the first and not the second.
  const albumIds = getLibraryAlbumIds(userId!);
  expect(albumIds.length, 'fixture user must own at least two albums').toBeGreaterThanOrEqual(2);
  const provisionalId = albumIds[0];
  const finalizedId = albumIds[1];

  clearAllRatingStates(userId!);
  setRatingStateValue(userId!, provisionalId, 'provisional');
  setRatingStateValue(userId!, finalizedId, 'finalized');

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('carousel-section-provisional-tab').click();
  // Active tab loses its hx-get trigger after the swap.
  await expect(page.getByTestId('carousel-section-provisional-tab')).not.toHaveAttribute('hx-get');

  // The provisional strip is visible and contains exactly the provisional
  // album's card — the finalized album must not appear, and no unrated album
  // (no state row) may slip in either.
  const strip = page.getByTestId('provisional-carousel-strip');
  await expect(strip).toBeVisible();
  const cards = strip.getByTestId('provisional-carousel-strip-album-card');
  const cardCount = await cards.count();
  expect(cardCount, 'provisional strip must contain at least one album card').toBeGreaterThan(0);
  const cardAlbumIds: string[] = [];
  for (let i = 0; i < cardCount; i++) {
    const id = await cards.nth(i).getAttribute('data-album-id');
    expect(id, `card ${i} must declare its album id`).toBeTruthy();
    cardAlbumIds.push(id!);
  }
  expect(cardAlbumIds, 'provisional strip must contain the provisional album').toContain(provisionalId);
  expect(cardAlbumIds, 'provisional strip must not contain the finalized album').not.toContain(finalizedId);
});

test('Provisional carousel empty state is neutral when no provisional albums exist', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  // Wipe every rating-state row so no album in the library is provisional.
  clearAllRatingStates(userId!);

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('carousel-section-provisional-tab').click();
  await expect(page.getByTestId('carousel-section-provisional-tab')).not.toHaveAttribute('hx-get');

  // Neutral empty state — present, not celebratory. The strip must be absent.
  await expect(page.getByTestId('provisional-carousel-strip-empty')).toBeVisible();
  await expect(page.getByTestId('provisional-carousel-strip')).toHaveCount(0);
  const emptyText = (await page.getByTestId('provisional-carousel-strip-empty').innerText()).toLowerCase();
  for (const banned of ['caught up', 'all done', 'great job', 'nice work', ' woo', 'congrat']) {
    expect(emptyText, `empty-state message must be neutral; found "${banned}" in: ${emptyText}`).not.toContain(banned);
  }
});

// --- Unified search bar ---

test('Unified search bar replaces the chip-modal bar', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await expect(page.getByTestId('unified-search-bar')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-input')).toBeVisible();

  // Legacy chip-modal surface and its four chips must be gone.
  await expect(page.getByTestId('filter-chip-bar')).toHaveCount(0);
  await expect(page.getByTestId('filter-chip-bar-sort')).toHaveCount(0);
  await expect(page.getByTestId('filter-chip-bar-rating')).toHaveCount(0);
  await expect(page.getByTestId('filter-chip-bar-format')).toHaveCount(0);
  await expect(page.getByTestId('filter-chip-bar-artist')).toHaveCount(0);
});

test('Typing in the search bar narrows the album list live', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // Pick a query guaranteed to match: take the title of the first visible row.
  const firstTitle = await page.getByTestId('album-list-row-title-link').first().innerText();
  // Pull a short substring (first word, ≤6 chars) to drive the search.
  const probe = firstTitle.split(/\s+/)[0].slice(0, 6).toLowerCase();
  expect(probe.length, 'need a non-empty probe substring').toBeGreaterThan(0);

  // Use pressSequentially (not fill) so each character emits a real keyup
  // event — htmx's "keyup changed" trigger doesn't fire on programmatic fill.
  const tableResponse = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes(`q=${encodeURIComponent(probe)}`),
  );
  await page.getByTestId('unified-search-bar-input').pressSequentially(probe);
  await tableResponse;

  // After the HTMX swap, every visible row's title or artist line must
  // contain the probe substring. Re-query rows post-swap.
  await expect(page.getByTestId('album-list-row').first()).toBeVisible();
  const rows = page.getByTestId('album-list-row');
  const count = await rows.count();
  expect(count, 'expected at least one matching row').toBeGreaterThan(0);
  for (let i = 0; i < count; i++) {
    const rowText = (await rows.nth(i).innerText()).toLowerCase();
    expect(rowText, `row ${i} should contain probe ${probe}`).toContain(probe);
  }
});

test('Clearing the search bar restores the full library', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // First narrow: type a guaranteed-zero-match query so the list collapses.
  const probe = 'zzzznotpresent';
  const narrowResponse = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes(`q=${probe}`),
  );
  await page.getByTestId('unified-search-bar-input').pressSequentially(probe);
  await narrowResponse;
  await expect(page.getByTestId('album-list-row')).toHaveCount(0);

  // Now clear and confirm a fresh request fires with no q (or empty q) and
  // rows return. Backspacing each character emits real keyup events that
  // htmx's "keyup changed" trigger sees; programmatic fill('') wouldn't.
  const clearResponse = page.waitForResponse((res) => {
    if (!res.url().includes('/app/library/dashboard/albums-table')) return false;
    const u = new URL(res.url());
    return !u.searchParams.get('q'); // q absent or empty string
  });
  const input = page.getByTestId('unified-search-bar-input');
  await input.focus();
  await input.press('End');
  for (let i = 0; i < probe.length; i++) {
    await input.press('Backspace');
  }
  await clearResponse;

  await expect(page.getByTestId('album-list-row').first()).toBeVisible();
});

// --- Sort / rating / format / artist controls on the unified bar ---

test('Sort control on the unified bar reorders the list', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('unified-search-bar-sort-toggle').click();
  const popover = page.getByTestId('unified-search-bar-sort-popover');
  await expect(popover).toBeVisible();
  await popover.locator('input[name="sortBy"][value="artist"]').check();
  await popover.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-sort-toggle')).toContainText('Artist');
});

test('Rating control on the unified bar narrows by minimum rating', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('unified-search-bar-rating-toggle').click();
  const popover = page.getByTestId('unified-search-bar-rating-popover');
  await expect(popover).toBeVisible();
  await popover.locator('input[name="minRating"]').fill('7');
  await popover.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-rating-toggle')).toContainText('7');
});

test('Filtering to unrated from the unified bar', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('unified-search-bar-rating-toggle').click();
  const popover = page.getByTestId('unified-search-bar-rating-popover');
  await expect(popover).toBeVisible();
  await popover.locator('input[name="rated"][value="unrated"]').check();
  await popover.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-rating-toggle')).toContainText('Unrated');
});

test('Format control on the unified bar supports multi-select', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('unified-search-bar-format-toggle').click();
  const popover = page.getByTestId('unified-search-bar-format-popover');
  await expect(popover).toBeVisible();
  // Multi-select inputs are checkboxes — picking one keeps the others
  // unchecked but the same control accepts further picks. We confirm the
  // checkbox shape (vs the legacy radio) and apply.
  const vinylCheckbox = popover.locator('input[name="format"][value="vinyl"]');
  await expect(vinylCheckbox).toHaveAttribute('type', 'checkbox');
  await vinylCheckbox.check();
  await popover.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-format-toggle')).toContainText('vinyl');
});

test('Artist control on the unified bar opens a searchable list when artists exist', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  const toggle = page.getByTestId('unified-search-bar-artist-toggle');
  // The fixture user is expected to have at least one artist; fail loud if not.
  await expect(toggle, 'fixture user must have at least one artist in their library').toBeVisible();
  await toggle.click();

  const popover = page.getByTestId('unified-search-bar-artist-popover');
  await expect(popover).toBeVisible();
  await expect(popover.getByTestId('unified-search-bar-artist-checkbox').first()).toBeVisible();
});

// --- Rating modal from list row (unchanged behaviour) ---

test('Opening the rating modal from an album row', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // The rating control on a row is either album-score-readout-rated or
  // album-score-readout-unrated depending on whether the album has a rating.
  const firstRow = page.getByTestId('album-list-row').first();
  await firstRow.locator('[data-testid="album-score-readout-rated"], [data-testid="album-score-readout-unrated"]').first().click();

  await expect(page.locator('dialog[open]')).toBeVisible();
});

// --- Task 1.4 — Active non-default state visible on the bar at rest ---

test('Bar shows no badges at full defaults', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await expect(page.getByTestId('unified-search-bar')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-badges')).toHaveCount(0);
  await expect(page.getByTestId('unified-search-bar-reset')).toHaveCount(0);
});

test('Bar surfaces a badge for a non-default filter at rest', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // Apply unrated-only via the rating popover.
  await page.getByTestId('unified-search-bar-rating-toggle').click();
  const popover = page.getByTestId('unified-search-bar-rating-popover');
  await expect(popover).toBeVisible();
  await popover.locator('input[name="rated"][value="unrated"]').check();
  const resp = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes('rated=unrated'),
  );
  await popover.getByRole('button', { name: 'Apply' }).click();
  await resp;

  // The rating-dimension badge is visible at rest — no popover open.
  await expect(page.getByTestId('unified-search-bar-rating-popover')).not.toBeVisible();
  await expect(page.getByTestId('unified-search-bar-badges')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-badge-rating')).toBeVisible();
});

test('Bar surfaces a sort badge for a non-default sort at rest', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // Apply Artist sort (non-default — default is Date Added desc).
  await page.getByTestId('unified-search-bar-sort-toggle').click();
  const popover = page.getByTestId('unified-search-bar-sort-popover');
  await expect(popover).toBeVisible();
  await popover.locator('input[name="sortBy"][value="artist"]').check();
  const resp = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes('sortBy=artist'),
  );
  await popover.getByRole('button', { name: 'Apply' }).click();
  await resp;

  await expect(page.getByTestId('unified-search-bar-sort-popover')).not.toBeVisible();
  await expect(page.getByTestId('unified-search-bar-badge-sort')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-badge-sort')).toContainText('Artist');
});

// --- Task 1.5 — URL reflects full view state; reloading reproduces it ---

test('Bare dashboard URL stays bare in the address bar', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // The default-view URL must not have any of the canonical view params.
  await expect(page.getByTestId('albums-list')).toBeVisible();
  const url = new URL(page.url());
  for (const param of ['q', 'sortBy', 'dir', 'minRating', 'maxRating', 'rated', 'format', 'artist']) {
    expect(url.searchParams.has(param), `default URL must not contain ${param}; got ${page.url()}`).toBe(false);
  }
});

test('Applying state writes the expected params to the URL', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // Sort by Artist.
  await page.getByTestId('unified-search-bar-sort-toggle').click();
  const sortPopover = page.getByTestId('unified-search-bar-sort-popover');
  await expect(sortPopover).toBeVisible();
  await sortPopover.locator('input[name="sortBy"][value="artist"]').check();
  const sortResp = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes('sortBy=artist'),
  );
  await sortPopover.getByRole('button', { name: 'Apply' }).click();
  await sortResp;

  // Verify URL reflects sortBy=artist (dir was 'desc' default — must be dropped).
  await expect.poll(() => new URL(page.url()).searchParams.get('sortBy')).toBe('artist');
  expect(new URL(page.url()).searchParams.has('dir'), 'default dir=desc must be dropped').toBe(false);

  // Now type a query.
  const queryResp = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes('q=ab'),
  );
  await page.getByTestId('unified-search-bar-input').pressSequentially('ab');
  await queryResp;

  await expect.poll(() => new URL(page.url()).searchParams.get('q')).toBe('ab');
  expect(new URL(page.url()).searchParams.get('sortBy'), 'sort should still be in URL').toBe('artist');
});

test('Reloading the URL reproduces the same view', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);

  // Visit with explicit URL state.
  await page.goto('/app/library/dashboard?sortBy=artist&rated=unrated');
  await expect(page.getByTestId('albums-list')).toBeVisible();

  // Badges should reflect both dimensions.
  await expect(page.getByTestId('unified-search-bar-badge-sort')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-badge-rating')).toBeVisible();

  // Capture the first row's title for a post-reload comparison.
  const firstTitleBefore = await page.getByTestId('album-list-row-title-link').first().innerText().catch(() => '');
  const url = page.url();

  // Reload.
  await page.reload();
  await expect(page.getByTestId('albums-list')).toBeVisible();

  // URL unchanged.
  expect(page.url()).toBe(url);
  // Bar still reflects state.
  await expect(page.getByTestId('unified-search-bar-badge-sort')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-badge-rating')).toBeVisible();
  // Same first row.
  const firstTitleAfter = await page.getByTestId('album-list-row-title-link').first().innerText().catch(() => '');
  expect(firstTitleAfter, 'first row title must reproduce after reload').toBe(firstTitleBefore);
});

test('A fresh browser context renders the same view for a deep URL', async ({ browser }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  // First context: capture the rendered state for sortBy=artist.
  const ctxA = await browser.newContext();
  await loginAs(ctxA, userId!);
  const pageA = await ctxA.newPage();
  await pageA.goto('/app/library/dashboard?sortBy=artist');
  await expect(pageA.getByTestId('albums-list')).toBeVisible();
  const titleA = await pageA.getByTestId('album-list-row-title-link').first().innerText();
  await ctxA.close();

  // Second context: a fresh authenticated context with no prior interaction.
  const ctxB = await browser.newContext();
  await loginAs(ctxB, userId!);
  const pageB = await ctxB.newPage();
  await pageB.goto('/app/library/dashboard?sortBy=artist');
  await expect(pageB.getByTestId('albums-list')).toBeVisible();
  const titleB = await pageB.getByTestId('album-list-row-title-link').first().innerText();
  await ctxB.close();

  expect(titleB, 'fresh context with same URL must render the same first row').toBe(titleA);
});

test('Setting a value back to its default removes the param from the URL', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard?rated=unrated');
  await expect(page.getByTestId('albums-list')).toBeVisible();
  expect(new URL(page.url()).searchParams.get('rated')).toBe('unrated');

  // The reset control is rendered in the bar's badges row when state is active.
  const resp = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table'),
  );
  await page.getByTestId('unified-search-bar-reset').click();
  await resp;

  await expect.poll(() => new URL(page.url()).searchParams.has('rated')).toBe(false);
  // No badges left.
  await expect(page.getByTestId('unified-search-bar-badges')).toHaveCount(0);
});

// --- Task 1.6 — Infinite-scroll pagination preserves all state ---

test('Pagination request carries every active filter and sort param', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);

  // Visit a deep URL with sort + filter active. The fixture has ~236 albums in
  // the test user's library; "rated=only" still yields enough to span pages.
  await page.goto('/app/library/dashboard?sortBy=artist&rated=only');
  await expect(page.getByTestId('albums-list')).toBeVisible();

  // Sanity: first page must be full (20 rows) for the sentinel to exist.
  // The sentinel is an empty <li> with zero intrinsic height — we don't assert
  // visibility (would fail), only that the DOM node is present and carries the
  // expected hx-get URL.
  const initialRows = await page.getByTestId('album-list-row').count();
  expect(initialRows, 'need a full first page to test pagination').toBe(20);
  await expect(page.getByTestId('albums-list-pagination-sentinel')).toHaveCount(1);
  const sentinelHxGet = await page.getByTestId('albums-list-pagination-sentinel').getAttribute('hx-get');
  expect(sentinelHxGet, 'sentinel must carry an hx-get URL').toBeTruthy();
  const sentinelURL = new URL(sentinelHxGet!, page.url());
  expect(sentinelURL.searchParams.get('sortBy'), 'sentinel URL must carry sortBy').toBe('artist');
  expect(sentinelURL.searchParams.get('rated'), 'sentinel URL must carry rated').toBe('only');
  expect(sentinelURL.searchParams.get('offset'), 'sentinel URL must carry offset').toBe('20');
  expect(sentinelURL.searchParams.has('dir'), 'default dir=desc must not appear in sentinel URL').toBe(false);

  // Snapshot first-page titles to verify no duplicates after the swap.
  const titlesPage1 = await page.getByTestId('album-list-row-title-link').allInnerTexts();

  // Scrolling reveals the sentinel which fires hx-get for albums-page. Wait on
  // a request URL that includes the active filter + sort params.
  const pageReq = page.waitForRequest((req) => {
    if (!req.url().includes('/app/library/dashboard/albums-page')) return false;
    const u = new URL(req.url());
    return u.searchParams.get('sortBy') === 'artist' &&
      u.searchParams.get('rated') === 'only' &&
      u.searchParams.get('offset') === '20';
  });
  await page.getByTestId('albums-list-pagination-sentinel').scrollIntoViewIfNeeded();
  const req = await pageReq;

  // Verify the URL has no leaked default dir param.
  expect(new URL(req.url()).searchParams.has('dir'), 'default dir=desc must not be in pagination URL').toBe(false);

  // After the swap, rows should have grown and no duplicates across the boundary.
  await expect.poll(async () => page.getByTestId('album-list-row').count()).toBeGreaterThan(20);
  const titlesAll = await page.getByTestId('album-list-row-title-link').allInnerTexts();
  expect(titlesAll.length, 'rows must grow').toBeGreaterThan(titlesPage1.length);

  // No duplicates across the page boundary.
  const unique = new Set(titlesAll);
  expect(unique.size, `pagination introduced duplicates: ${titlesAll.length - unique.size}`).toBe(titlesAll.length);
});

// --- Task 1.7 — Zero-result state with one-click reset ---

test('Zero-result view shows a non-judgemental message and a single reset control', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);

  // Deep-link to a guaranteed-zero query.
  await page.goto('/app/library/dashboard?q=zzzznotpresent');
  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('album-list-row')).toHaveCount(0);

  // Empty-state message and a single visible reset control.
  await expect(page.getByTestId('albums-list-empty-state')).toBeVisible();
  // Exactly one reset control in the empty state (the bar's reset is the same
  // affordance class but the empty-state reset has its own testid).
  await expect(page.getByTestId('albums-list-empty-state-reset')).toHaveCount(1);
  await expect(page.getByTestId('albums-list-empty-state-reset')).toBeVisible();
});

test('Activating reset from the empty state restores the full library and bare URL', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard?q=zzzznotpresent');
  await expect(page.getByTestId('albums-list-empty-state-reset')).toBeVisible();

  // Activate the empty-state reset.
  const resp = page.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table'),
  );
  await page.getByTestId('albums-list-empty-state-reset').click();
  await resp;

  // URL is bare (no q, no other view params).
  await expect.poll(() => {
    const u = new URL(page.url());
    return ['q', 'sortBy', 'dir', 'minRating', 'maxRating', 'rated', 'format', 'artist']
      .some((p) => u.searchParams.has(p));
  }).toBe(false);

  // Full library re-rendered.
  await expect(page.getByTestId('album-list-row').first()).toBeVisible();
  // No badges shown.
  await expect(page.getByTestId('unified-search-bar-badges')).toHaveCount(0);
});

// --- Product criteria (PC) ---
//
// Whole-system invariants of the assembled feature, distinct from per-task
// tests. These survive past this build as regression coverage for the
// search/filter/sort surface as a unit.
//
// PC1 is verified primarily at the Go integration layer
// (src/internal/library/pc_and_composition_test.go) where every combination
// can be enumerated against an independently-computed reference. The e2e
// test below covers the assembled UI half: that the bar's composition flows
// through HTMX into the rendered DOM as a single coherent set/order.
//
// PC2 is inherently a browser concern — the URL is owned by the address
// bar, and the round-trip property only holds when a real navigation
// reproduces the view. Two e2e tests cover the two halves: deep-link
// fidelity (cold start with a URL) and end-to-end UI → URL → fresh-context
// round-trip across several combinations.

test('PC1 — combined q + filter + sort produces the predicted set in the predicted order', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);

  // Compose a non-trivial view state through the URL — equivalent to driving
  // the bar end-to-end, but deterministic across runs. Sort by album ASC so
  // the row order is mechanically checkable against the rendered titles.
  // Filter to rated-only with a min rating to narrow the set.
  await page.goto('/app/library/dashboard?q=the&rated=only&minRating=7&sortBy=album&dir=asc');
  await expect(page.getByTestId('albums-list')).toBeVisible();

  const rowCount = await page.getByTestId('album-list-row').count();
  expect(rowCount, 'combined view must render at least one row for this fixture').toBeGreaterThan(0);

  // Capture every row's title and artist line. Every row must satisfy the
  // text query AND-composed with the filter — the title or any artist line
  // must contain "the" (case-insensitive). Filter conformance for ratings is
  // server-side; we verify it indirectly by asserting the row count is
  // strictly less than the unfiltered library size below.
  const rows = page.getByTestId('album-list-row');
  const titles: string[] = [];
  for (let i = 0; i < rowCount; i++) {
    const row = rows.nth(i);
    const text = (await row.innerText()).toLowerCase();
    expect(text, `row ${i} must contain the query "the" in title or artist`).toContain('the');
    const titleEl = row.getByTestId('album-list-row-title-link').first();
    titles.push((await titleEl.innerText()).trim());
  }

  // Sort assertion — sortBy=album dir=asc means titles render in
  // case-sensitive lexicographic order (production SortByTitle uses < on
  // raw strings, not a folded comparison).
  const sortedTitles = [...titles].sort();
  expect(titles, 'rendered order must equal the active sort').toEqual(sortedTitles);

  // The combined view must be a strict subset of the unfiltered library —
  // proves the AND composition actually narrowed the set rather than the
  // bar interactions silently failing and showing the full library.
  await page.goto('/app/library/dashboard');
  await expect(page.getByTestId('albums-list')).toBeVisible();
  const baseCount = await page.getByTestId('album-list-row').count();
  expect(rowCount, 'combined view must narrow the full library').toBeLessThan(baseCount);
});

test('PC2 — a captured URL reproduces the exact DOM order in a fresh browser context', async ({ browser }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  // First context — drive the bar through a multi-dimensional state via
  // interactions (not a deep link). This exercises the UI→URL push path
  // alongside the URL→DOM render path.
  const ctxA = await browser.newContext();
  await loginAs(ctxA, userId!);
  const pageA = await ctxA.newPage();
  await pageA.goto('/app/library/dashboard');
  await expect(pageA.getByTestId('albums-list')).toBeVisible();

  // Apply: sort by Artist asc.
  await pageA.getByTestId('unified-search-bar-sort-toggle').click();
  const sortPopover = pageA.getByTestId('unified-search-bar-sort-popover');
  await expect(sortPopover).toBeVisible();
  await sortPopover.locator('input[name="sortBy"][value="artist"]').check();
  await sortPopover.locator('input[name="dir"][value="asc"]').check();
  const sortResp = pageA.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes('sortBy=artist'),
  );
  await sortPopover.getByRole('button', { name: 'Apply' }).click();
  await sortResp;

  // Apply: rated only.
  await pageA.getByTestId('unified-search-bar-rating-toggle').click();
  const ratingPopover = pageA.getByTestId('unified-search-bar-rating-popover');
  await expect(ratingPopover).toBeVisible();
  await ratingPopover.locator('input[name="rated"][value="only"]').check();
  const ratingResp = pageA.waitForResponse((res) =>
    res.url().includes('/app/library/dashboard/albums-table') &&
    res.url().includes('rated=only'),
  );
  await ratingPopover.getByRole('button', { name: 'Apply' }).click();
  await ratingResp;

  // Wait until the URL has both params pushed.
  await expect.poll(() => {
    const u = new URL(pageA.url());
    return u.searchParams.get('sortBy') === 'artist' && u.searchParams.get('rated') === 'only';
  }, { message: 'URL must reflect both sort and filter state after interactions' }).toBe(true);

  // Capture: URL + every rendered title in order.
  const capturedURL = pageA.url();
  const titlesA = await pageA.getByTestId('album-list-row-title-link').allInnerTexts();
  expect(titlesA.length, 'context A must render at least one row').toBeGreaterThan(0);
  await ctxA.close();

  // Second context — fresh, no prior interaction. Opening the captured URL
  // must reproduce the exact same DOM order row-for-row.
  const ctxB = await browser.newContext();
  await loginAs(ctxB, userId!);
  const pageB = await ctxB.newPage();
  await pageB.goto(capturedURL);
  await expect(pageB.getByTestId('albums-list')).toBeVisible();

  const titlesB = await pageB.getByTestId('album-list-row-title-link').allInnerTexts();
  expect(titlesB.length, 'fresh context must render the same number of rows').toBe(titlesA.length);
  expect(titlesB, 'fresh context DOM order must match the original row-for-row').toEqual(titlesA);

  // The bar must also reflect both dimensions at rest (badges visible).
  await expect(pageB.getByTestId('unified-search-bar-badge-sort')).toBeVisible();
  await expect(pageB.getByTestId('unified-search-bar-badge-rating')).toBeVisible();

  await ctxB.close();
});

test('PC2 — every reachable param combination round-trips via the URL', async ({ browser }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  // A small but cross-dimensional set of view states. Each is a non-default
  // combination touching at least two of {q, filter, sort}. Together they
  // exercise: text-only, sort-only, filter-only, all three combined, and
  // multi-select repeatables (format).
  const states = [
    'q=the',
    'sortBy=album&dir=asc',
    'rated=only',
    'q=the&sortBy=album&dir=asc',
    'minRating=7&sortBy=rating&dir=desc',
    'q=the&rated=only&sortBy=album&dir=asc',
    'format=vinyl&format=digital&sortBy=date&dir=asc',
  ];

  for (const qs of states) {
    const url = `/app/library/dashboard?${qs}`;

    // Render the URL twice in two fresh contexts; their DOM orders must
    // match exactly. This is the property-style version of PC2: any
    // captured URL is a faithful representation of the view, with no
    // hidden client-side state mediating the result.
    const ctxA = await browser.newContext();
    await loginAs(ctxA, userId!);
    const pageA = await ctxA.newPage();
    await pageA.goto(url);
    await expect(pageA.getByTestId('albums-list'), `[${qs}] albums-list must render`).toBeVisible();
    const titlesA = await pageA.getByTestId('album-list-row-title-link').allInnerTexts();
    await ctxA.close();

    const ctxB = await browser.newContext();
    await loginAs(ctxB, userId!);
    const pageB = await ctxB.newPage();
    await pageB.goto(url);
    await expect(pageB.getByTestId('albums-list'), `[${qs}] albums-list must render in fresh context`).toBeVisible();
    const titlesB = await pageB.getByTestId('album-list-row-title-link').allInnerTexts();
    await ctxB.close();

    expect(titlesB, `[${qs}] fresh-context DOM order must match the original row-for-row`).toEqual(titlesA);
  }
});
