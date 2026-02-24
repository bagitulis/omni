import { Page } from "@playwright/test";

/**
 * Reset test state between tests.
 * Clears localStorage/sessionStorage to ensure clean state.
 * Does NOT reset database — for read-only E2E tests this is sufficient.
 */
export async function resetTestState(page: Page): Promise<void> {
  await page.evaluate(() => {
    // Clear auth tokens and cached data
    const keysToKeep: string[] = [];
    const keysToRemove = Object.keys(localStorage).filter(
      (k) => !keysToKeep.includes(k),
    );
    for (const k of keysToRemove) {
      localStorage.removeItem(k);
    }
    sessionStorage.clear();
  });
}

/**
 * Navigate to the base URL with clean state.
 */
export async function resetAndNavigate(
  page: Page,
  path: string = "/",
): Promise<void> {
  await resetTestState(page);
  await page.goto(`http://localhost:5174${path}`);
}
