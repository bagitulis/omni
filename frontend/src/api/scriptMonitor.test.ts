import { beforeEach, describe, expect, it, vi } from "vitest";
import { getAutoFunctions, getMonitorData } from "./scriptMonitor";

const { mockClientGet } = vi.hoisted(() => ({
  mockClientGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockClientGet,
  },
}));

describe("getAutoFunctions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns configs array when backend wraps configs", async () => {
    mockClientGet.mockResolvedValue({
      success: true,
      data: {
        configs: [
          {
            id: 1,
            name: "locked_today",
            enabled: true,
            interval_minutes: 1440,
          },
        ],
        total: 1,
      },
    });

    const result = await getAutoFunctions();

    expect(result).toHaveLength(1);
    expect(result[0].name).toBe("locked_today");
  });

  it("returns array when backend returns array directly", async () => {
    mockClientGet.mockResolvedValue({
      success: true,
      data: [
        {
          id: 2,
          name: "sync_from_sheets",
          enabled: true,
          interval_minutes: 30,
        },
      ],
    });

    const result = await getAutoFunctions();

    expect(result).toHaveLength(1);
    expect(result[0].name).toBe("sync_from_sheets");
  });

  it("returns empty array when configs missing", async () => {
    mockClientGet.mockResolvedValue({
      success: true,
      data: {
        total: 0,
      },
    });

    const result = await getAutoFunctions();

    expect(result).toEqual([]);
  });

  it("throws when backend returns error", async () => {
    mockClientGet.mockResolvedValue({
      success: false,
      error: "auto-functions failed",
    });

    await expect(getAutoFunctions()).rejects.toThrow("auto-functions failed");
  });
});

describe("getMonitorData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns monitor data when success", async () => {
    mockClientGet.mockResolvedValue({
      success: true,
      data: {
        current_job: null,
        pending_queue: [],
        recent_history: [],
        total_pending: 0,
        total_completed: 0,
      },
    });

    const result = await getMonitorData();

    expect(result.total_pending).toBe(0);
    expect(result.pending_queue).toEqual([]);
  });

  it("throws when monitor fetch fails", async () => {
    mockClientGet.mockResolvedValue({
      success: false,
      error: "monitor failed",
    });

    await expect(getMonitorData()).rejects.toThrow("monitor failed");
  });
});
