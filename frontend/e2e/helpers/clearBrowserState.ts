import { Page } from "@playwright/test";

/**
 * Clear browser state between tests: localStorage, sessionStorage, and any
 * cached auth tokens the app persists in the browser.
 *
 * Does NOT touch the backend database — spec files that need DB isolation
 * must mock via `page.route(...)` or seed via a dedicated helper. Naming this
 * helper `clearBrowserState` (not `resetTestState` or `dbReset`) keeps that
 * contract obvious at every call site.
 */
export async function clearBrowserState(page: Page): Promise<void> {
  await page.evaluate(() => {
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
