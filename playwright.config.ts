import { defineConfig, devices } from '@playwright/test';
import { config } from 'dotenv';

config(); // load .env so vars are available to helpers

export default defineConfig({
  testDir: './e2e/spec',
  fullyParallel: false,
  retries: 0,
  use: {
    baseURL: 'http://127.0.0.1:4691',
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
