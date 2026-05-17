import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/discover.feature

const userId = process.env.E2E_TEST_USER_ID;

test('Searching Spotify for an album returns matching results', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  await expect(page.getByTestId('discover-page-search-input')).toBeVisible();

  // Typing fires hx-get with a 300ms debounce on the `keyup` event; the
  // response swaps innerHTML of #discover-results, replacing the empty-query
  // placeholder with either the results list or a no-hits message. We use
  // pressSequentially (not fill) so real keyup events are emitted, and wait
  // on the observable DOM signal: the results list appearing.
  await page.getByTestId('discover-page-search-input').pressSequentially('the beatles');

  await expect(page.getByTestId('discover-search-results')).toBeVisible();
  await expect(page.getByTestId('discover-search-result-row').first()).toBeVisible();
});

test('Visiting the discover page shows the radar', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  await expect(page.getByTestId('radar-carousel')).toBeVisible();

  // Radar shows either at least one carousel item or the empty-state message.
  const items = page.getByTestId('radar-carousel-item');
  const empty = page.getByTestId('radar-carousel-empty');
  const itemCount = await items.count();
  if (itemCount > 0) {
    await expect(items.first()).toBeVisible();
  } else {
    await expect(empty).toBeVisible();
  }
});
