import { render, screen, fireEvent } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { ProductClassificationPage } from "./ProductClassificationPage";

// ---------------------------------------------------------------------------
// Browser API stubs required by Ant Design
// ---------------------------------------------------------------------------
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((q: string) => ({
    matches: false,
    media: q,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------
const mockSetSearchParams = vi.fn();
const mockSearchParams = new URLSearchParams();

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useSearchParams: () => [mockSearchParams, mockSetSearchParams],
  };
});

const mockRefetch = vi.fn();

vi.mock("@/hooks/useAnalyticsIntelligence", () => ({
  useClassifiedProducts: () => ({
    data: {
      scale_up: [
        {
          product_id: "p1",
          product_name: "Product Scale Up",
          action_label: "Scale Up",
          roas: 6.0,
        },
      ],
      maintain: [
        {
          product_id: "p2",
          product_name: "Product Maintain",
          action_label: "Maintain",
          roas: 3.0,
        },
      ],
      reduce: [],
      stop: [],
    },
    isLoading: false,
    refetch: mockRefetch,
  }),
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("ProductClassificationPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams.delete("tab");
  });

  it("renders title and default scale up products", () => {
    render(<ProductClassificationPage />);
    expect(screen.getByText("Product Classification")).toBeInTheDocument();
    expect(screen.getByText("Scale Up (1)")).toBeInTheDocument();
    expect(screen.getByText("Product Scale Up")).toBeInTheDocument();
  });

  it("renders maintain products when tab switched", () => {
    // Override search params for this test case or simulate click
    // Simulating click on Maintain summary card
    render(<ProductClassificationPage />);

    const maintainCard = screen.getByText("Maintain");
    fireEvent.click(maintainCard);

    expect(mockSetSearchParams).toHaveBeenCalledWith(
      { tab: "maintain" },
      { replace: true },
    );
  });

  it("handles refresh", () => {
    render(<ProductClassificationPage />);
    const refreshBtn = screen.getByText("Refresh");
    fireEvent.click(refreshBtn);
    expect(mockRefetch).toHaveBeenCalled();
  });
});
