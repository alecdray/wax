import { BrowserContext } from '@playwright/test';
import { createHmac } from 'crypto';
import { config } from 'dotenv';
import { resolve } from 'path';

const _env = config({ path: resolve(__dirname, '../../.env') });

const JWT_COOKIE_NAME = 'wax_token';

function base64url(input: string | Buffer): string {
  const buf = typeof input === 'string' ? Buffer.from(input) : input;
  return buf.toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

/**
 * Build a signed HS256 JWT for the given userId using the app's JWT_SECRET.
 * Mirrors the Claims struct in src/internal/core/app/jwt.go.
 */
function buildJWT(userId: string, secret: string): string {
  const header = base64url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));

  const now = Math.floor(Date.now() / 1000);
  const payload = base64url(JSON.stringify({
    iss: 'wax',
    iat: now,
    nbf: now,
    exp: now + 86400, // 1 day, matching jwtTTL in jwt.go
    user_id: userId,
  }));

  const signature = base64url(
    createHmac('sha256', secret).update(`${header}.${payload}`).digest(),
  );

  return `${header}.${payload}.${signature}`;
}

/**
 * Inject a valid wax_token cookie into the browser context so that the app
 * treats subsequent requests as coming from the given user.
 *
 * The userId must exist in the database. Use this in tests that cover
 * authenticated pages to avoid going through the Spotify OAuth flow.
 *
 * Requires JWT_SECRET to be set in the environment (loaded from .env by
 * the test runner).
 *
 * @example
 * test('user sees their library', async ({ context, page }) => {
 *   await loginAs(context, 'user-uuid-here');
 *   await page.goto('/app/library/dashboard');
 * });
 */
export async function loginAs(context: BrowserContext, userId: string): Promise<void> {
  // Mirror the default from src/internal/core/app/config.go — in local env JWT_SECRET
  // falls back to "secret" when not explicitly set.
  const secret = _env.parsed?.JWT_SECRET ?? process.env.JWT_SECRET ?? 'secret';


  const token = buildJWT(userId, secret);

  await context.addCookies([{
    name: JWT_COOKIE_NAME,
    value: token,
    domain: '127.0.0.1',
    path: '/',
    httpOnly: true,
    secure: false,
    sameSite: 'Lax',
  }]);
}
