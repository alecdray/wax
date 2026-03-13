import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/album_detail.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;
// An album that has listening history recorded. May be the same as albumId
// if that album has been played, or a dedicated env var when they differ.
const albumWithHistoryId = process.env.E2E_TEST_ALBUM_WITH_HISTORY_ID ?? albumId;

test('Viewing an album in the library', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await expect(page.getByTestId('album-detail-cover')).toBeVisible();
  await expect(page.getByTestId('album-detail-title')).toBeVisible();
  await expect(page.getByTestId('album-detail-artists')).toBeVisible();
  await expect(page.getByTestId('album-detail-releases')).toBeVisible();
  await expect(page.getByTestId('album-detail-rating')).toBeVisible();
});

test('Navigating to the detail page from the dashboard', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('album-row-title-link').first().click();

  await expect(page).toHaveURL(/\/app\/library\/albums\//);
  await expect(page.getByTestId('album-detail-title')).toBeVisible();
});

test('Rating an album from the detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  // Click the rating trigger (either the badge or the "Rate" button)
  await page.getByTestId('album-detail-rating').locator('[hx-get*="rating-recommender"]').click();

  await expect(page.locator('dialog[open]')).toBeVisible();
});

test('Editing notes from the detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-notes').locator('button').click();

  await expect(page.locator('dialog[open]')).toBeVisible();
});

test('Editing tags from the detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-tags-edit').click();

  await expect(page.locator('dialog[open]')).toBeVisible();
});

test('Accessing an album not in the library', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  const response = await page.goto('/app/library/albums/00000000-0000-0000-0000-000000000000');

  expect(response?.status()).toBe(404);
});

test('Last played date is shown when available', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumWithHistoryId, 'E2E_TEST_ALBUM_WITH_HISTORY_ID (or E2E_TEST_ALBUM_ID) must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumWithHistoryId}`);

  await expect(page.getByTestId('album-detail-last-played')).toBeVisible();
});

test('Last played date is absent when not available', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await expect(page.getByTestId('album-detail-last-played')).not.toBeVisible();
});

test('Back navigation to the dashboard', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-back-link').click();

  await expect(page).toHaveURL('/app/library/dashboard');
});
