import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/library.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

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

// --- Sort chip ---

test('Sort chip is visible and shows default sort', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  const chip = page.getByTestId('filter-chip-bar-sort');
  await expect(chip).toBeVisible();
  await expect(chip).toContainText('Date Added');
});

test('Sort chip opens a modal', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('filter-chip-bar-sort').click();

  await expect(page.locator('dialog[open]')).toBeVisible();
  await expect(page.locator('dialog[open] input[name="sortBy"]').first()).toBeVisible();
});

test('Sorting by artist via sort chip reloads the list', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('filter-chip-bar-sort').click();
  await page.locator('dialog[open] input[name="sortBy"][value="artist"]').check();
  await page.locator('dialog[open] button[type="submit"]').click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('filter-chip-bar-sort')).toContainText('Artist');
});

// --- Rating chip ---

test('Rating chip opens a modal with min/max inputs', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('filter-chip-bar-rating').click();

  await expect(page.locator('dialog[open]')).toBeVisible();
  await expect(page.locator('dialog[open] input[name="minRating"]')).toBeVisible();
  await expect(page.locator('dialog[open] input[name="maxRating"]')).toBeVisible();
  await expect(page.locator('dialog[open] input[name="rated"]').first()).toBeVisible();
});

test('Rating chip becomes active after applying a min rating filter', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('filter-chip-bar-rating').click();
  await page.locator('dialog[open] input[name="minRating"]').fill('7');
  await page.locator('dialog[open] button[type="submit"]').click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('filter-chip-bar-rating')).toContainText('7');
});

test('Filtering to unrated only shows unrated chip label', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('filter-chip-bar-rating').click();
  await page.locator('dialog[open] input[name="rated"][value="unrated"]').check();
  await page.locator('dialog[open] button[type="submit"]').click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('filter-chip-bar-rating')).toContainText('Unrated');
});

// --- Format chip ---

test('Format chip opens a modal with format options', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('filter-chip-bar-format').click();

  await expect(page.locator('dialog[open]')).toBeVisible();
  await expect(page.locator('dialog[open] input[name="format"][value="vinyl"]')).toBeVisible();
  await expect(page.locator('dialog[open] input[name="format"][value="digital"]')).toBeVisible();
});

test('Format chip becomes active after selecting vinyl', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('filter-chip-bar-format').click();
  await page.locator('dialog[open] input[name="format"][value="vinyl"]').check();
  await page.locator('dialog[open] button[type="submit"]').click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('filter-chip-bar-format')).toContainText('vinyl');
});

// --- Artist chip ---

test('Artist chip opens a modal when artists exist', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  const chip = page.getByTestId('filter-chip-bar-artist');
  if (!await chip.isVisible()) {
    // No artists in library for this test user — skip
    test.skip();
    return;
  }

  await chip.click();

  await expect(page.locator('dialog[open]')).toBeVisible();
  await expect(page.locator('dialog[open]').getByTestId('filter-chip-bar-artist-checkbox').first()).toBeVisible();
});

// --- Rating modal from list row ---

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
