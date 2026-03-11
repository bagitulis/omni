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
    mockClientPost.mockReset();
    mockGet.mockReset();
    mockPost.mockReset();
    mockLogout.mockReset();
    mockLogout.mockResolvedValue(undefined);
  });

  // --- login ---
  describe("login", () => {
    it("returns login response on success", async () => {
      const mockData = {
        success: true,
        data: {
          token: "tok",
          access_token: "acc",
          user: { id: "u1" },
          tenant_id: "t1",
          expires_in: 3600,
        },
      };
      mockPost.mockResolvedValueOnce(mockData);

      const result = await login({ username: "admin", password: "secret" });

      expect(mockPost).toHaveBeenCalledWith(
        "/auth/login",
        { username: "admin", password: "secret" },
        { withCredentials: true },
      );
      expect(result).toEqual(mockData.data);
    });

    it("throws when success is false", async () => {
      mockPost.mockResolvedValueOnce({
        success: false,
        error: "Invalid credentials",
      });

      await expect(
        login({ username: "admin", password: "wrong" }),
      ).rejects.toThrow("Invalid credentials");
    });

    it("throws default message when error field is missing", async () => {
      mockPost.mockResolvedValueOnce({ success: false });

      await expect(login({})).rejects.toThrow("Login failed");
    });

    it("throws when wrapped login payload is missing token fields", async () => {
      mockPost.mockResolvedValueOnce({
        success: true,
        data: { user: { id: "u1" }, tenant_id: "t1", expires_in: 3600 },
      });

      await expect(
        login({ username: "admin", password: "secret" }),
      ).rejects.toThrow("Login failed");
    });
  });

  // --- devLogin ---
  describe("devLogin", () => {
    it("returns dev login response on success", async () => {
      const mockData = {
        success: true,
        dev_mode: true,
        token: "devtok",
        access_token: "devtok",
        user: { id: "dev-user" },
        tenant_id: "dev-tenant",
        expires_in: 28800,
      };
      mockClientPost.mockResolvedValueOnce({ data: mockData });

      const result = await devLogin({ tenant_id: "dev-tenant" });

      expect(mockClientPost).toHaveBeenCalledWith(
        "/auth/dev-login",
        { tenant_id: "dev-tenant" },
        { withCredentials: true },
      );
      expect(result).toEqual(mockData);
    });

    it("throws on failure", async () => {
      mockClientPost.mockResolvedValueOnce({
        data: {
          success: false,
          error: "Dev login disabled",
        },
      });

      await expect(devLogin({ tenant_id: "t1" })).rejects.toThrow(
        "Dev login disabled",
      );
    });

    it("throws default message when error field is missing", async () => {
      mockClientPost.mockResolvedValueOnce({ data: { success: false } });

      await expect(devLogin({ tenant_id: "t1" })).rejects.toThrow(
        "Dev login failed",
      );
    });

    it("throws when dev login payload is missing token fields", async () => {
      mockClientPost.mockResolvedValueOnce({
        data: {
          success: true,
          dev_mode: true,
          user: { id: "dev-user" },
          tenant_id: "dev-tenant",
        },
      });

      await expect(devLogin({ tenant_id: "dev-tenant" })).rejects.toThrow(
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

    it("returns null for 401 auth errors", async () => {
      mockGet.mockRejectedValueOnce(
        Object.assign(new Error("Unauthorized"), {
          isAxiosError: true,
          response: { status: 401 },
        }),
      );

      const result = await getCurrentUser();
      expect(result).toBeNull();
    });

    it("rethrows unexpected request errors", async () => {
      mockGet.mockRejectedValueOnce(new Error("Network error"));

      await expect(getCurrentUser()).rejects.toThrow("Network error");
    });
  });

  // --- changePassword ---
  describe("changePassword", () => {
    it("resolves without error on success", async () => {
      mockPost.mockResolvedValueOnce({ success: true });

      await expect(
        changePassword("oldpass", "newpass"),
      ).resolves.toBeUndefined();
      expect(mockPost).toHaveBeenCalledWith("/auth/change-password", {
        current_password: "oldpass",
        new_password: "newpass",
      });
    });

    it("throws when success is false", async () => {
      mockPost.mockResolvedValueOnce({
        success: false,
        error: "Incorrect password",
      });

      await expect(changePassword("wrong", "newpass")).rejects.toThrow(
        "Incorrect password",
      );
    });

    it("throws default message when error field is missing", async () => {
      mockPost.mockResolvedValueOnce({ success: false });

      await expect(changePassword("wrong", "newpass")).rejects.toThrow(
        "Failed to change password",
      );
    });
  });
});
