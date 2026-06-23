import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/app_header.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

test('The app header appears on every authenticated page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);

  const paths = [
    '/app/library/dashboard',
    '/app/library/discover',
    `/app/library/albums/${albumId}`,
  ];

  for (const path of paths) {
    await page.goto(path);
    await expect(page.getByTestId('library-app-header'), `header on ${path}`).toBeVisible();
    await expect(page.getByTestId('library-app-header-wordmark'), `wordmark on ${path}`).toBeVisible();
    await expect(page.getByTestId('feeds-dropdown-button'), `feeds control on ${path}`).toBeVisible();
  }
});
