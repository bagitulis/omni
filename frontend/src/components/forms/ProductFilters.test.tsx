import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ProductFilters } from "@/components/forms/ProductFilters";

// Mock antd Segmented as it might be complex
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Segmented: ({
      options,
      onChange,
      value,
    }: {
      options?: Array<{ value: string; label?: string }>;
      onChange?: (val: string) => void;
      value?: string;
    }) => (
      <div data-testid="segmented">
        {(options ?? []).map((opt) => (
          <button
            key={opt.value}
            onClick={() => onChange?.(opt.value)}
            data-active={value === opt.value}
          >
            {opt.value}
          </button>
        ))}
      </div>
    ),
  };
});

describe("ProductFilters", () => {
  const mockOnFilterChange = vi.fn();
  const mockOnViewModeChange = vi.fn();
  const filters = {
    search: "test search",
    status: "active",
    platform: "shopee",
    category: "electronics",
  };

  const renderFilters = (props = {}) => {
    return render(
      <ProductFilters
        filters={filters}
        onFilterChange={mockOnFilterChange}
        viewMode="list"
        onViewModeChange={mockOnViewModeChange}
        {...props}
      />,
    );
  };

  it("renders filter inputs correctly", () => {
    renderFilters();
    expect(
      screen.getByPlaceholderText("Search products..."),
    ).toBeInTheDocument();
    expect(screen.getByDisplayValue("test search")).toBeInTheDocument();

    // Selects are tricky, but value should be displayed if we assume antd Select renders value in a span or input
    // In JSDOM, antd Select value is often found by text
    expect(screen.getByText("Shopee")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Electronics")).toBeInTheDocument();
  });

  it("calls onFilterChange when search input changes", () => {
    renderFilters();
    const searchInput = screen.getByPlaceholderText("Search products...");
    fireEvent.change(searchInput, { target: { value: "new search" } });
    expect(mockOnFilterChange).toHaveBeenCalledWith("search", "new search");
  });

  it("renders view mode toggles", () => {
    renderFilters();
    expect(screen.getByTestId("segmented")).toBeInTheDocument();
    expect(screen.getByText("list")).toBeInTheDocument();
    expect(screen.getByText("grid")).toBeInTheDocument();
  });

  it("calls onViewModeChange when toggled", () => {
    renderFilters();
    const gridBtn = screen.getByText("grid");
    fireEvent.click(gridBtn);
    expect(mockOnViewModeChange).toHaveBeenCalledWith("grid");
  });
});
