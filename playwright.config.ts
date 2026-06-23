import { defineConfig, devices } from '@playwright/test';
import { config } from 'dotenv';

config(); // load .env so vars are available to helpers

const port = process.env.PORT || '4691';

export default defineConfig({
  testDir: './e2e/spec',
  fullyParallel: false,
  // One worker: specs share a single SQLite DB with no per-test isolation, so
  // running spec files across parallel workers lets rating/library tests race on
  // the same user's state. Serial execution keeps the suite deterministic.
  workers: 1,
  retries: 0,
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    trace: 'on-first-retry',
    testIdAttribute: 'data-testid',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
