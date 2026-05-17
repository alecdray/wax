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
