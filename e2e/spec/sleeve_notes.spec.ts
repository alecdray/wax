import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/sleeve_notes.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

test('Saving a sleeve note from the album detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  const section = page.getByTestId('sleeve-notes-section');
  await expect(section).toBeVisible();

  // The editor trigger lives in either the display or empty view depending
  // on whether the album already has a sleeve note.
  const editTrigger = section.locator(
    '[data-testid="sleeve-notes-display-edit"], [data-testid="sleeve-notes-empty-edit"]',
  );
  await editTrigger.click();

  // HTMX swap completed when the editor textarea is in the DOM.
  const textarea = page.getByTestId('sleeve-notes-editor-textarea');
  await expect(textarea).toBeVisible();

  // Unique per-run note text so re-runs do not false-positive on prior content,
  // and so this run can assert THIS write hit the database.
  const uniqueMarker = `e2e-sleeve-note-${Date.now()}`;
  const noteText = `End-to-end test note ${uniqueMarker}`;
  await textarea.fill(noteText);

  await page.getByTestId('sleeve-notes-editor-save').click();

  // After save, the editor is replaced by the rendered display view.
  const display = page.getByTestId('sleeve-notes-display');
  await expect(display).toBeVisible();
  await expect(page.getByTestId('sleeve-notes-editor')).toHaveCount(0);

  // The rendered note (markdown-rendered HTML) reflects the text just entered.
  await expect(page.getByTestId('sleeve-notes-display-content')).toContainText(uniqueMarker);
});
