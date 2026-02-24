import { describe, it, expect, vi, beforeEach } from "vitest";
import { getUserAuditLogs, getTenantAuditLogs, cleanupOldLogs } from "./audit";

const { mockGet, mockPost } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
  },
}));

describe("audit API", () => {
  beforeEach(() => vi.clearAllMocks());

  describe("getUserAuditLogs", () => {
    it("returns logs on success with default limit", async () => {
      const logs = [{ action: "login", user_id: "u1", status: "success" }];
      mockGet.mockResolvedValue({ success: true, data: { logs } });
      const result = await getUserAuditLogs();
      expect(mockGet).toHaveBeenCalledWith("/audit/logs/user?limit=50");
      expect(result).toEqual(logs);
    });

    it("returns logs with custom limit", async () => {
      const logs = [{ action: "logout", user_id: "u2", status: "success" }];
      mockGet.mockResolvedValue({ success: true, data: { logs } });
      const result = await getUserAuditLogs(10);
      expect(mockGet).toHaveBeenCalledWith("/audit/logs/user?limit=10");
      expect(result).toEqual(logs);
    });

    it("returns empty array when logs is null", async () => {
      mockGet.mockResolvedValue({ success: true, data: {} });
      const result = await getUserAuditLogs();
      expect(result).toEqual([]);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Access denied" });
      await expect(getUserAuditLogs()).rejects.toThrow("Access denied");
    });
  });

  describe("getTenantAuditLogs", () => {
    it("returns logs on success with default limit", async () => {
      const logs = [
        { action: "create_user", user_id: "admin", status: "success" },
      ];
      mockGet.mockResolvedValue({ success: true, data: { logs } });
      const result = await getTenantAuditLogs();
      expect(mockGet).toHaveBeenCalledWith("/audit/logs/tenant?limit=100");
      expect(result).toEqual(logs);
    });

    it("returns empty array when logs missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      const result = await getTenantAuditLogs();
      expect(result).toEqual([]);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Server error" });
      await expect(getTenantAuditLogs()).rejects.toThrow("Server error");
    });
  });

  describe("cleanupOldLogs", () => {
    it("resolves on success with default days", async () => {
      mockPost.mockResolvedValue({ success: true });
      await expect(cleanupOldLogs()).resolves.toBeUndefined();
      expect(mockPost).toHaveBeenCalledWith("/audit/cleanup", { days_old: 90 });
    });

    it("resolves with custom days", async () => {
      mockPost.mockResolvedValue({ success: true });
      await cleanupOldLogs(30);
      expect(mockPost).toHaveBeenCalledWith("/audit/cleanup", { days_old: 30 });
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({ success: false, error: "Cleanup failed" });
      await expect(cleanupOldLogs()).rejects.toThrow("Cleanup failed");
    });
  });
});
