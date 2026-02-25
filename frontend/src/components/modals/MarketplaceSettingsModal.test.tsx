import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MarketplaceSettingsModal } from "@/components/modals/MarketplaceSettingsModal";
import {
  useInventoryConfig,
  useUpdateInventoryConfig,
} from "@/hooks/useInventory";

// Mock hooks
vi.mock("@/hooks/useInventory", () => ({
  useInventoryConfig: vi.fn(),
  useUpdateInventoryConfig: vi.fn(),
}));

// Mock antd
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Modal: ({ children, open, title, onCancel, footer }: any) =>
      open ? (
        <div data-testid="modal" role="dialog">
          <div data-testid="modal-title">{title}</div>
          <button onClick={onCancel}>Close</button>
          {children}
          <div data-testid="modal-footer">{footer}</div>
        </div>
      ) : null,
  };
});

describe("MarketplaceSettingsModal", () => {
  const mockOnClose = vi.fn();
  const mockUpdateConfig = vi.fn();
  const mockSchemaColumns = [
    { column_name: "SKU", column_type: "text" },
    { column_name: "Total Stock", column_type: "number" },
  ];

  beforeEach(() => {
    vi.resetAllMocks();
    (useInventoryConfig as any).mockReturnValue({
      data: {
        selected_columns: [],
      },
      isLoading: false,
    });
    (useUpdateInventoryConfig as any).mockReturnValue({
      mutate: mockUpdateConfig,
      isPending: false,
    });
  });

  const renderModal = () => {
    return render(
      <MarketplaceSettingsModal
        open={true}
        onClose={mockOnClose}
        schemaColumns={mockSchemaColumns}
      />,
    );
  };

  it("renders correctly when open", () => {
    renderModal();
    expect(screen.getByTestId("modal")).toBeInTheDocument();
    expect(screen.getByText("Key Column")).toBeInTheDocument();
    expect(screen.getByText("Column Mapping")).toBeInTheDocument();
    expect(screen.getByText("Allocation Ratios")).toBeInTheDocument();
  });

  it("renders buttons in footer", () => {
    renderModal();
    expect(screen.getByText("Reset Defaults")).toBeInTheDocument();
    expect(screen.getByText("Save Settings")).toBeInTheDocument();
  });

  it("shows preview calculation", () => {
    renderModal();
    expect(screen.getByText(/Shopee:/)).toBeInTheDocument();
    expect(screen.getByText(/TikTok:/)).toBeInTheDocument();
    expect(screen.getByText(/Lazada:/)).toBeInTheDocument();
    expect(screen.getByText(/Total:/)).toBeInTheDocument();
  });

  it("validates key column on save", async () => {
    renderModal();
    const saveBtn = screen.getByText("Save Settings");
    fireEvent.click(saveBtn);

    // Should show error message "Key column is required"
    // Since we mock message, we can check if it was called
    // But message.error is mocked as simple fn, we can't assert on DOM unless we mocked implementation
    // Instead, we can spy on message.error
    // But mocking antd in setup or here is required.
    // Assuming simple mock above doesn't capture calls unless we spy.

    // We can just check that mockUpdateConfig was NOT called if key column is missing
    expect(mockUpdateConfig).not.toHaveBeenCalled();
  });
});
