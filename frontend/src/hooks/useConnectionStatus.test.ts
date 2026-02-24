import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";

const setConnectionStatusMock = vi.fn();

vi.mock("@/stores/appStore", () => ({
  useAppStore: vi.fn(
    (
      selector: (state: {
        connectionStatus: string;
        isConnected: boolean;
        setConnectionStatus: typeof setConnectionStatusMock;
      }) => unknown,
    ) =>
      selector({
        connectionStatus: "connected",
        isConnected: true,
        setConnectionStatus: setConnectionStatusMock,
      }),
  ),
}));

const healthCheckMock = vi.fn();

vi.mock("@/api/client", () => ({
  default: {
    healthCheck: (...args: unknown[]) => healthCheckMock(...args),
  },
}));

import { useConnectionStatus } from "./useConnectionStatus";

describe("useConnectionStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    healthCheckMock.mockResolvedValue({ success: true });
  });

  it("returns connectionStatus from appStore", () => {
    const { result } = renderHook(() => useConnectionStatus());
    expect(result.current.connectionStatus).toBe("connected");
  });

  it("returns isConnected from appStore", () => {
    const { result } = renderHook(() => useConnectionStatus());
    expect(result.current.isConnected).toBe(true);
  });
});
