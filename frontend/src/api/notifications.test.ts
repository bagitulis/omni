import { beforeEach, describe, expect, it, vi } from "vitest";
import { notificationApi } from "./notifications";

const { mockAxiosCreate, mockGet, mockPost, mockPatch, mockDelete, mockPut } = vi.hoisted(() => ({
  mockAxiosCreate: vi.fn(() => ({
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  })),
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockPatch: vi.fn(),
  mockDelete: vi.fn(),
  mockPut: vi.fn(),
}));

vi.mock("axios", () => ({
  default: {
    create: mockAxiosCreate,
  },
}));

vi.mock("./client", () => ({
  apiClient: {
    get: mockGet,
    post: mockPost,
    patch: mockPatch,
    delete: mockDelete,
    put: mockPut,
  },
}));

describe("notificationApi", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("list", () => {
    it("returns notifications list with params", async () => {
      const mockData = {
        success: true,
        data: { items: [{ id: 1, type: "info", title: "Test" }], count: 1 },
      };
      mockGet.mockResolvedValue(mockData);

      const result = await notificationApi.list({ limit: 10 });

      expect(mockGet).toHaveBeenCalledWith("/notifications", {
        params: { limit: 10 },
      });
      expect(result).toEqual(mockData);
    });

    it("returns empty list when no params", async () => {
      mockGet.mockResolvedValue({ success: true, data: { items: [], count: 0 } });

      const result = await notificationApi.list();

      expect(mockGet).toHaveBeenCalledWith("/notifications", { params: {} });
      expect(result).toEqual({ success: true, data: { items: [], count: 0 } });
    });
  });


  describe("getUnreadCount", () => {
    it("returns unread count", async () => {
      mockGet.mockResolvedValue({ success: true, data: { unread_count: 5 } });

      const result = await notificationApi.getUnreadCount();

      expect(mockGet).toHaveBeenCalledWith("/notifications/unread-count");
      expect(result).toEqual({ success: true, data: { unread_count: 5 } });
    });
  });

  describe("markAsRead", () => {
    it("calls patch with correct endpoint", async () => {
      mockPatch.mockResolvedValue({ success: true });

      const result = await notificationApi.markAsRead(42);

      expect(mockPatch).toHaveBeenCalledWith("/notifications/42/read");
      expect(result).toEqual({ success: true });
    });
  });

  describe("markAllAsRead", () => {
    it("calls patch for read-all endpoint", async () => {
      mockPatch.mockResolvedValue({ success: true });

      const result = await notificationApi.markAllAsRead();

      expect(mockPatch).toHaveBeenCalledWith("/notifications/read-all");
      expect(result).toEqual({ success: true });
    });
  });

  describe("delete", () => {
    it("calls delete with correct id", async () => {
      mockDelete.mockResolvedValue({ success: true });

      const result = await notificationApi.delete(42);

      expect(mockDelete).toHaveBeenCalledWith("/notifications/42");
      expect(result).toEqual({ success: true });
    });
  });

  describe("deleteAll", () => {
    it("calls delete all endpoint", async () => {
      mockDelete.mockResolvedValue({ success: true });

      const result = await notificationApi.deleteAll();

      expect(mockDelete).toHaveBeenCalledWith("/notifications");
      expect(result).toEqual({ success: true });
    });
  });

  describe("getSettings", () => {
    it("returns notification settings", async () => {
      const mockSettings = { success: true, data: { retention_days: 30 } };
      mockGet.mockResolvedValue(mockSettings);

      const result = await notificationApi.getSettings();

      expect(mockGet).toHaveBeenCalledWith("/notifications/settings");
      expect(result).toEqual(mockSettings);
    });
  });

  describe("updateSettings", () => {
    it("calls put with settings payload", async () => {
      mockPut.mockResolvedValue({ success: true });

      const result = await notificationApi.updateSettings({ retention_days: 90 });

      expect(mockPut).toHaveBeenCalledWith("/notifications/settings", {
        retention_days: 90,
      });
      expect(result).toEqual({ success: true });
    });
  });
});
