import { test } from "@playwright/test";

// Credential lifecycle spec: connect/refresh/disconnect for 3 platforms
// Tests the full OAuth/API credential flow for Shopee, Lazada, and TikTok Shop.

const PLATFORMS = ["shopee", "lazada", "tiktok"] as const;

test.describe("credential lifecycle — connect", () => {
  for (const platform of PLATFORMS) {
    test.skip(`${platform} — connect credential flow`, async ({ page }) => {
      // TODO: Navigate to /settings/credentials, initiate connect for {platform},
      // mock OAuth callback, verify credential appears as connected
    });
  }
});

test.describe("credential lifecycle — refresh", () => {
  for (const platform of PLATFORMS) {
    test.skip(`${platform} — token refresh succeeds`, async ({ page }) => {
      // TODO: Seed an existing connected credential for {platform},
      // trigger refresh, mock token endpoint, verify updated expiry displayed
    });
  }
});

test.describe("credential lifecycle — disconnect", () => {
  for (const platform of PLATFORMS) {
    test.skip(`${platform} — disconnect removes credential`, async ({ page }) => {
      // TODO: Seed a connected credential, click disconnect, confirm dialog,
      // verify credential row shows disconnected state
    });
  }
});

test.describe("credential lifecycle — error handling", () => {
  test.skip("failed OAuth callback shows error message", async ({ page }) => {
    // TODO: Mock a failed OAuth callback, verify error toast or alert is shown
  });

  test.skip("expired token triggers refresh automatically", async ({ page }) => {
    // TODO: Seed an expired credential, perform an action that uses it,
    // verify auto-refresh is triggered and action succeeds
  });
});
