import { test } from "@playwright/test";

// Developer context spec: impersonation context propagation
// Tests the developer impersonation feature — switching user/tenant context
// and verifying it propagates correctly across API calls and page navigation.

test.describe("developer context — impersonation", () => {
  test.skip("impersonating a user sets correct auth headers on API calls", async ({
    page,
  }) => {
    // TODO: Activate impersonation mode for a target user,
    // intercept subsequent API calls, verify X-Impersonate-User header is set
  });

  test.skip("impersonation context persists across page navigation", async ({
    page,
  }) => {
    // TODO: Start impersonation, navigate to multiple pages,
    // verify each page's API calls carry the impersonation context
  });

  test.skip("impersonation banner is visible while active", async ({ page }) => {
    // TODO: Activate impersonation, verify a visual indicator/banner
    // appears on all pages showing the impersonated user
  });

  test.skip("stopping impersonation restores original session", async ({
    page,
  }) => {
    // TODO: Start and then stop impersonation, verify API calls
    // revert to the original user's credentials
  });
});

test.describe("developer context — tenant switching", () => {
  test.skip("switching tenant updates API base context", async ({ page }) => {
    // TODO: Switch tenant via developer tools, verify subsequent API
    // calls include the new tenant_id in headers or query params
  });
});
