import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import SimplifiedInventoryPage from "./SimplifiedInventoryPage";

// Mock dependencies
vi.mock("@/hooks/useInventory", () => ({
  useInventory: vi.fn(() => ({
    data: { records: [{ id: 1, product_name: "Test Product" }], total: 10 },
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  })),
  useInventoryConfig: vi.fn(() => ({ data: {} })),
  useSyncFromSheets: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
  useSyncToSheets: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

vi.mock("@/stores/inventoryFilterStore", () => ({
  useInventoryFilterStore: vi.fn(() => ({
    search: "",
    platform: [],
    stockStatus: [],
    syncStatus: [],
    page: 1,
    pageSize: 20,
    setSearch: vi.fn(),
    setPage: vi.fn(),
    setPageSize: vi.fn(),
  })),
}));

// Mock the zustand store's getState method for the lock panel
vi.mock("@/stores/inventoryFilterStore", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/stores/inventoryFilterStore")>();
  return {
    ...actual,
    useInventoryFilterStore: Object.assign(
      vi.fn(() => ({
        search: "",
        platform: [],
        stockStatus: [],
        syncStatus: [],
        page: 1,
        pageSize: 20,
        setSearch: vi.fn(),
        setPage: vi.fn(),
        setPageSize: vi.fn(),
      })),
      {
        getState: () => ({
          setLockedColumns: vi.fn(),
        }),
      },
    ),
  };
});

vi.mock("./utils/inventoryColumnFilters", () => ({
  applyInventoryColumnFilters: vi.fn((records) => records),
}));

vi.mock("./utils/marketplaceAllocation", () => ({
  deriveMarketplaceAllocationSettings: vi.fn(() => ({})),
}));

vi.mock("./hooks/useInventoryColumns", () => ({
  useInventoryColumns: vi.fn(() => ({
    availableColumns: ["product_name", "sku"],
    resolvedVisibleColumns: ["product_name", "sku"],
    resolvedLockedColumns: [],
    columnConfigs: {},
    handleColumnChange: vi.fn(),
    handleColumnReset: vi.fn(),
    columnFilters: {},
  })),
}));

vi.mock("./components/SimplifiedInventoryHeader", () => ({
  SimplifiedInventoryHeader: ({ onSearch }: { onSearch: (val: string) => void }) => (
    <div data-testid="inventory-header">
      <input placeholder="Search" onChange={(e) => onSearch(e.target.value)} />
    </div>
  ),
}));

vi.mock("./components/InventoryMainTab", () => ({
  InventoryMainTab: () => (
    <div data-testid="inventory-main-tab">Inventory Table</div>
  ),
}));

vi.mock("./components/InventoryLockPanel", () => ({
  InventoryLockPanel: () => (
    <div data-testid="inventory-lock-panel">Lock Panel</div>
  ),
}));

vi.mock("./components/InventoryPagination", () => ({
  InventoryPagination: ({
    onChange,
  }: {
    onChange: (page: number, size: number) => void;
  }) => (
    <div data-testid="inventory-pagination">
      <button onClick={() => onChange(2, 20)}>Next Page</button>
    </div>
  ),
}));

vi.mock("./components/SyncHistoryTab", () => ({
  SyncHistoryTab: () => <div data-testid="sync-history-tab">Sync History</div>,
}));

// Mock Ant Design
vi.mock("antd", async () => {
  const actual = await vi.importActual("antd");
  return {
    ...actual,
    Grid: {
      useBreakpoint: () => ({ md: true }),
    },
    theme: {
      useToken: () => ({
        token: { colorBgContainer: "#fff" },
      }),
    },
    Tabs: ({
      items,
      onChange,
    }: {
      items: { key: string; label: string; children: React.ReactNode }[];
      onChange: (key: string) => void;
    }) => (
      <div>
        {items.map((item) => (
          <button key={item.key} onClick={() => onChange(item.key)}>
            {item.label}
          </button>
        ))}
        {items.find((item) => item.key === "inventory")?.children}
      </div>
    ),
  };
});

describe("SimplifiedInventoryPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Mock matchMedia
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: vi.fn().mockImplementation((query) => ({
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
  });

  it("renders main components", () => {
    render(<SimplifiedInventoryPage />);
    expect(screen.getByTestId("inventory-header")).toBeInTheDocument();
    expect(screen.getByTestId("inventory-lock-panel")).toBeInTheDocument();
    expect(screen.getByTestId("inventory-main-tab")).toBeInTheDocument();
    expect(screen.getByTestId("inventory-pagination")).toBeInTheDocument();
  });

  it("updates search filter", async () => {
    const { useInventoryFilterStore } =
      await import("@/stores/inventoryFilterStore");
    const setSearch = vi.fn();
    (useInventoryFilterStore as unknown as Mock).mockImplementation(() => ({
      search: "",
      platform: [],
      stockStatus: [],
      syncStatus: [],
      page: 1,
      pageSize: 20,
      setSearch,
      setPage: vi.fn(),
      setPageSize: vi.fn(),
    }));

    render(<SimplifiedInventoryPage />);
    const input = screen.getByPlaceholderText("Search");
    fireEvent.change(input, { target: { value: "test" } });
    expect(setSearch).toHaveBeenCalledWith("test");
  });

  it("handles pagination change", async () => {
    const { useInventoryFilterStore } =
      await import("@/stores/inventoryFilterStore");
    const setPage = vi.fn();
    (useInventoryFilterStore as unknown as Mock).mockImplementation(() => ({
      search: "",
      platform: [],
      stockStatus: [],
      syncStatus: [],
      page: 1,
      pageSize: 20,
      setSearch: vi.fn(),
      setPage,
      setPageSize: vi.fn(),
    }));

    render(<SimplifiedInventoryPage />);
    fireEvent.click(screen.getByText("Next Page"));
    expect(setPage).toHaveBeenCalledWith(2);
  });
});
