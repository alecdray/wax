import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';
import { seedAlbumGenre, clearAlbumGenres } from '../helpers/db';

// Scenarios from e2e/feat/genres.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

// Wikidata Q-ids in the curated primary set. hyperpop resolves to pop +
// electronic; reggae is a primary no fixture album carries.
const POP = 'Q37073';
const REGGAE = 'Q9794';
const HYPERPOP = 'Q104695865';

test.beforeAll(() => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();
  // hyperpop → pop + electronic. The fixture album is the only one with genres,
  // so it is the sole match for any specific-genre filter.
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
  // Every visible row matched the pop filter; the fixture album carries a
  // primary-genre badge, so at least one is present in the narrowed view.
  await expect(page.getByTestId('album-list-row')).toHaveCount(1);
  await expect(page.getByTestId('album-row-primary-genre').first()).toBeVisible();
});

test('Filtering by a genre no album has shows no albums', async ({ context, page }) => {
  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  await page.getByTestId('unified-search-bar-genre-toggle').click();
  const popover = page.getByTestId('unified-search-bar-genre-popover');
  await expect(popover).toBeVisible();
  await popover.locator(`input[name="primary"][value="${REGGAE}"]`).check();
  await popover.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByTestId('albums-list')).toBeVisible();
  await expect(page.getByTestId('album-list-row')).toHaveCount(0);
});
