import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useOrdersLogic } from "./useOrdersLogic";

const setSearchParamsMock = vi.fn();
const navigateMock = vi.fn();
const cancelQueriesMock = vi.fn();
const refetchMock = vi.fn();
const syncActiveTabMock = vi.fn();

let currentSearchParams = new URLSearchParams();

vi.mock("react-router-dom", () => ({
  useSearchParams: () => [currentSearchParams, setSearchParamsMock],
  useParams: () => ({ platform: "shopee" }),
  useNavigate: () => navigateMock,
}));

vi.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({
    cancelQueries: cancelQueriesMock,
  }),
}));

vi.mock("@/hooks/useOrders", () => ({
  useOrders: () => ({
    data: { orders: [], total: 0 },
    isLoading: false,
    refetch: refetchMock,
  }),
}));

vi.mock("./useOrderSync", () => ({
  useOrderSync: () => ({
    isSyncing: false,
    syncActiveTab: syncActiveTabMock,
  }),
}));

vi.mock("./useOrderBulkActions", () => ({
  useOrderBulkActions: () => ({
    handleBulkShip: vi.fn(),
    handleBulkPrint: vi.fn(),
    handleRetryFailedPrint: vi.fn(),
    handleBulkCancel: vi.fn(),
    isShipping: false,
    isPrinting: false,
    isCancelling: false,
    shipProgress: null,
    printProgress: null,
    cancelProgress: null,
    shipResult: null,
    printResult: null,
    cancelResult: null,
  }),
}));

vi.mock("./useOrderSingleActions", () => ({
  useOrderSingleActions: () => ({
    handleSingleShip: vi.fn(),
    handleShipConfirm: vi.fn(),
    handleSingleCancel: vi.fn(),
    handleCancelConfirm: vi.fn(),
    handleSinglePrint: vi.fn(),
    handleViewDetails: vi.fn(),
    isSingleShipping: false,
  }),
}));

describe("useOrdersLogic", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    currentSearchParams = new URLSearchParams();
  });

  it("normalizes alias URL tab to canonical tab", async () => {
    currentSearchParams = new URLSearchParams("type=READY_TO_SHIP");

    const { result } = renderHook(() => useOrdersLogic());

    expect(result.current.state.activeTab).toBe("unprocess");
    await waitFor(() => {
      expect(setSearchParamsMock).toHaveBeenCalled();
    });

    const [nextParams, options] = setSearchParamsMock.mock.calls[0];
    expect(nextParams.get("type")).toBe("unprocess");
    expect(options).toEqual({ replace: true });
  });

  it("does not rewrite URL when tab is already canonical", () => {
    currentSearchParams = new URLSearchParams("type=unprocess");

    const { result } = renderHook(() => useOrdersLogic());

    expect(result.current.state.activeTab).toBe("unprocess");
    expect(setSearchParamsMock).not.toHaveBeenCalled();
  });
});
