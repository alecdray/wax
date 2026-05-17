import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/reviews.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

// Open the rating modal and navigate to the rating confirmation form.
// The modal may open to the questionnaire, the rerate prompt, or the
// confirmation form depending on the album's rating state; this helper
// drives whichever entry point appears to the confirmation form.
async function openConfirmForm(page: any) {
  const trigger = page.locator('[data-testid="album-score-badge-rated"], [data-testid="album-score-badge-unrated"]').first();
  await trigger.click();

  const dialog = page.locator('dialog[open]');
  await expect(dialog).toBeVisible();

  // Wait for one of the three possible modal entry points to render.
  await expect(
    dialog.locator(
      '[data-testid="rerate-prompt"], [data-testid="base-questions-form"], [data-testid="rating-confirm-form"]',
    ),
  ).toBeVisible();

  if (await dialog.getByTestId('rerate-prompt').isVisible().catch(() => false)) {
    await dialog.getByTestId('rerate-prompt-rate-now').click();
    await expect(dialog.getByTestId('base-questions-form')).toBeVisible();
  }

  if (await dialog.getByTestId('base-questions-form').isVisible().catch(() => false)) {
    await answerQuestionnaire(page);
    await dialog.getByTestId('base-questions-form-submit').click();
  }

  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
}

// Pick the first ("Strongly disagree") option for every question. All-1s
// answers produce a 0.0 score and avoid the finalized-mode contradiction check.
async function answerQuestionnaire(page: any) {
  const fieldsets = page.locator('dialog[open] [data-testid="base-question-fieldset"]');
  const count = await fieldsets.count();
  for (let i = 0; i < count; i++) {
    // base-question-radio testid is on the wrapping label; clicking the label
    // selects its child radio input.
    await fieldsets.nth(i).getByTestId('base-question-radio').first().click();
  }
}

async function submitRating(page: any, score: string, note?: string) {
  await openConfirmForm(page);
  const dialog = page.locator('dialog[open]');
  await dialog.getByTestId('rating-confirm-form-input').fill(score);
  if (note !== undefined) {
    await dialog.getByTestId('rating-confirm-form-note').fill(note);
  }
  await dialog.getByTestId('rating-confirm-form-lock-in').click();
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

test('Completing the questionnaire produces a score', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  const trigger = page.locator('[data-testid="album-score-badge-rated"], [data-testid="album-score-badge-unrated"]').first();
  await trigger.click();

  const dialog = page.locator('dialog[open]');
  await expect(dialog).toBeVisible();

  if (await dialog.getByTestId('rerate-prompt').isVisible().catch(() => false)) {
    await dialog.getByTestId('rerate-prompt-rate-now').click();
  }
  if (await dialog.getByTestId('rating-confirm-form').isVisible().catch(() => false)) {
    // Already on confirm form (album was previously finalized); navigate back to the questionnaire.
    await dialog.getByTestId('rating-confirm-form-back-to-questions').click();
  }

  await expect(dialog.getByTestId('base-questions-form')).toBeVisible();
  await answerQuestionnaire(page);
  await dialog.getByTestId('base-questions-form-submit').click();

  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
  await expect(dialog.getByTestId('rating-confirm-form-input')).not.toHaveValue('');
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

test('No delete button in the rating modal', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  const trigger = page.locator('[data-testid="album-score-badge-rated"], [data-testid="album-score-badge-unrated"]').first();
  await trigger.click();

  const dialog = page.locator('dialog[open]');
  await expect(dialog).toBeVisible();
  // Semantic absence: the rating modal exposes no button accessible as "Delete".
  // The only delete affordance for ratings lives in the album-rating-history
  // section on the album detail page, outside the modal.
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

