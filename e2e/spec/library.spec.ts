import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

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
