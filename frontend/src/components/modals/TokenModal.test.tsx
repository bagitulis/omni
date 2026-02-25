import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import TokenModal from "@/components/modals/TokenModal";
import {
  useAllTokenStatus,
  useRefreshToken,
  useRefreshAllTokens,
} from "@/hooks/useTokens";

// Mock hooks
vi.mock("@/hooks/useTokens", () => ({
  useAllTokenStatus: vi.fn(),
  useRefreshToken: vi.fn(),
  useRefreshAllTokens: vi.fn(),
}));

// Mock antd
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Modal: ({ children, open, title, onClose }: any) =>
      open ? (
        <div data-testid="modal" role="dialog">
          <div data-testid="modal-title">{title}</div>
          <button onClick={onClose}>Close</button>
          {children}
        </div>
      ) : null,
  };
});

describe("TokenModal", () => {
  const mockRefreshToken = vi.fn();
  const mockRefreshAll = vi.fn();
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.resetAllMocks();
    (useRefreshToken as any).mockReturnValue({
      mutate: mockRefreshToken,
      isPending: false,
    });
    (useRefreshAllTokens as any).mockReturnValue({
      mutate: mockRefreshAll,
      isPending: false,
    });
  });

  const renderModal = (statusData: any, loading = false, error: Error | null = null) => {
    (useAllTokenStatus as any).mockReturnValue({
      data: statusData,
      isLoading: loading,
      error: error,
    });

    render(<TokenModal open={true} onClose={mockOnClose} />);
  };

  it("renders loading state", () => {
    renderModal(null, true);
    expect(screen.getByText("Loading token status...")).toBeInTheDocument();
  });

  it("renders error state", () => {
    renderModal(null, false, new Error("API Error"));
    expect(screen.getByText("Failed to Load Token Status")).toBeInTheDocument();
    expect(screen.getByText("API Error")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    renderModal({}, false);
    expect(screen.getByText("No Tokens Found")).toBeInTheDocument();
  });

  it("renders token list", () => {
    const mockStatus = {
      shopee: {
        platform: "shopee",
        isValid: true,
        expiresAt: "2023-12-31T23:59:59Z",
        shopName: "Shopee Store",
      },
      tiktok: {
        platform: "tiktok",
        isValid: false,
        isExpired: true,
        needsRefresh: true,
      },
    };

    renderModal(mockStatus);

    expect(screen.getByText("Shopee")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();

    expect(screen.getByText("Tiktok")).toBeInTheDocument();
    expect(screen.getByText("Expired")).toBeInTheDocument();
  });

  it("calls refresh token action", () => {
    const mockStatus = {
      shopee: {
        platform: "shopee",
        isValid: true,
      },
    };
    renderModal(mockStatus);

    const refreshBtns = screen.getAllByText("Refresh");
    fireEvent.click(refreshBtns[0]);
    expect(mockRefreshToken).toHaveBeenCalledWith("shopee");
  });
});
