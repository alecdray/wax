import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/discover.feature

const userId = process.env.E2E_TEST_USER_ID;

test('Searching Spotify for an album returns matching results', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  await expect(page.getByTestId('discover-page-search-input')).toBeVisible();

  // Typing fires hx-get with a 300ms debounce on the `keyup` event; the
  // response swaps innerHTML of #discover-results, replacing the empty-query
  // placeholder with either the results list or a no-hits message. We use
  // pressSequentially (not fill) so real keyup events are emitted, and wait
  // on the observable DOM signal: the results list appearing.
  await page.getByTestId('discover-page-search-input').pressSequentially('the beatles');

  await expect(page.getByTestId('discover-search-results')).toBeVisible();
  await expect(page.getByTestId('discover-search-result-row').first()).toBeVisible();
});

test('Visiting the discover page shows the radar', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  await expect(page.getByTestId('radar-carousel')).toBeVisible();

  // Radar shows either at least one carousel item or the empty-state message.
  const items = page.getByTestId('radar-carousel-item');
  const empty = page.getByTestId('radar-carousel-empty');
  const itemCount = await items.count();
  if (itemCount > 0) {
    await expect(items.first()).toBeVisible();
  } else {
    await expect(empty).toBeVisible();
  }
});

test('The discover page offers the Spotify radar inbox control', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  await expect(page.getByTestId('radar-inbox-control')).toBeVisible();

  // The control is in one of two states depending on whether this user has
  // already opted in: the enable button, or a link to their radar playlist.
  const enable = page.getByTestId('radar-inbox-enable');
  const link = page.getByTestId('radar-inbox-link');
  if ((await enable.count()) > 0) {
    await expect(enable).toBeVisible();
  } else {
    await expect(link).toBeVisible();
  }
});

test('A clear control is offered only while the search has a value', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  await expect(page.getByTestId('discover-page-search-input')).toBeVisible();
  await expect(page.getByTestId('discover-page-search-clear')).not.toBeVisible();

  const query = 'the beatles';
  const input = page.getByTestId('discover-page-search-input');
  await input.pressSequentially(query);
  await expect(page.getByTestId('discover-page-search-clear')).toBeVisible();

  await input.focus();
  for (let i = 0; i < query.length; i++) {
    await page.keyboard.press('Backspace');
  }

  await expect(input).toHaveValue('');
  await expect(page.getByTestId('discover-page-search-clear')).not.toBeVisible();
});

test('Activating the clear control resets the search to its empty state', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  const input = page.getByTestId('discover-page-search-input');
  await input.pressSequentially('the beatles');

  // Wait until the server has swapped in results — the empty-query placeholder
  // must be gone before we can meaningfully assert it returns after clearing.
  await expect(page.getByTestId('discover-search-results')).toBeVisible();
  await expect(page.getByTestId('discover-search-results-empty-query')).not.toBeVisible();

  await page.getByTestId('discover-page-search-clear').click();

  await expect(input).toHaveValue('');
  await expect(page.getByTestId('discover-search-results-empty-query')).toBeVisible();
  await expect(page.getByTestId('discover-search-results')).not.toBeVisible();
  await expect(input).toBeFocused();
});

test('Typing several characters quickly issues a single search after the debounce', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');
  await expect(page.getByTestId('discover-page-search-input')).toBeVisible();

  // Count every search request the page fires so we can assert that a burst of
  // keystrokes collapses into a single network round-trip (the contract of the
  // existing keyup-changed debounce). Without it, each keystroke would emit
  // its own request.
  let searchRequests = 0;
  page.on('request', (req) => {
    if (req.url().includes('/app/library/discover/search')) {
      searchRequests++;
    }
  });

  const query = 'beatles';
  const responsePromise = page.waitForResponse((res) =>
    res.url().includes('/app/library/discover/search') &&
    res.url().includes(`q=${encodeURIComponent(query)}`),
  );
  await page.getByTestId('discover-page-search-input').pressSequentially(query);
  await responsePromise;

  // The results swap is the observable signal that the debounced request has
  // landed; nothing more is in flight after this point.
  await expect(page.getByTestId('discover-search-results')).toBeVisible();

  expect(searchRequests, 'rapid typing should collapse to one search request').toBe(1);
});

test('Submitting the search bypasses the debounce and fires immediately', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');
  const input = page.getByTestId('discover-page-search-input');
  await expect(input).toBeVisible();

  // Type one character — on its own this would be delayed 300ms by the
  // keyup-changed debounce. Immediately dispatch the `search` event (what the
  // browser fires when the user submits a type="search" input) before that
  // window elapses, and expect the request to be SENT well inside it. We
  // assert on request-send timing rather than response-receive timing because
  // the Spotify round trip itself can exceed the debounce window.
  await input.focus();
  await page.keyboard.type('b');

  const requestPromise = page.waitForRequest(
    (req) => req.url().includes('/app/library/discover/search') && req.url().includes('q=b'),
    { timeout: 250 },
  );
  const start = Date.now();
  await input.evaluate((el: HTMLInputElement) => {
    el.dispatchEvent(new Event('search'));
  });
  await requestPromise;
  const elapsed = Date.now() - start;
  expect(elapsed, 'search-event request must be sent before the 300ms debounce').toBeLessThan(300);
});

