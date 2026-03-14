import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/library.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

test('Viewing the library dashboard', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await expect(page.getByTestId('library-stats')).toBeVisible();
  await expect(page.getByTestId('carousel-recently-spun-tab')).toBeVisible();
  await expect(page.getByTestId('albums-table')).toBeVisible();
});

test('Default carousel shows Recently Spun', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  const tab = page.getByTestId('carousel-recently-spun-tab');
  await expect(tab).toBeVisible();
  // Active tab has full opacity; inactive tabs are dimmed via text-base-content/40
  await expect(tab).not.toHaveAttribute('hx-get');
});

test('Switching the carousel to Unrated', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('carousel-unrated-tab').click();

  // After the HTMX swap the unrated tab becomes active (loses its hx-get trigger)
  await expect(page.getByTestId('carousel-unrated-tab')).not.toHaveAttribute('hx-get');
});

test('Sorting albums by artist', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // Click the Artists sortable column header
  await page.locator('#album-table th', { hasText: 'Artists' }).click();

  // Table reloads — wait for it to settle and verify it is still present
  await expect(page.getByTestId('albums-table')).toBeVisible();
  await expect(page.getByTestId('album-row-title-link').first()).toBeVisible();
});

test('Opening the rating modal from an album row', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('album-row-rating').first().click();

  await expect(page.locator('dialog[open]')).toBeVisible();
});


test('Opening the tags modal from an album row', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('album-row-menu').first().click();
  await page.getByTestId('album-row-tags-button').first().click();

  await expect(page.locator('dialog[open]')).toBeVisible();
});
