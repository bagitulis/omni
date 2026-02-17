import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { UnifiedBatchBar } from "./UnifiedBatchBar";
import "@testing-library/jest-dom";

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

  it("should display all seven batch actions", () => {
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
      screen.getByRole("button", { name: /^Wholesale$/i }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /MPQ/i })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /^Clone$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /Del Wholesale/i }),
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
    fireEvent.click(screen.getByRole("button", { name: /^Wholesale$/i }));
    fireEvent.click(screen.getByRole("button", { name: /MPQ/i }));
    fireEvent.click(screen.getByRole("button", { name: /^Clone$/i }));
    fireEvent.click(screen.getByRole("button", { name: /Del Wholesale/i }));
    fireEvent.click(screen.getByRole("button", { name: /^Delete$/i }));

    expect(mockOnAction).toHaveBeenCalledWith("sync_stock");
    expect(mockOnAction).toHaveBeenCalledWith("update_price");
    expect(mockOnAction).toHaveBeenCalledWith("wholesale");
    expect(mockOnAction).toHaveBeenCalledWith("mpq");
    expect(mockOnAction).toHaveBeenCalledWith("clone");
    expect(mockOnAction).toHaveBeenCalledWith("delete_wholesale");
    expect(mockOnAction).toHaveBeenCalledWith("delete_products");
    expect(mockOnAction).toHaveBeenCalledTimes(7);
  });

  it("should disable actions from disabledActions map and show tooltip reason", async () => {
    render(
      <UnifiedBatchBar
        selectedCount={2}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
        disabledActions={{ wholesale: "Wholesale is Shopee only" }}
      />,
    );

    const wholesaleButton = screen.getByRole("button", {
      name: /^Wholesale$/i,
    });
    expect(wholesaleButton).toBeDisabled();

    fireEvent.mouseEnter(wholesaleButton.parentElement as HTMLElement);
    expect(
      await screen.findByText("Wholesale is Shopee only"),
    ).toBeInTheDocument();

    fireEvent.click(wholesaleButton);
    expect(mockOnAction).not.toHaveBeenCalledWith("wholesale");
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

  it("should render danger style on delete actions", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        onAction={mockOnAction}
        onClearSelection={mockOnClearSelection}
      />,
    );

    const deleteWholesaleButton = screen.getByRole("button", {
      name: /^Del Wholesale$/i,
    });
    const deleteButton = screen.getByRole("button", { name: /^Delete$/i });

    expect(deleteWholesaleButton).toHaveClass("ant-btn-dangerous");
    expect(deleteButton).toHaveClass("ant-btn-dangerous");
  });
});
