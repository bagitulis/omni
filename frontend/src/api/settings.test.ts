import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  updateProfile,
  changePassword,
  saveGeneralSettings,
  saveWebhookConfig,
} from "./settings";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: vi.fn(),
    post: mockPost,
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

beforeEach(() => vi.clearAllMocks());

describe("updateProfile", () => {
  it("calls POST /auth/profile with payload", async () => {
    mockPost.mockResolvedValue({ success: true });

    await updateProfile({ name: "John Doe", email: "john@example.com" });

    expect(mockPost).toHaveBeenCalledWith("/auth/profile", {
      name: "John Doe",
      email: "john@example.com",
    });
  });

  it("resolves without error on success", async () => {
    mockPost.mockResolvedValue({ success: true });

    await expect(
      updateProfile({ name: "Test User", email: "test@example.com" }),
    ).resolves.toBeUndefined();
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "Email already in use",
    });

    await expect(
      updateProfile({ name: "Test User", email: "test@example.com" }),
    ).rejects.toThrow("Email already in use");
  });

  it("throws default message when error is missing", async () => {
    mockPost.mockResolvedValue({ success: false });

    await expect(
      updateProfile({ name: "Test", email: "test@example.com" }),
    ).rejects.toThrow("Failed to update profile");
  });
});

describe("changePassword", () => {
  it("sends only current_password and new_password (not confirm_password)", async () => {
    mockPost.mockResolvedValue({ success: true });

    await changePassword({
      current_password: "old-secret",
      new_password: "new-secret",
      confirm_password: "new-secret",
    });

    expect(mockPost).toHaveBeenCalledWith("/auth/change-password", {
      current_password: "old-secret",
      new_password: "new-secret",
    });

    // confirm_password must NOT be sent
    const body = mockPost.mock.calls[0][1];
    expect(body).not.toHaveProperty("confirm_password");
  });

  it("resolves without error on success", async () => {
    mockPost.mockResolvedValue({ success: true });

    await expect(
      changePassword({
        current_password: "old",
        new_password: "new",
        confirm_password: "new",
      }),
    ).resolves.toBeUndefined();
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "Current password is incorrect",
    });

    await expect(
      changePassword({
        current_password: "wrong",
        new_password: "new",
        confirm_password: "new",
      }),
    ).rejects.toThrow("Current password is incorrect");
  });

  it("throws default message when error is missing", async () => {
    mockPost.mockResolvedValue({ success: false });

    await expect(
      changePassword({
        current_password: "old",
        new_password: "new",
        confirm_password: "new",
      }),
    ).rejects.toThrow("Failed to change password");
  });
});

describe("saveGeneralSettings", () => {
  const payload = {
    language: "en",
    timezone: "UTC",
    notifications_email: true,
    notifications_browser: false,
    auto_sync: true,
    sync_interval: "1h",
  };

  it("calls POST /settings/general with payload", async () => {
    mockPost.mockResolvedValue({ success: true });

    await saveGeneralSettings(payload);

    expect(mockPost).toHaveBeenCalledWith("/settings/general", payload);
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Invalid timezone" });

    await expect(saveGeneralSettings(payload)).rejects.toThrow(
      "Invalid timezone",
    );
  });
});

describe("saveWebhookConfig", () => {
  it("calls POST /webhooks/config with payload", async () => {
    mockPost.mockResolvedValue({ success: true });

    await saveWebhookConfig({
      custom_url: "https://example.com/webhook",
      secret_key: "my-secret-key",
    });

    expect(mockPost).toHaveBeenCalledWith("/webhooks/config", {
      custom_url: "https://example.com/webhook",
      secret_key: "my-secret-key",
    });
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Invalid URL" });

    await expect(saveWebhookConfig({ custom_url: "bad-url" })).rejects.toThrow(
      "Invalid URL",
    );
  });

  it("throws default message when error is missing", async () => {
    mockPost.mockResolvedValue({ success: false });

    await expect(
      saveWebhookConfig({ custom_url: "https://example.com/wh" }),
    ).rejects.toThrow("Failed to save webhook configuration");
  });
});
