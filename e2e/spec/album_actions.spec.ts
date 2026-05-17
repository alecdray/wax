import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/album_actions.feature

const userId = process.env.E2E_TEST_USER_ID;

test('Opening the actions modal for a non-library album from a discover search result', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  // Search for albums by a high-recall artist so we get plenty of hits to
  // pick from. As in discover.spec.ts we pressSequentially to fire keyup
  // events for the debounced hx-trigger.
  await page.getByTestId('discover-page-search-input').pressSequentially('the beatles');
  await expect(page.getByTestId('discover-search-results')).toBeVisible();

  // Find the first result that is NOT already in the user's library — only
  // those rows fire the album-actions modal (in-library rows navigate to the
  // album detail page instead).
  const rows = page.getByTestId('discover-search-result-row');
  const nonLibraryRow = rows.filter({
    hasNot: page.locator('[data-testid="discover-result-state-badge-in-library"]'),
  }).first();
  await expect(nonLibraryRow).toBeVisible();

  // Capture the row title so we can assert it appears in the modal header.
  const rowTitle = (await nonLibraryRow.getByTestId('discover-search-result-row-title').textContent())?.trim() ?? '';
  expect(rowTitle.length).toBeGreaterThan(0);

  await nonLibraryRow.click();

  // The modal swaps into #global-modal-container as an oob swap. Wait on the
  // observable DOM signal: the open dialog plus the album-actions content.
  const dialog = page.locator('dialog[open]');
  await expect(dialog).toBeVisible();
  await expect(dialog.getByTestId('album-actions-modal-content')).toBeVisible();

  // Header: the album title from the row appears in the modal.
  await expect(dialog.getByTestId('album-actions-modal-content')).toContainText(rowTitle);

  // Open-in-Spotify is present for every non-library state.
  const spotifyLink = dialog.getByTestId('album-actions-modal-content-open-spotify');
  await expect(spotifyLink).toBeVisible();
  await expect(spotifyLink).toHaveAttribute('href', /open\.spotify\.com\/album\//);

  // At least one state-specific primary action is visible. Which one depends
  // on the album's relationship to the user:
  //   New     → Add to radar
  //   OnRadar → Add to library (+ Remove from radar)
  //   Removed → Re-acquire
  const stateAction = dialog.locator([
    '[data-testid="album-actions-modal-content-add-radar"]',
    '[data-testid="album-actions-modal-content-add-to-library"]',
    '[data-testid="album-actions-modal-content-reacquire"]',
  ].join(', '));
  await expect(stateAction.first()).toBeVisible();
});
