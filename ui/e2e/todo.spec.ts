import { expect, test } from "@playwright/test"

// todo: authentication

test("has title", async ({ page }) => {
    await page.goto("/")
    await expect(page).toHaveTitle(/Hexes/)
})

