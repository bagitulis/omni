/**
 * Test data and credentials — loaded from environment variables.
 * NEVER hardcode credentials in this file.
 */

export const TEST_CREDENTIALS = {
  username: process.env.TEST_USERNAME || "",
  password: process.env.TEST_PASSWORD || "",
} as const;

export const TEST_BASE_URL = process.env.BASE_URL || "http://localhost:5174";

export const TEST_TIMEOUTS = {
  navigation: 10000,
  networkIdle: 15000,
  elementVisible: 5000,
} as const;

export const MOCK_RESPONSES = {
  shopeeConnected: {
    connected: true,
    shop_name: "Test Shop",
    status: "active",
  },
  shopeeDisconnected: { connected: false, status: "inactive" },
  lazadaConnected: {
    connected: true,
    shop_name: "Test Lazada Shop",
    status: "active",
  },
  tiktokConnected: {
    connected: true,
    shop_name: "Test TikTok Shop",
    status: "active",
  },
} as const;
