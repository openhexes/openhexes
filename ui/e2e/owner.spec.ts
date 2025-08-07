import { expect, test } from "@playwright/test"

test.beforeEach(async ({ context }) => {
    await context.addCookies([
        {
            name: "hexes.auth.google",
            value: "owner",
            domain: "localhost",
            path: "/",
            sameSite: "Strict",
        },
    ])
})

test("has title", async ({ page }) => {
    page.on("console", (msg) => {
        console.log(msg.text())
    })

    await page.goto("/nope")
    await page.getByText("Requested page doesn't exist:").isVisible()
    await expect(page).toHaveTitle(/Hexes \| Page not found/)
})
