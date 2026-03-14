import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/reviews.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

async function openRatingModal(page: any) {
  await page.getByTestId('album-detail-rating').locator('[hx-get*="rating-recommender"]').click();
  await expect(page.getByTestId('rating-confirm')).toBeVisible();
}

async function submitRating(page: any, score: string) {
  await openRatingModal(page);
  await page.getByTestId('rating-input').fill(score);
  await page.getByTestId('rating-lock-in').click();
  await expect(page.locator('dialog[open]')).not.toBeVisible();
}

async function openRatingHistory(page: any) {
  const history = page.getByTestId('album-detail-rating-history');
  const checkbox = history.locator('input[type="checkbox"]');
  if (!await checkbox.isChecked()) {
    await checkbox.click({ force: true });
  }
}

async function deleteEntry(page: any) {
  await openRatingHistory(page);
  const responsePromise = page.waitForResponse(
    (resp: any) => resp.url().includes('/rating-log/') && resp.status() === 200
  );
  await page.getByTestId('rating-history-delete').first().click();
  await responsePromise;
}

async function clearAllEntries(page: any) {
  await openRatingHistory(page);
  const count = await page.getByTestId('rating-history-delete').count();
  for (let i = 0; i < count; i++) {
    await deleteEntry(page);
  }
}

test('Rating modal opens to the confirmation form', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await openRatingModal(page);

  await expect(page.getByTestId('rating-input')).toBeVisible();
  await expect(page.getByTestId('rating-lock-in')).toBeVisible();
});

test('Navigating to the questionnaire from the confirmation form', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await openRatingModal(page);
  await page.getByTestId('rating-confirm').locator('[hx-get*="questions"]').click();

  await expect(page.getByTestId('rating-questionnaire')).toBeVisible();
  await expect(page.getByTestId('rating-calculate')).toBeVisible();
});

test('Completing the questionnaire produces a score', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await openRatingModal(page);
  await page.getByTestId('rating-confirm').locator('[hx-get*="questions"]').click();
  await expect(page.getByTestId('rating-questionnaire')).toBeVisible();

  const fieldsets = page.locator('[data-testid="rating-questionnaire"] fieldset');
  const count = await fieldsets.count();
  for (let i = 0; i < count; i++) {
    await fieldsets.nth(i).locator('input[type="radio"]').first().check();
  }

  await page.getByTestId('rating-calculate').click();

  await expect(page.getByTestId('rating-confirm')).toBeVisible();
  await expect(page.getByTestId('rating-input')).not.toHaveValue('');
});

test('Saving a rating', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await submitRating(page, '7');
});

test('Saving a rating with a note', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await openRatingModal(page);
  await page.getByTestId('rating-input').fill('8');
  await page.getByTestId('rating-note').fill('A great listen.');
  await page.getByTestId('rating-lock-in').click();

  await expect(page.locator('dialog[open]')).not.toBeVisible();
  await expect(page.getByTestId('rating-history-note').first()).toContainText('A great listen.');
});

test('No delete button in the rating modal', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await openRatingModal(page);

  await expect(page.getByTestId('rating-delete')).not.toBeVisible();
});

test('Rating history is shown on the album detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await clearAllEntries(page);

  await submitRating(page, '6');
  await submitRating(page, '7.5');

  await expect(page.getByTestId('rating-history-entry')).toHaveCount(2);
  await expect(page.getByTestId('album-detail-rating-history')).toBeVisible();
});

test('Deleting the only rating entry clears the album rating', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await clearAllEntries(page);

  await submitRating(page, '5');
  await expect(page.getByTestId('rating-history-entry')).toHaveCount(1);

  await deleteEntry(page);

  await expect(page.getByTestId('rating-history-entry')).toHaveCount(0);
});

test('No notes on dashboard', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await expect(page.getByTestId('album-row-notes')).not.toBeVisible();
  await expect(page.getByTestId('album-row-notes-button')).not.toBeVisible();
});
