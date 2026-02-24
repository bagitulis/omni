import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  login,
  devLogin,
  logout,
  getCurrentUser,
  changePassword,
} from "./auth";

const { mockClientPost, mockGet, mockPost, mockLogout } = vi.hoisted(() => ({
  mockClientPost: vi.fn(),
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockLogout: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    client: {
      post: mockClientPost,
    },
  },
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: {
    getState: vi.fn().mockReturnValue({ logout: mockLogout }),
  },
}));

describe("auth API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLogout.mockResolvedValue(undefined);
  });

  // --- login ---
  describe("login", () => {
    it("returns login response on success", async () => {
      const mockData = {
        success: true,
        token: "tok",
        access_token: "acc",
        user: { id: "u1" },
        tenant_id: "t1",
        expires_in: 3600,
      };
      mockClientPost.mockResolvedValue({ data: mockData });

      const result = await login({ username: "admin", password: "secret" });

      expect(mockClientPost).toHaveBeenCalledWith(
        "/auth/login",
        { username: "admin", password: "secret" },
        { withCredentials: true },
      );
      expect(result).toEqual(mockData);
    });

    it("throws when success is false", async () => {
      mockClientPost.mockResolvedValue({
        data: { success: false, error: "Invalid credentials" },
      });

      await expect(
        login({ username: "admin", password: "wrong" }),
      ).rejects.toThrow("Invalid credentials");
    });

    it("throws default message when error field is missing", async () => {
      mockClientPost.mockResolvedValue({ data: { success: false } });

      await expect(login({})).rejects.toThrow("Login failed");
    });
  });

  // --- devLogin ---
  describe("devLogin", () => {
    it("returns dev login response on success", async () => {
      const mockData = {
        success: true,
        dev_mode: true,
        token: "devtok",
        user: { id: "dev-user" },
        tenant_id: "dev-tenant",
      };
      mockClientPost.mockResolvedValue({ data: mockData });

      const result = await devLogin({ tenant_id: "dev-tenant" });

      expect(mockClientPost).toHaveBeenCalledWith(
        "/auth/dev-login",
        { tenant_id: "dev-tenant" },
        { withCredentials: true },
      );
      expect(result).toEqual(mockData);
    });

    it("throws on failure", async () => {
      mockClientPost.mockResolvedValue({
        data: { success: false, error: "Dev login disabled" },
      });

      await expect(devLogin({ tenant_id: "t1" })).rejects.toThrow(
        "Dev login disabled",
      );
    });

    it("throws default message when error field is missing", async () => {
      mockClientPost.mockResolvedValue({ data: { success: false } });

      await expect(devLogin({ tenant_id: "t1" })).rejects.toThrow(
        "Dev login failed",
      );
    });
  });

  // --- logout ---
  describe("logout", () => {
    it("calls authStore logout", async () => {
      await logout();
      expect(mockLogout).toHaveBeenCalledOnce();
    });
  });

  // --- getCurrentUser ---
  describe("getCurrentUser", () => {
    it("returns user on success", async () => {
      const user = { id: "u1", username: "admin" };
      mockGet.mockResolvedValue({ success: true, data: user });

      const result = await getCurrentUser();
      expect(result).toEqual(user);
    });

    it("returns null when data is missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: undefined });

      const result = await getCurrentUser();
      expect(result).toBeNull();
    });

    it("returns null when request throws", async () => {
      mockGet.mockRejectedValue(new Error("Network error"));

      const result = await getCurrentUser();
      expect(result).toBeNull();
    });
  });

  // --- changePassword ---
  describe("changePassword", () => {
    it("resolves without error on success", async () => {
      mockPost.mockResolvedValue({ success: true });

      await expect(
        changePassword("oldpass", "newpass"),
      ).resolves.toBeUndefined();
      expect(mockPost).toHaveBeenCalledWith("/auth/change-password", {
        current_password: "oldpass",
        new_password: "newpass",
      });
    });

    it("throws when success is false", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Incorrect password",
      });

      await expect(changePassword("wrong", "newpass")).rejects.toThrow(
        "Incorrect password",
      );
    });

    it("throws default message when error field is missing", async () => {
      mockPost.mockResolvedValue({ success: false });

      await expect(changePassword("wrong", "newpass")).rejects.toThrow(
        "Failed to change password",
      );
    });
  });
});
