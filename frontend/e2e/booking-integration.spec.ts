import { test } from "@playwright/test";

// Booking integration spec: booking list, parent links, drawer navigation
// Tests the booking management UI including list rendering, navigation
// to parent entities, and drawer/detail panel interactions.

test.describe("booking integration — list", () => {
  test.skip("booking list renders seeded bookings", async ({ page }) => {
    // TODO: Mock /api/bookings with seeded data, navigate to /bookings,
    // verify table rows match seeded booking entries
  });

  test.skip("booking list filters by status", async ({ page }) => {
    // TODO: Seed bookings with different statuses, apply status filter,
    // verify only matching bookings are displayed
  });

  test.skip("booking list shows empty state when no bookings", async ({
    page,
  }) => {
    // TODO: Mock empty bookings response, verify empty state message is shown
  });
});

test.describe("booking integration — parent links", () => {
  test.skip("clicking order link navigates to order detail", async ({
    page,
  }) => {
    // TODO: Seed a booking with an associated order_sn, click the order link,
    // verify navigation to /orders/{order_sn}
  });

  test.skip("clicking platform link navigates to platform settings", async ({
    page,
  }) => {
    // TODO: Click the platform name in a booking row, verify navigation
    // to /settings/credentials or relevant platform page
  });
});

test.describe("booking integration — drawer navigation", () => {
  test.skip("clicking a booking row opens detail drawer", async ({ page }) => {
    // TODO: Click a booking row, verify detail drawer slides in with
    // booking information displayed
  });

  test.skip("drawer close button dismisses the drawer", async ({ page }) => {
    // TODO: Open booking drawer, click close button, verify drawer is hidden
  });

  test.skip("drawer shows booking metadata and actions", async ({ page }) => {
    // TODO: Open drawer for a seeded booking, verify status, timestamps,
    // and any available action buttons are rendered
  });
});
