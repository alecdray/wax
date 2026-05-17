import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/tags.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

test('Tags modal opens from the album detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-page-tags-edit').click();

  await expect(page.locator('dialog[open]')).toBeVisible();
  await expect(page.getByTestId('tags-form-input')).toBeVisible();
});

test('Typing a tag name adds a chip', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-page-tags-edit').click();
  await expect(page.locator('dialog[open]')).toBeVisible();

  await page.getByTestId('tags-form-input').fill('e2e-test-tag');
  await page.getByTestId('tags-form-input').press('Enter');

  // Chip should appear inside the modal
  const chips = page.locator('dialog[open]').getByTestId('tags-form-chip');
  await expect(chips).toContainText('e2e-test-tag');
});

test('Saving tags closes the modal', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-page-tags-edit').click();
  await expect(page.locator('dialog[open]')).toBeVisible();

  await page.getByTestId('tags-form-save').click();

  await expect(page.locator('dialog[open]')).not.toBeVisible();
});
