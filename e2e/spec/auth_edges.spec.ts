import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/auth_edges.feature

const userId = process.env.E2E_TEST_USER_ID;

test('Logged-in user logs out and returns to the login page', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/app/library/dashboard');

  // Open the user menu in the header, then click the logout entry.
  await page.getByTestId('library-header-bar-user-menu').click();
  const logout = page.getByTestId('library-header-bar-logout');
  await expect(logout).toBeVisible();
  await logout.click();

  // Logout clears the JWT cookie and 307s to `/`. With no cookie, `/`
  // renders the login page rather than redirecting to the dashboard.
  await expect(page).toHaveURL('/');
  await expect(page.getByTestId('login-page-button')).toBeVisible();
});

test('Unauthenticated user is shown the unauthorized page on a protected route', async ({ page }) => {
  // No `loginAs` — visit a protected route with no cookie. The auth
  // middleware 303s to `/unauthorized`, which the browser then loads.
  await page.goto('/app/library/dashboard');

  await expect(page).toHaveURL('/unauthorized');
  await expect(page.getByTestId('unauthorized-page')).toBeVisible();
});
