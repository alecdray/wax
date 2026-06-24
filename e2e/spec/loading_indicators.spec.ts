import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/loading_indicators.feature

const userId = process.env.E2E_TEST_USER_ID;

test('Global progress bar is present in the layout', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');
  await expect(page.locator('#global-progress')).toBeAttached();
});

test('Discover search results region carries a loading overlay', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  await loginAs(context, userId!);
  await page.goto('/app/library/radar');
  await expect(page.locator('#radar-results-region [data-testid="region-overlay"]')).toBeAttached();
});

test('Add-to-library button declares the disable-on-request contract', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  await loginAs(context, userId!);
  await page.goto('/app/library/radar');

  // Fire keyup events via pressSequentially (same pattern as radar.spec.ts
  // and album_actions.spec.ts) to trigger the debounced hx-get.
  await page.getByTestId('radar-page-search-input').pressSequentially('the beatles');
  await expect(page.getByTestId('discover-search-results')).toBeVisible();

  // Find the first result that is NOT already in the user's library — only
  // non-library rows open the album-actions modal (in-library rows navigate to
  // the detail page instead). New/OnRadar/Removed rows all render an action
  // button with the disable-on-request contract we are verifying here.
  const nonLibraryRow = page.getByTestId('discover-search-result-row').filter({
    hasNot: page.locator('[data-testid="discover-result-state-badge-in-library"]'),
  }).first();
  await expect(nonLibraryRow).toBeVisible();
  await nonLibraryRow.click();

  // The modal swaps into #global-modal-container as an oob swap.
  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('album-actions-modal-content')).toBeVisible();

  // The primary action button (add-radar, add-to-library, or reacquire — the
  // specific one depends on the album's state relative to the user's library)
  // must carry both the btn-busy class and hx-disabled-elt="this". These are
  // the static, structural contracts for the disable-on-request pattern.
  const actionBtn = dialog.locator([
    '[data-testid="album-actions-modal-content-add-radar"]',
    '[data-testid="album-actions-modal-content-add-to-library"]',
    '[data-testid="album-actions-modal-content-reacquire"]',
  ].join(', ')).first();
  await expect(actionBtn).toHaveAttribute('hx-disabled-elt', 'this');
  await expect(actionBtn).toHaveClass(/btn-busy/);
});
