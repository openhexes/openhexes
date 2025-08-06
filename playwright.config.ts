import { defineConfig, devices } from "@playwright/test"

export default defineConfig({
    testDir: "./ui/e2e",
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    use: {
        baseURL: "http://localhost:5174",
        trace: "on-first-retry",
    },
    projects: [
        {
            name: "chromium",
            use: { ...devices["Desktop Chrome"] },
        },
        {
            name: "firefox",
            use: { ...devices["Desktop Firefox"] },
        },
        {
            name: "webkit",
            use: { ...devices["Desktop Safari"] },
        },
    ],
    webServer: [
        {
            name: "vite",
            command: "pnpm run dev --port 5174",
            url: "http://localhost:5174",
            reuseExistingServer: false,
            env: {
                VITE_API_ADDRESS: "http://localhost:9090",
            },
        },
        {
            name: "gapi",
            command: "go run .",
            cwd: "./gapi",
            reuseExistingServer: false,
            url: "http://localhost:9090/ping",
            env: {
                MODE: "testing",
                SERVER__ADDRESS: "localhost:9090",
                SERVER__ALLOWED_ORIGINS: "http://localhost:5174",
                LOGGING__LEVEL: "1",
            },
            gracefulShutdown: {
                signal: "SIGINT",
                timeout: 10 * 1000,
            },
        },
    ],
})
