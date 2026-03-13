import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/login.feature

test('Unauthenticated user sees the login button', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByTestId('login-button')).toBeVisible();
});

test('Login button links to Spotify OAuth', async ({ page }) => {
  await page.goto('/');
  const href = await page.getByTestId('login-link').getAttribute('href');
  expect(href).toContain('accounts.spotify.com');
});

test('Login page has the app title', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle(/wax/i);
});

test('Authenticated user is redirected to the library', async ({ context, page }) => {
  const userId = process.env.E2E_TEST_USER_ID;
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto('/');
  await expect(page).toHaveURL('/app/library/dashboard');
});