test('A library change refreshes the discover results panel', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');
  await expect(page.getByTestId('discover-page-results')).toBeVisible();

  // The results panel listens for `libraryUpdated` and `radarUpdated` on body
  // and re-fetches its contents. Dispatching the event mirrors what the
  // server does via HX-Trigger after a library- or radar-changing action.
  const responsePromise = page.waitForResponse((res) =>
    res.url().includes('/app/library/discover/search'),
  );
  await page.evaluate(() => {
    document.body.dispatchEvent(new CustomEvent('libraryUpdated', { bubbles: true }));
  });
  await responsePromise;

  // Re-render still leaves the results panel in place and showing the
  // empty-query content (since the search input has no value).
  await expect(page.getByTestId('discover-page-results')).toBeVisible();
  await expect(page.getByTestId('discover-search-results-empty-query')).toBeVisible();
});

test('A new search result swaps into the same panel and leaves the rest of the page intact', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  // Snapshot the radar above the search bar — its DOM should not be touched
  // by the search swap, only the results panel below should change.
  const radarBefore = await page.getByTestId('radar-carousel').innerHTML();

  await page.getByTestId('discover-page-search-input').pressSequentially('beatles');
  await expect(page.getByTestId('discover-search-results')).toBeVisible();

  // Results landed inside the dedicated panel.
  const panel = page.getByTestId('discover-page-results');
  await expect(panel.getByTestId('discover-search-results')).toBeVisible();

  // The radar above is byte-for-byte the same — unaffected by the swap.
  const radarAfter = await page.getByTestId('radar-carousel').innerHTML();
  expect(radarAfter, 'radar above the search bar must be unaffected by the search swap').toBe(radarBefore);
});

test('A new search result scrolls the results panel back to the top', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  const input = page.getByTestId('discover-page-search-input');
  await input.pressSequentially('beatles');
  await expect(page.getByTestId('discover-search-results')).toBeVisible();
  await expect(page.getByTestId('discover-search-result-row').first()).toBeVisible();

  const panel = page.getByTestId('discover-page-results');

  // Force the panel into an overflowing layout for the duration of this test
  // so its scrollTop can move off zero. The hx-swap "scroll:#discover-results:top"
  // directive in the running page targets this element; we just need the
  // element to have somewhere to scroll back from.
  await panel.evaluate((el) => {
    (el as HTMLElement).style.height = '200px';
    (el as HTMLElement).style.overflowY = 'auto';
  });

  await panel.evaluate((el) => { el.scrollTop = el.scrollHeight; });
  const scrolled = await panel.evaluate((el) => el.scrollTop);
  expect(scrolled, 'panel must be scrollable for this assertion to be meaningful').toBeGreaterThan(0);

  // Run a new search by appending more characters; wait for the swap to land.
  const responsePromise = page.waitForResponse((res) =>
    res.url().includes('/app/library/discover/search') && res.url().includes('q=beatles%20s'),
  );
  await input.pressSequentially(' s');
  await responsePromise;
  await expect(page.getByTestId('discover-search-result-row').first()).toBeVisible();

  await expect.poll(
    async () => panel.evaluate((el) => el.scrollTop),
    { message: 'panel must be scrolled to top after the swap' },
  ).toBe(0);
});

test('Erasing the query back to empty restores the same empty-query message as a fresh visit', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/discover');

  // Capture the fresh-page empty-query placeholder so we can compare it to
  // what reappears after the round trip (type → erase).
  const emptyOnLoad = page.getByTestId('discover-search-results-empty-query');
  await expect(emptyOnLoad).toBeVisible();
  const freshHTML = await emptyOnLoad.innerHTML();

  const input = page.getByTestId('discover-page-search-input');
  const query = 'beatles';
  await input.pressSequentially(query);
  await expect(page.getByTestId('discover-search-results')).toBeVisible();

  // Erase to empty via real backspaces so htmx's keyup-changed trigger fires
  // and the server returns the empty-query branch.
  await input.focus();
  for (let i = 0; i < query.length; i++) {
    await page.keyboard.press('Backspace');
  }
  await expect(input).toHaveValue('');

  const emptyAgain = page.getByTestId('discover-search-results-empty-query');
  await expect(emptyAgain).toBeVisible();
  const afterHTML = await emptyAgain.innerHTML();
  expect(afterHTML, 'empty-query content after clearing must match the fresh-load content').toBe(freshHTML);
});
