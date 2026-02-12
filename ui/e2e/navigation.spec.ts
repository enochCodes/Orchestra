import { test, expect } from "@playwright/test";

test.describe("Navigation (authenticated)", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    await page.getByPlaceholder(/admin@orchestra/i).fill("admin@orchestra.local");
    await page.getByPlaceholder(/••••••••/).fill("admin123");
    await page.getByRole("button", { name: /sign in/i }).click();
    await expect(page).toHaveURL("/");
  });

  test("should navigate to Servers", async ({ page }) => {
    await page.getByRole("link", { name: /servers/i }).click();
    await expect(page).toHaveURL(/\/servers/);
  });

  test("should navigate to Clusters", async ({ page }) => {
    await page.getByRole("link", { name: /clusters/i }).click();
    await expect(page).toHaveURL(/\/clusters/);
  });

  test("should navigate to Applications", async ({ page }) => {
    await page.getByRole("link", { name: /applications/i }).click();
    await expect(page).toHaveURL(/\/applications/);
  });
});
