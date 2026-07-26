import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';
import { seedAlbumGenre, clearAlbumGenres } from '../helpers/db';

// Scenarios from e2e/feat/genres.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

// Wikidata Q-ids in the curated primary set. hyperpop resolves to pop +
// electronic; reggae is a primary the fixture album does not carry.
const POP = 'Q37073';
const REGGAE = 'Q9794';
const HYPERPOP = 'Q104695865';

test.beforeAll(() => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();
  // hyperpop → pop + electronic. The fixture album will appear under a pop
  // filter and will not appear under a reggae filter.
  seedAlbumGenre(albumId!, HYPERPOP, 'hyperpop');
});

test.afterAll(() => {
  clearAlbumGenres(albumId!);
});

test("An album's primary genres show as badges on its detail page", async ({ context, page }) => {
  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await expect(page.getByTestId('album-detail-page-genres')).toBeVisible();
  const badges = page.getByTestId('album-detail-page-primary-genre');
  await expect(badges.first()).toBeVisible();
  const labels = (await badges.allInnerTexts()).map((t) => t.toLowerCase());
  expect(labels).toContain('pop');
  expect(labels).toContain('electronic');
});

test('Filtering the library by a primary genre keeps matching albums', async ({ context, page }) => {
  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('unified-search-bar-genre-toggle').click();
  const popover = page.getByTestId('unified-search-bar-genre-popover');
  await expect(popover).toBeVisible();
  await popover.locator(`input[name="primary"][value="${POP}"]`).check();
  await popover.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('unified-search-bar-genre-toggle')).toContainText('pop');
  // At least one row is visible and every genre badge on the page shows "pop"
  // (or a genre that maps to pop). The fixture album may be paginated past
  // page 0 if there are many pop albums in the DB.
  await expect(page.getByTestId('album-row-primary-genre').first()).toBeVisible();
  const badgeTexts = await page.getByTestId('album-row-primary-genre').allInnerTexts();
  expect(badgeTexts.map((t) => t.toLowerCase())).toContain('pop');
});

test('Filtering by a genre the fixture album lacks excludes it from results', async ({ context, page }) => {
  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('unified-search-bar-genre-toggle').click();
  const popover = page.getByTestId('unified-search-bar-genre-popover');
  await expect(popover).toBeVisible();
  await popover.locator(`input[name="primary"][value="${REGGAE}"]`).check();
  await popover.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  // The fixture album is seeded with hyperpop → pop + electronic, not reggae,
  // so it must not appear in the reggae-filtered results.
  await expect(page.locator(`a[href="/app/library/albums/${albumId!}"]`)).not.toBeVisible();
});
