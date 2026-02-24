import "@testing-library/jest-dom";
import { render, screen, fireEvent } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ProductFilterValues, ColumnConfig } from "@/types/shared";
import { UnifiedProductsControls } from "./UnifiedProductsControls";

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

vi.mock("@/components/shared/ProductFilters", () => ({
  ProductFilters: ({
    onChange,
  }: {
    values: ProductFilterValues;
    onChange: (v: ProductFilterValues) => void;
  }) => (
    <button
      type="button"
      data-testid="product-filters"
      onClick={() => onChange({ search: "test" } as ProductFilterValues)}
    >
      Filters
    </button>
  ),
}));

vi.mock("@/components/shared/ColumnManager", () => ({
  ColumnManager: () => <div data-testid="column-manager">ColumnManager</div>,
}));

const defaultFilters: ProductFilterValues = {
  search: "",
  platform: "all",
  status: "all",
  category: "all",
  mapping: "all",
};

const mockColumns: ColumnConfig[] = [
  { key: "name", title: "Name", visible: true, order: 0 },
];

describe("UnifiedProductsControls", () => {
  const onFilterChange = vi.fn();
  const onViewModeChange = vi.fn();
  const onColumnsChange = vi.fn();
  const onResetColumns = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders List and Grid toggle buttons", () => {
    render(
      <UnifiedProductsControls
        filters={defaultFilters}
        onFilterChange={onFilterChange}
        viewMode="list"
        onViewModeChange={onViewModeChange}
        isMobile={false}
        columns={mockColumns}
        onColumnsChange={onColumnsChange}
        onResetColumns={onResetColumns}
      />,
    );
    expect(screen.getByRole("button", { name: /list/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /grid/i })).toBeInTheDocument();
  });

  it("calls onViewModeChange with 'grid' when Grid clicked", () => {
    render(
      <UnifiedProductsControls
        filters={defaultFilters}
        onFilterChange={onFilterChange}
        viewMode="list"
        onViewModeChange={onViewModeChange}
        isMobile={false}
        columns={mockColumns}
        onColumnsChange={onColumnsChange}
        onResetColumns={onResetColumns}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /grid/i }));
    expect(onViewModeChange).toHaveBeenCalledWith("grid");
  });

  it("calls onViewModeChange with 'list' when List clicked", () => {
    render(
      <UnifiedProductsControls
        filters={defaultFilters}
        onFilterChange={onFilterChange}
        viewMode="grid"
        onViewModeChange={onViewModeChange}
        isMobile={false}
        columns={mockColumns}
        onColumnsChange={onColumnsChange}
        onResetColumns={onResetColumns}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /list/i }));
    expect(onViewModeChange).toHaveBeenCalledWith("list");
  });

  it("renders ColumnManager when not mobile", () => {
    render(
      <UnifiedProductsControls
        filters={defaultFilters}
        onFilterChange={onFilterChange}
        viewMode="list"
        onViewModeChange={onViewModeChange}
        isMobile={false}
        columns={mockColumns}
        onColumnsChange={onColumnsChange}
        onResetColumns={onResetColumns}
      />,
    );
    expect(screen.getByTestId("column-manager")).toBeInTheDocument();
  });

  it("hides ColumnManager when mobile", () => {
    render(
      <UnifiedProductsControls
        filters={defaultFilters}
        onFilterChange={onFilterChange}
        viewMode="list"
        onViewModeChange={onViewModeChange}
        isMobile={true}
        columns={mockColumns}
        onColumnsChange={onColumnsChange}
        onResetColumns={onResetColumns}
      />,
    );
    expect(screen.queryByTestId("column-manager")).not.toBeInTheDocument();
  });

  it("renders ProductFilters", () => {
    render(
      <UnifiedProductsControls
        filters={defaultFilters}
        onFilterChange={onFilterChange}
        viewMode="list"
        onViewModeChange={onViewModeChange}
        isMobile={false}
        columns={mockColumns}
        onColumnsChange={onColumnsChange}
        onResetColumns={onResetColumns}
      />,
    );
    expect(screen.getByTestId("product-filters")).toBeInTheDocument();
  });

  it("calls onFilterChange when filters change", () => {
    render(
      <UnifiedProductsControls
        filters={defaultFilters}
        onFilterChange={onFilterChange}
        viewMode="list"
        onViewModeChange={onViewModeChange}
        isMobile={false}
        columns={mockColumns}
        onColumnsChange={onColumnsChange}
        onResetColumns={onResetColumns}
      />,
    );
    fireEvent.click(screen.getByTestId("product-filters"));
    expect(onFilterChange).toHaveBeenCalledWith({ search: "test" });
  });
});
