import { test, expect } from "@playwright/test";

/**
 * Full flow E2E tests — require API + UI running (e.g. docker compose up -d).
 * These validate navigation and list views after login.
 */
test.describe("Full flows (authenticated)", () => {
  test.setTimeout(60000); // Allow time for API calls on each page

  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    await page.getByPlaceholder(/admin@orchestra/i).fill("admin@orchestra.local");
    await page.getByPlaceholder(/••••••••/).fill("admin123");
    await page.getByRole("button", { name: /sign in/i }).click();
    await expect(page).toHaveURL("/");
  });

  test("Servers page shows list or empty state", async ({ page }) => {
    await page.getByRole("link", { name: /servers/i }).click();
    await expect(page).toHaveURL(/\/servers/);
    await expect(page.locator("main").getByRole("heading", { name: "Servers", exact: true })).toBeVisible({ timeout: 5000 });
    const hasTable = await page.locator("table").isVisible();
    const hasEmpty = await page.getByText(/no servers|add server|register/i).isVisible();
    expect(hasTable || hasEmpty).toBeTruthy();
  });

  test("Clusters page shows list or empty state", async ({ page }) => {
    await page.getByRole("link", { name: /clusters/i }).click();
    await expect(page).toHaveURL(/\/clusters/);
    await expect(page.locator("main").getByRole("heading", { name: "Clusters", exact: true })).toBeVisible({ timeout: 5000 });
    // Page has either cluster cards or empty state ("Create Cluster", "Open Designer", etc.)
    const hasContent = await page.locator("main").getByRole("button", { name: "Create Cluster" }).isVisible();
    expect(hasContent).toBeTruthy();
  });

  test("Applications page shows list or empty state", async ({ page }) => {
    await page.getByRole("link", { name: /applications/i }).click();
    await expect(page).toHaveURL(/\/applications/);
    await expect(page.locator("main").getByRole("heading", { name: /applications/i })).toBeVisible({ timeout: 5000 });
    const hasContent = await page.getByText(/application|deploy|new deployment/i).first().isVisible();
    expect(hasContent).toBeTruthy();
  });

  test("Deployments page shows list or empty state", async ({ page }) => {
    await page.getByRole("link", { name: /deployments/i }).click();
    await expect(page).toHaveURL(/\/deployments/);
    await expect(page.locator("main").getByRole("heading", { name: /deployments/i })).toBeVisible({ timeout: 5000 });
    const hasContent = await page.getByText(/deployment|no deployments/i).first().isVisible();
    expect(hasContent).toBeTruthy();
  });

  test("Environments page loads", async ({ page }) => {
    await page.getByRole("link", { name: /environments/i }).click();
    await expect(page).toHaveURL(/\/environments/);
    await expect(page.locator("main").getByRole("heading", { name: "Environments", exact: true })).toBeVisible({ timeout: 5000 });
  });

  test("Monitoring page loads", async ({ page }) => {
    await page.getByRole("link", { name: /monitoring/i }).click();
    await expect(page).toHaveURL(/\/monitoring/);
    await expect(page.locator("main").getByRole("heading", { name: /monitoring/i })).toBeVisible({ timeout: 5000 });
  });

  test("Settings page loads", async ({ page }) => {
    await page.getByRole("link", { name: /settings/i }).click();
    await expect(page).toHaveURL(/\/settings/);
    await expect(page.locator("main").getByRole("heading", { name: "Settings", exact: true })).toBeVisible({ timeout: 10000 });
  });
});
