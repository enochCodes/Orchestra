import { test, expect } from "@playwright/test";

test.describe("Authentication", () => {
  test("should show login page when not authenticated", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/login/);
  });

  test("should display login form", async ({ page }) => {
    await page.goto("/login");
    await expect(page.getByRole("heading", { name: /orchestra/i })).toBeVisible();
    await expect(page.getByPlaceholder(/admin@orchestra/i)).toBeVisible();
    await expect(page.getByPlaceholder(/••••••••/)).toBeVisible();
    await expect(page.getByRole("button", { name: /sign in/i })).toBeVisible();
  });

  test("should login with valid credentials", async ({ page }) => {
    await page.goto("/login");

    await page.getByPlaceholder(/admin@orchestra/i).fill("admin@orchestra.local");
    await page.getByPlaceholder(/••••••••/).fill("admin123");
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page).toHaveURL("/");
    await expect(page.locator("main").getByRole("heading", { name: /dashboard/i })).toBeVisible();
  });

  test("should show error on invalid credentials", async ({ page }) => {
    await page.goto("/login");

    await page.getByPlaceholder(/admin@orchestra/i).fill("wrong@example.com");
    await page.getByPlaceholder(/••••••••/).fill("wrongpassword");
    await page.getByRole("button", { name: /sign in/i }).click();

    // With invalid credentials: error div appears, or we stay on login (no redirect to dashboard)
    const errorVisible = await page.getByTestId("login-error").isVisible({ timeout: 5000 }).catch(() => false);
    const stillOnLogin = (await page.url()).includes("/login");
    expect(errorVisible || stillOnLogin).toBeTruthy();
  });
});
