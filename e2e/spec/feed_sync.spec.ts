import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/feed_sync.feature

const userId = process.env.E2E_TEST_USER_ID;

test('Triggering a feed sync from the dashboard shows a syncing indicator', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // Open the feeds dropdown — the feeds control in the app header.
  await page.getByTestId('feeds-dropdown-button').click();
  const dropdown = page.getByTestId('feeds-dropdown-content');
  await expect(dropdown).toBeVisible();

  // Trigger sync on the first (and only, for the test user) feed.
  const syncButton = dropdown.getByTestId('feeds-dropdown-sync-button').first();
  await expect(syncButton).toBeEnabled();
  await syncButton.click();

  // The HTMX swap returns the dropdown with the feed marked as syncing —
  // the immediate, user-visible signal that the sync was accepted.
  await expect(dropdown.getByTestId('feeds-dropdown-sync-indicator')).toBeVisible();
  await expect(dropdown.getByTestId('feeds-dropdown-sync-button').first()).toBeDisabled();
});
