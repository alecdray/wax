import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/navigation.feature

const userId = process.env.E2E_TEST_USER_ID;

test('The bottom nav marks the current destination on the dashboard', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await expect(page.getByTestId('bottom-nav')).toBeVisible();

  // The active tab carries aria-current="page" (see docs/design/testids.md
  // "Selected state"); the inactive tab does not.
  await expect(page.getByTestId('bottom-nav-library')).toHaveAttribute('aria-current', 'page');
  await expect(page.getByTestId('bottom-nav-radar')).not.toHaveAttribute('aria-current', 'page');
});

test('Selecting Radar from the bottom nav navigates to the radar page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // The bottom nav is hx-boosted; clicking swaps the body and pushes the URL.
  await page.getByTestId('bottom-nav-radar').click();

  await expect(page).toHaveURL('/app/library/radar');
  await expect(page.getByTestId('radar-page')).toBeVisible();
  await expect(page.getByTestId('bottom-nav-radar')).toHaveAttribute('aria-current', 'page');
});
