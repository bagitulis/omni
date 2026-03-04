import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { UnifiedBatchBar } from "./UnifiedBatchBar";
import "@testing-library/jest-dom";

// Ant Design components use matchMedia internally — mock it for jsdom
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

describe("UnifiedBatchBar", () => {
  const mockOnAction = vi.fn();
  const mockOnClearSelection = vi.fn();

  beforeEach(() => {
    mockOnAction.mockReset();
    mockOnClearSelection.mockReset();
  });

  it("should not render when selectedCount is 0", () => {
    render(
      <UnifiedBatchBar
        selectedCount={0}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
      />,
    );
    expect(screen.queryByRole("toolbar")).toBeNull();
    expect(screen.queryByText("0 selected")).toBeNull();
  });

  it("should render when selectedCount > 0", () => {
    render(
      <UnifiedBatchBar
        selectedCount={5}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
      />,
    );
    expect(
      screen.getByRole("toolbar", { name: /Batch actions/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("5 selected")).toBeInTheDocument();
  });

  it("should display all batch actions", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
      />,
    );

    expect(
      screen.getByRole("button", { name: /Sync Stock/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /Update Price/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /Bulk Pricing/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /^Clone$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /^Delete$/i }),
    ).toBeInTheDocument();
  });

  it("should call onAction with each batch action key", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /Sync Stock/i }));
    fireEvent.click(screen.getByRole("button", { name: /Update Price/i }));
    fireEvent.click(screen.getByRole("button", { name: /Bulk Pricing/i }));
    fireEvent.click(screen.getByRole("button", { name: /^Clone$/i }));
    fireEvent.click(screen.getByRole("button", { name: /^Delete$/i }));

    expect(mockOnAction).toHaveBeenCalledWith("sync_stock");
    expect(mockOnAction).toHaveBeenCalledWith("update_price");
    expect(mockOnAction).toHaveBeenCalledWith("bulk_pricing");
    expect(mockOnAction).toHaveBeenCalledWith("clone");
    expect(mockOnAction).toHaveBeenCalledWith("delete_products");
    expect(mockOnAction).toHaveBeenCalledTimes(5);
  });

  it("should disable actions from disabledActions map and show tooltip reason", async () => {
    render(
      <UnifiedBatchBar
        selectedCount={2}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
        disabledActions={{ bulk_pricing: "No items selected with valid price" }}
      />,
    );

    const bulkPricingButton = screen.getByRole("button", {
      name: /Bulk Pricing/i,
    });
    expect(bulkPricingButton).toBeDisabled();

    fireEvent.mouseEnter(bulkPricingButton.parentElement as HTMLElement);
    expect(
      await screen.findByText("No items selected with valid price"),
    ).toBeInTheDocument();

    fireEvent.click(bulkPricingButton);
    expect(mockOnAction).not.toHaveBeenCalledWith("bulk_pricing");
  });

  it("should call onClearSelection when clear button is clicked", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Clear selection/i }));
    expect(mockOnClearSelection).toHaveBeenCalledTimes(1);
  });

  it("should render danger style on delete action", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
      />,
    );

    const deleteButton = screen.getByRole("button", { name: /^Delete$/i });
    expect(deleteButton).toHaveClass("ant-btn-dangerous");
  });
});
