import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';
import {
  getAlbumTitle,
  getLibraryAlbumIds,
  resetAlbumRating,
  seedRatingLogEntry,
  setRatingStateValue,
} from '../helpers/db';
import { execSync } from 'node:child_process';

// Scenarios from e2e/feat/rating-lifecycle.feature
//
// Cross-surface invariants of the rating-modal rework. The per-surface
// behaviour is covered in reviews.spec.ts and library.spec.ts; this file
// covers the assembled-system properties that hold across surfaces.

const userId = process.env.E2E_TEST_USER_ID;
const DB_PATH = process.env.GOOSE_DBSTRING ?? './tmp/db.sql';

// execSql is a local sqlite shell-out for fixtures the helpers/db.ts API
// doesn't expose — namely seeding two rating-log rows that share a created_at
// for the pre-fill tie-break invariant.
function execSql(sql: string): string {
  return execSync(`sqlite3 ${DB_PATH} ${JSON.stringify(sql)}`, { encoding: 'utf8' });
}

// --- Modal opens directly to the score-entry form from every UI affordance ---
//
// PC4 says GET /app/review/rating-recommender always renders the score-entry
// form. To exercise the invariant across rating states (not just the modal
// route in isolation), set up two albums in two different states and open the
// modal for each from the dashboard score readout. Both must land on the same
// form, with no alternate first view appearing.

test("Modal opens directly to the score-entry form from every UI affordance", async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  const albumIds = getLibraryAlbumIds(userId!);
  expect(albumIds.length, 'fixture user must own at least two albums').toBeGreaterThanOrEqual(2);
  const finalizedId = albumIds[0];
  const provisionalId = albumIds[1];

  // Position the two albums: one finalized, one provisional, each with a
  // rating-log entry so the rated readout is rendered for both.
  resetAlbumRating(userId!, finalizedId);
  seedRatingLogEntry(userId!, finalizedId, 8.4);
  setRatingStateValue(userId!, finalizedId, 'finalized');

  resetAlbumRating(userId!, provisionalId);
  seedRatingLogEntry(userId!, provisionalId, 6.2);
  setRatingStateValue(userId!, provisionalId, 'provisional');

  await loginAs(context, userId!);

  for (const id of [finalizedId, provisionalId]) {
    const title = getAlbumTitle(id);
    expect(title, `album ${id} must exist`).toBeTruthy();
    await page.goto(`/app/library/dashboard?q=${encodeURIComponent(title)}`);
    await expect(page.getByTestId('albums-list')).toBeVisible();

    const row = page.getByTestId('album-list-row').first();
    await expect(row).toBeVisible();
    await row.getByTestId('album-score-readout-rated').click();

    const dialog = page.locator('dialog[open]');
    await expect(dialog).toBeVisible();
    await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
    // Neither retired modal first view appears.
    await expect(dialog.getByTestId('rerate-prompt')).toHaveCount(0);
    await expect(dialog.getByTestId('base-questions-form')).toHaveCount(0);

    // Navigate to a clean slate between iterations — reopening the modal on
    // the next album must reproduce the same first view from scratch.
    await page.goto('/app/library/dashboard');
    await expect(page.locator('dialog[open]')).toHaveCount(0);
  }
});

// --- Score-entry pre-fill: id-DESC tie-break on equal created_at ---
//
// PC5 calls for a property-style assertion that exercises the running
// system: two log entries with identical created_at, different ids, and the
// greater id's score wins the pre-fill. Sidesteps the second-resolution
// CURRENT_TIMESTAMP pitfall by seeding both rows with an explicit shared
// timestamp.

test("Score-entry pre-fill resolves the tie between two same-timestamp log entries by greater id", async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  const albumIds = getLibraryAlbumIds(userId!);
  expect(albumIds.length, 'fixture user must own at least one album').toBeGreaterThanOrEqual(1);
  const targetAlbum = albumIds[0];

  // Clean slate: no rating-log entries, no state row.
  resetAlbumRating(userId!, targetAlbum);

  // Seed two log rows that share a created_at down to the second. Use ids that
  // sort unambiguously (a-prefix vs z-prefix) so 'z…' is the greater id under
  // SQLite's text ORDER BY semantics; the row with the 9.3 rating must win
  // the pre-fill. Picks a non-round score so the type="number" input's value
  // normalisation (which strips trailing zeroes on round numbers) doesn't
  // hide the assertion.
  const sharedTimestamp = "2026-04-01 12:00:00";
  const aID = `tie-a-${Date.now()}`;
  const zID = `tie-z-${Date.now()}`;
  execSql(
    `INSERT INTO album_rating_log (id, user_id, album_id, rating, created_at)` +
      ` VALUES ('${aID}', '${userId}', '${targetAlbum}', 5.2, '${sharedTimestamp}');`,
  );
  execSql(
    `INSERT INTO album_rating_log (id, user_id, album_id, rating, created_at)` +
      ` VALUES ('${zID}', '${userId}', '${targetAlbum}', 9.3, '${sharedTimestamp}');`,
  );

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${targetAlbum}`);

  const trigger = page
    .locator('[data-testid="album-score-badge-rated"], [data-testid="album-score-badge-unrated"]')
    .first();
  await trigger.click();
  const dialog = page.locator('dialog[open]');
  await expect(dialog).toBeVisible();
  await expect(dialog.getByTestId('rating-confirm-form')).toBeVisible();
  await expect(dialog.getByTestId('rating-confirm-form-input')).toHaveValue('9.3');
});

// --- Dashboard carousel HTMX swap target id is stable ---
//
// PC9 says the carousel section's DOM element id is `carousel-section`, and
// HTMX swaps targeting #carousel-section continue to land on the same element.
// The post-rework rename of the Rerate Due tab to Provisional didn't change
// the id; a click on a non-active tab triggers an HTMX swap with
// hx-target="#carousel-section", and the post-swap DOM must still expose the
// same id on the new section.

test("Dashboard carousel HTMX swap target id is stable", async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // The section exists with the stable testid (which is the same string as
  // the DOM id) and the DOM id matches.
  const section = page.getByTestId('carousel-section');
  await expect(section).toBeVisible();
  await expect(section).toHaveAttribute('id', 'carousel-section');

  // Every tab declares hx-target="#carousel-section" — the swap contract that
  // PC9 says is stable across the rerate→provisional rename. Asserting the
  // attribute on each tab pins the contract at the templ boundary.
  await expect(page.getByTestId('carousel-section-unrated-tab')).toHaveAttribute('hx-target', '#carousel-section');
  await expect(page.getByTestId('carousel-section-provisional-tab')).toHaveAttribute('hx-target', '#carousel-section');

  // Exercise an actual swap to confirm the post-swap DOM still carries the
  // stable id (so the next swap has a valid target). Uses the Unrated tab,
  // whose handler renders the same surface as the recently-played view —
  // exercising the swap end-to-end without depending on any other tab's
  // data shape.
  const unratedTab = page.getByTestId('carousel-section-unrated-tab');
  await unratedTab.click();
  // After the swap, the active tab loses its hx-get trigger — the suite's
  // established signal that the swap completed.
  await expect(unratedTab).not.toHaveAttribute('hx-get');
  // And the swapped-in section still carries the stable id, ready for the
  // next swap to land on the same DOM element.
  await expect(page.getByTestId('carousel-section')).toHaveAttribute('id', 'carousel-section');
});
