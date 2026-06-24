import { test, expect, type Page } from '@playwright/test';
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
  await dialog.getByTestId('rating-confirm-form-save').click();
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
  const before = await page.getByTestId('album-rating-history-delete').count();
  await page.getByTestId('album-rating-history-delete').first().click();
  // The delete response only fires the `album-changed` event; the rating
  // history is refreshed by a follow-up /app/library/album-surfaces request.
  // Wait on that observable DOM change, not the mutation response.
  await expect(page.getByTestId('album-rating-history-delete')).toHaveCount(before - 1);
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

test('Save only on a finalized album demotes it to provisional', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);
  await setupFinalized(page, '7.0');

  // Save a different score on the now-finalized album via Save only — a single
  // submission with no confirmation prompt. Save only always demotes a
  // finalized album to provisional.
  await openModal(page);
  const dialog = page.locator('dialog[open]');
  await dialog.getByTestId('rating-confirm-form-input').fill('7.5');
  await dialog.getByTestId('rating-confirm-form-save').click();
  await expect(page.locator('dialog[open]')).toHaveCount(0);

  // The detail-page score badge renders the provisional state-icon only when
  // the album is provisional — its presence confirms the demotion.
  await page.goto(`/app/library/albums/${albumId}`);
  await expect(page.getByTestId('album-score-badge-state-icon')).toBeVisible();
});

test('Both save buttons are visible on the score-entry form regardless of state', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  // Each setup helper ends in page.reload(), which tears down any open dialog,
  // so the next iteration starts from a clean, modal-free page.
  for (const setup of [resetToUnrated, (p: Page) => setupProvisional(p, '6.0'), (p: Page) => setupFinalized(p, '8.0')]) {
    await setup(page);
    await expect(page.locator('dialog[open]')).toHaveCount(0);
    await openModal(page);
    const dialog = page.locator('dialog[open]');
    await expect(dialog.getByTestId('rating-confirm-form-finalize')).toBeVisible();
    await expect(dialog.getByTestId('rating-confirm-form-save')).toBeVisible();
  }
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

  // Album is now finalized: the detail-page score badge shows no provisional
  // state-icon (the icon renders only while an album is provisional).
  await page.goto(`/app/library/albums/${albumId}`);
  await expect(page.getByTestId('album-score-badge-state-icon')).toHaveCount(0);
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

