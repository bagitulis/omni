import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { UnifiedBatchBar } from "./UnifiedBatchBar";
import { BatchActionItem } from "@/types/shared";
import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import "@testing-library/jest-dom";

describe("UnifiedBatchBar", () => {
  const mockActions: BatchActionItem[] = [
    {
      key: "update_price",
      label: "Update Price",
      icon: <EditOutlined />,
    },
    {
      key: "delete_products",
      label: "Delete",
      icon: <DeleteOutlined />,
      danger: true,
    },
    {
      key: "sync_stock",
      label: "Sync Stock",
      icon: <EditOutlined />,
      disabled: true,
    },
  ];

  const mockOnAction = vi.fn();
  const mockOnClear = vi.fn();

  it("should not render when selectedCount is 0", () => {
    render(
      <UnifiedBatchBar
        selectedCount={0}
        actions={mockActions}
        onAction={mockOnAction}
        onClear={mockOnClear}
      />,
    );
    expect(screen.queryByRole("toolbar")).toBeNull();
    expect(screen.queryByText("0 Selected")).toBeNull();
  });

  it("should render when selectedCount > 0", () => {
    render(
      <UnifiedBatchBar
        selectedCount={5}
        actions={mockActions}
        onAction={mockOnAction}
        onClear={mockOnClear}
      />,
    );
    expect(
      screen.getByRole("toolbar", { name: "Batch Actions" }),
    ).toBeInTheDocument();
    expect(screen.getByText("5 Selected")).toBeInTheDocument();
  });

  it("should display all actions", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        actions={mockActions}
        onAction={mockOnAction}
        onClear={mockOnClear}
      />,
    );
    // Button names include icon labels if present
    expect(
      screen.getByRole("button", { name: /Update Price/i }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Delete/i })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /Sync Stock/i }),
    ).toBeInTheDocument();
  });

  it("should call onAction with correct key when action button is clicked", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        actions={mockActions}
        onAction={mockOnAction}
        onClear={mockOnClear}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Update Price/i }));
    expect(mockOnAction).toHaveBeenCalledWith("update_price");
  });

  it("should handle disabled actions", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        actions={mockActions}
        onAction={mockOnAction}
        onClear={mockOnClear}
      />,
    );
    const syncButton = screen.getByRole("button", { name: /Sync Stock/i });
    expect(syncButton).toBeDisabled();
    fireEvent.click(syncButton);
    expect(mockOnAction).not.toHaveBeenCalledWith("sync_stock");
  });

  it("should call onClear when clear button is clicked", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        actions={mockActions}
        onAction={mockOnAction}
        onClear={mockOnClear}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Clear/i }));
    expect(mockOnClear).toHaveBeenCalled();
  });

  it("should render danger button with correct class", () => {
    render(
      <UnifiedBatchBar
        selectedCount={1}
        actions={mockActions}
        onAction={mockOnAction}
        onClear={mockOnClear}
      />,
    );
    const deleteButton = screen.getByRole("button", { name: /Delete/i });
    expect(deleteButton).toHaveClass("ant-btn-dangerous");
  });
});
