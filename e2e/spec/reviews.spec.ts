import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';
import {
  getAlbumTitle,
  resetAlbumRating,
  seedRatingLogEntry,
  setRatingStateValue,
} from '../helpers/db';

// Scenarios from e2e/feat/reviews.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

// Open the rating modal. The modal always opens to the score-entry form,
// regardless of the album's rating state.
async function openModal(page: any) {
  const trigger = page
    .locator('[data-testid="album-score-badge-rated"], [data-testid="album-score-badge-unrated"]')
    .first();
  await trigger.click();

  const dialog = page.locator('dialog[open]');
  await expect(dialog).toBeVisible();
  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
}

// Pick the first ("Strongly disagree") option for every question. All-1s
// answers produce a 0.0 score and avoid the finalized-mode contradiction check.
async function answerQuestionnaire(page: any) {
  const fieldsets = page.locator('dialog[open] [data-testid="base-question-fieldset"]');
  const count = await fieldsets.count();
  for (let i = 0; i < count; i++) {
    await fieldsets.nth(i).getByTestId('base-question-radio').first().click();
  }
}

async function submitRating(page: any, score: string, note?: string) {
  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await dialog.getByTestId('rating-confirm-form-input').fill(score);
  if (note !== undefined) {
    await dialog.getByTestId('rating-confirm-form-note').fill(note);
  }
  await dialog.getByTestId('rating-confirm-form-lock-in').click();
  await expect(page.locator('dialog[open]')).toHaveCount(0);
}

async function finalizeRating(page: any, score: string) {
  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await dialog.getByTestId('rating-confirm-form-input').fill(score);
  await dialog.getByTestId('rating-confirm-form-finalize').click();
  await expect(page.locator('dialog[open]')).toHaveCount(0);
}

async function openRatingHistory(page: any) {
  const history = page.getByTestId('album-rating-history');
  await expect(history).toBeVisible();
  const checkbox = history.getByTestId('album-rating-history-toggle');
  if (!(await checkbox.isChecked())) {
    await checkbox.click({ force: true });
  }
}

async function deleteEntry(page: any) {
  await openRatingHistory(page);
  const responsePromise = page.waitForResponse(
    (resp: any) => resp.url().includes('/rating-log/') && resp.status() === 200,
  );
  await page.getByTestId('album-rating-history-delete').first().click();
  await responsePromise;
}

async function clearAllEntries(page: any) {
  await openRatingHistory(page);
  let count = await page.getByTestId('album-rating-history-delete').count();
  while (count > 0) {
    await deleteEntry(page);
    count = await page.getByTestId('album-rating-history-delete').count();
  }
}

// Resets the album to a fully-unrated shape (no log entries, no state row).
// Forces a page reload so the in-DOM badges / readouts pick up the new state.
async function resetToUnrated(page: any) {
  resetAlbumRating(userId!, albumId!);
  await page.reload();
}

// Leaves the album with a provisional state row and one rating-log entry,
// dated comfortably in the past so any in-test save reads unambiguously as
// the new latest entry.
async function setupProvisional(page: any, score: string) {
  resetAlbumRating(userId!, albumId!);
  seedRatingLogEntry(userId!, albumId!, parseFloat(score));
  setRatingStateValue(userId!, albumId!, 'provisional');
  await page.reload();
}

// Leaves the album with a finalized state row and one rating-log entry,
// dated in the past for the same reason.
async function setupFinalized(page: any, score: string) {
  resetAlbumRating(userId!, albumId!);
  seedRatingLogEntry(userId!, albumId!, parseFloat(score));
  setRatingStateValue(userId!, albumId!, 'finalized');
  await page.reload();
}

test('Modal opens to the score-entry form for an unrated album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await resetToUnrated(page);

  await openModal(page);

  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
  // Neither legacy modal entry point appears as the first view.
  await expect(dialog.getByTestId('rerate-prompt')).toHaveCount(0);
  await expect(dialog.getByTestId('base-questions-form')).toHaveCount(0);
  // No prior log entry, so the input opens empty.
  await expect(dialog.getByTestId('rating-confirm-form-input')).toHaveValue('');
});

test('Score-entry form pre-fills with the latest rating for a provisional album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupProvisional(page, '6.5');

  await openModal(page);

  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
  await expect(dialog.getByTestId('rating-confirm-form-input')).toHaveValue('6.5');
});

