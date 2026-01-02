import { defineConfig, devices } from "@playwright/test"

const webPort = 3000
const mockPort = 3999

export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  expect: {
    timeout: 10_000,
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  use: {
    baseURL: `http://127.0.0.1:${webPort}`,
    trace: "on-first-retry",
  },
  webServer: {
    command: `npm run dev`,
    port: webPort,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      NEXT_PUBLIC_API_BASE: `http://127.0.0.1:${mockPort}`,
    },
  },
})
