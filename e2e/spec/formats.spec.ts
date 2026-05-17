import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/formats.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

test('Formats modal opens from the album detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-page-releases-btn').click();

  const modal = page.locator('dialog[open]');
  await expect(modal).toBeVisible();
  await expect(modal.getByTestId('formats-modal-content')).toBeVisible();
  await expect(modal.getByTestId('formats-modal-content-heading')).toHaveText('Formats');
  await expect(modal.getByTestId('digital-format-row')).toHaveCount(1);
  await expect(modal.getByTestId('physical-format-row')).toHaveCount(3);
});