test('Score-entry form pre-fills with the latest rating for a finalized album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupFinalized(page, '8.1');

  await openModal(page);

  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
  await expect(dialog.getByTestId('rating-confirm-form-input')).toHaveValue('8.1');
});

test("Opening the questionnaire from the score-entry form pre-fills the score after submit", async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await resetToUnrated(page);

  await openModal(page);
  const dialog = page.locator('dialog[open]');

  await dialog.getByTestId('rating-confirm-form-open-questionnaire').click();
  await expect(dialog.getByTestId('base-questions-form')).toBeVisible();

  await answerQuestionnaire(page);
  await dialog.getByTestId('base-questions-form-submit').click();

  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
  // Computed score is now in the input (some non-empty value).
  await expect(dialog.getByTestId('rating-confirm-form-input')).not.toHaveValue('');
});

test('Dismissing the questionnaire preserves the prior pre-fill', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupProvisional(page, '7.3');

  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('rating-confirm-form-input')).toHaveValue('7.3');

  await dialog.getByTestId('rating-confirm-form-open-questionnaire').click();
  await expect(dialog.getByTestId('base-questions-form')).toBeVisible();

  // Dismiss without answering or submitting.
  await dialog.getByTestId('base-questions-form-dismiss').click();

  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
  await expect(dialog.getByTestId('rating-confirm-form-input')).toHaveValue('7.3');
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

  await submitRating(page, '8', 'A great listen.');

  await expect(page.getByTestId('album-rating-history-note').first()).toContainText('A great listen.');
});

test('Saving on a finalized album keeps it finalized in one submission', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupFinalized(page, '7.0');

  // Save a different score on the now-finalized album — single submission, no
  // confirmation prompt, no extra round-trip.
  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await dialog.getByTestId('rating-confirm-form-input').fill('7.5');
  await dialog.getByTestId('rating-confirm-form-lock-in').click();
  await expect(page.locator('dialog[open]')).toHaveCount(0);

  // Reopen the modal and confirm the album is still finalized: opening it
  // again must not show the Finalize button (only provisional albums do).
  await openModal(page);
  const dialog2 = page.locator('dialog[open]');
  await expect(dialog2.getByTestId('rating-confirm-form-finalize')).toHaveCount(0);
  await expect(dialog2.getByTestId('rating-confirm-form-input')).toHaveValue('7.5');
});

test('Finalize button is visible on the score-entry form for a provisional album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupProvisional(page, '6.0');

  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('rating-confirm-form-finalize')).toBeVisible();
});

test('Finalize button is hidden on the score-entry form for an unrated album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await resetToUnrated(page);

  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('rating-confirm-form-finalize')).toHaveCount(0);
});

test('Finalize button is hidden on the score-entry form for a finalized album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupFinalized(page, '8.0');

  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await expect(dialog.getByTestId('rating-confirm-form-finalize')).toHaveCount(0);
});

test('Clicking Finalize promotes the album in one submission', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupProvisional(page, '6.0');

  await openRatingHistory(page);
  const beforeCount = await page.getByTestId('album-rating-history-entry').count();

  await finalizeRating(page, '8.0');

  // History grew by one entry, recording the score we submitted via Finalize.
  await openRatingHistory(page);
  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(beforeCount + 1);
  const scores = await page.getByTestId('album-rating-history-score').allTextContents();
  expect(scores.some((s) => s.startsWith('8'))).toBe(true);

  // Album is now finalized: reopening the modal hides the Finalize button.
  await openModal(page);
  await expect(page.locator('dialog[open]').getByTestId('rating-confirm-form-finalize')).toHaveCount(0);
});

test('No delete button in the rating modal', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  const trigger = page
    .locator('[data-testid="album-score-badge-rated"], [data-testid="album-score-badge-unrated"]')
    .first();
  await trigger.click();

  const dialog = page.locator('dialog[open]');
  await expect(dialog).toBeVisible();
  await expect(dialog.getByRole('button', { name: /delete/i })).toHaveCount(0);
});

test('Rating history is shown on the album detail page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await clearAllEntries(page);

  await submitRating(page, '6');
  await submitRating(page, '7.5');

  await expect(page.getByTestId('album-rating-history')).toBeVisible();
  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(2);
});

test('Submitting a second rating creates a new history entry', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await clearAllEntries(page);

  await submitRating(page, '6');
  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(1);

  await submitRating(page, '8');

  await openRatingHistory(page);
  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(2);
  // The current rating shown in the score badge is one of the two ratings (server picks
  // the most recent). Both ratings should be present as separate history entries.
  const scores = await page.getByTestId('album-rating-history-score').allTextContents();
  expect(scores.some((s) => s.startsWith('6'))).toBe(true);
  expect(scores.some((s) => s.startsWith('8'))).toBe(true);
});

test('Deleting the current rating from history rolls back to the previous one', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await clearAllEntries(page);

  await submitRating(page, '6');
  await submitRating(page, '8');

  await openRatingHistory(page);
  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(2);
  // The first entry in the (reverse-chronological) history is the current rating.
  const currentScore = (await page.getByTestId('album-rating-history-score').first().textContent()) ?? '';
  await expect(page.getByTestId('album-score-badge-rated')).toContainText(currentScore.split(' ')[0]);

  // Delete that current entry.
  await deleteEntry(page);

  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(1);
  // The remaining (previously second-most-recent) entry is now the current rating.
  const remainingScore = (await page.getByTestId('album-rating-history-score').first().textContent()) ?? '';
  expect(remainingScore).not.toEqual(currentScore);
  await expect(page.getByTestId('album-score-badge-rated')).toContainText(remainingScore.split(' ')[0]);
});

test('Deleting the only rating entry clears the album rating', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await clearAllEntries(page);

  await submitRating(page, '5');
  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(1);

  await deleteEntry(page);

  await expect(page.getByTestId('album-rating-history-entry')).toHaveCount(0);
});

// --- Score-readout icon scheme ---
//
// The dashboard list view's score-readout cell shows a `pen` icon when the
// album is provisional, and no status icon for finalized or unrated albums.
// Locates the readout via the page-scoped data-testid (rated vs unrated) for
// the fixture album, then asserts the status-icon child.

async function gotoDashboardAndFindRow(page: any, albumId: string) {
  // Filter the list down to rows whose title contains the album's exact title
  // — for the fixture album that collapses the visible set to exactly one row
  // (the test album's own row), which we then return for in-row assertions.
  const title = getAlbumTitle(albumId);
  expect(title, `album ${albumId} must exist in the DB`).toBeTruthy();
  await page.goto(`/app/library/dashboard?q=${encodeURIComponent(title)}`);
  await expect(page.getByTestId('albums-list')).toBeVisible();
  const row = page.getByTestId('album-list-row').first();
  await expect(row, `album ${albumId} must render in the filtered list`).toBeVisible();
  return row;
}

test('Score readout shows the pen icon for a provisional album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  // Position the fixture album as provisional with a real rating-log entry so
  // the rated readout is rendered (not the unrated one).
  resetAlbumRating(userId!, albumId!);
  seedRatingLogEntry(userId!, albumId!, 6.5);
  setRatingStateValue(userId!, albumId!, 'provisional');

  await loginAs(context, userId!);
  const row = await gotoDashboardAndFindRow(page, albumId!);

  const readout = row.getByTestId('album-score-readout-rated');
  await expect(readout).toBeVisible();
  const stateIcon = readout.getByTestId('album-score-readout-state-icon');
  await expect(stateIcon).toBeVisible();
  // The wrapper carries data-icon-name set from the templ's ratingStateIcon
  // selector — asserting on it pins the icon-scheme contract to a stable,
  // test-only attribute (vs the BI class set, which would tie the test to
  // the primitive's internal class shape).
  await expect(stateIcon).toHaveAttribute('data-icon-name', 'pen');
});

test('Score readout shows no status icon for a finalized album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  resetAlbumRating(userId!, albumId!);
  seedRatingLogEntry(userId!, albumId!, 8.0);
  setRatingStateValue(userId!, albumId!, 'finalized');

  await loginAs(context, userId!);
  const row = await gotoDashboardAndFindRow(page, albumId!);

  const readout = row.getByTestId('album-score-readout-rated');
  await expect(readout).toBeVisible();
  await expect(readout.getByTestId('album-score-readout-state-icon')).toHaveCount(0);
});

test('Score readout shows no status icon for an unrated album', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  // Wipe both the log and the state row so the unrated readout is rendered.
  resetAlbumRating(userId!, albumId!);

  await loginAs(context, userId!);
  const row = await gotoDashboardAndFindRow(page, albumId!);

  const readout = row.getByTestId('album-score-readout-unrated');
  await expect(readout).toBeVisible();
  await expect(readout.getByTestId('album-score-readout-state-icon')).toHaveCount(0);
});
