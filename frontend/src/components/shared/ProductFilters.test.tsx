import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";
import { ProductFilters } from "./ProductFilters";
import type { ProductFilterValues } from "@/types/shared";
import type { InputProps, SelectProps, ButtonProps } from "antd";

// Mock Ant Design components to ensure robust testing without JSDOM/Antd issues
vi.mock("antd", async () => {
  const actual = await vi.importActual("antd");
  return {
    ...actual,
    Input: React.forwardRef(
      (props: InputProps, _ref: React.Ref<HTMLInputElement>) => (
        <div data-testid="mock-input-wrapper">
          <input
            data-testid="mock-input"
            value={props.value as string}
            onChange={
              props.onChange as React.ChangeEventHandler<HTMLInputElement>
            }
            placeholder={props.placeholder}
          />
        </div>
      ),
    ),
    Select: (props: SelectProps) => (
      <div data-testid="mock-select-wrapper">
        {/* Simulate click if needed */}
        <div
          data-testid="mock-select-display"
          onClick={() => {}}
          onKeyUp={() => {}}
          role="button"
          tabIndex={0}
        >
          {/* Render options as clickable elements for testing */}
          {props.options?.map((opt) => (
            <div
              key={opt.value as string}
              data-testid={`select-option-${opt.value}`}
              onClick={() => props.onChange?.(opt.value, opt)}
              onKeyUp={() => {}}
              role="option"
              aria-selected={props.value === opt.value}
              tabIndex={0}
            >
              {opt.label}
            </div>
          ))}
          <span data-testid="selected-value">{props.value}</span>
        </div>
      </div>
    ),
    Button: (props: ButtonProps) => (
      <button
        type={(props.htmlType as "button" | "submit" | "reset") || "button"}
        onClick={props.onClick}
        data-testid={props["data-testid"] || "mock-button"}
      >
        {props.children}
      </button>
    ),
    theme: {
      useToken: () => ({
        token: {
          colorTextQuaternary: "#ccc",
          colorTextSecondary: "#999",
        },
      }),
    },
  };
});

// Mock Icons
vi.mock("@ant-design/icons", () => ({
  SearchOutlined: () => <span data-testid="icon-search" />,
  FilterOutlined: () => <span data-testid="icon-filter" />,
  ClearOutlined: () => <span data-testid="icon-clear" />,
}));

describe("ProductFilters", () => {
  const defaultValues: ProductFilterValues = {
    search: "",
    platform: "all",
    status: "all",
    category: "all",
  };

  const mockOnChange = vi.fn();

  beforeEach(() => {
    mockOnChange.mockClear();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders all filter inputs", () => {
    render(
      <ProductFilters
        values={defaultValues}
        onChange={mockOnChange}
        showCategory={true}
      />,
    );

    expect(screen.getByTestId("mock-input")).toBeTruthy();

    // There are multiple 'all' options (Platform, Status, Category)
    const allOptions = screen.getAllByTestId("select-option-all");
    expect(allOptions.length).toBeGreaterThanOrEqual(2);

    // Status specific option
    expect(screen.getByTestId("select-option-active")).toBeTruthy();

    // Platform specific option
    expect(screen.getByTestId("select-option-shopee")).toBeTruthy();
  });

  it("hides category filter when showCategory is false", () => {
    render(
      <ProductFilters
        values={defaultValues}
        onChange={mockOnChange}
        showCategory={false}
      />,
    );

    // If category is hidden, we expect fewer selects
    const selects = screen.getAllByTestId("mock-select-wrapper");
    // Platform + Status = 2
    expect(selects.length).toBe(2);
  });

  it("calls onChange with debounced search input", () => {
    render(
      <ProductFilters
        values={defaultValues}
        onChange={mockOnChange}
        showCategory={true}
      />,
    );

    const input = screen.getByTestId("mock-input");
    fireEvent.change(input, { target: { value: "test search" } });

    // Should not fire immediately
    expect(mockOnChange).not.toHaveBeenCalled();

    // Fast forward debounce time
    vi.advanceTimersByTime(500);

    expect(mockOnChange).toHaveBeenCalledWith({
      ...defaultValues,
      search: "test search",
    });
  });

  it("calls onChange when platform is selected", () => {
    render(
      <ProductFilters
        values={defaultValues}
        onChange={mockOnChange}
        showCategory={true}
      />,
    );

    // Find the specific option for Shopee in the Platform select
    const shopeeOption = screen.getByTestId("select-option-shopee");
    fireEvent.click(shopeeOption);

    expect(mockOnChange).toHaveBeenCalledWith({
      ...defaultValues,
      platform: "shopee",
    });
  });

  it("shows clear button only when filters are active", () => {
    const { rerender } = render(
      <ProductFilters values={defaultValues} onChange={mockOnChange} />,
    );

    expect(screen.queryByTestId("clear-filters-btn")).toBeNull();

    // Activate search
    rerender(
      <ProductFilters
        values={{ ...defaultValues, search: "iphone" }}
        onChange={mockOnChange}
      />,
    );
    expect(screen.getByTestId("clear-filters-btn")).toBeTruthy();

    // Activate platform
    rerender(
      <ProductFilters
        values={{ ...defaultValues, platform: "lazada" }}
        onChange={mockOnChange}
      />,
    );
    expect(screen.getByTestId("clear-filters-btn")).toBeTruthy();
  });

  it("resets all filters when clear button is clicked", () => {
    const activeValues: ProductFilterValues = {
      search: "iphone",
      platform: "lazada",
      status: "active",
      category: "electronics",
    };

    render(<ProductFilters values={activeValues} onChange={mockOnChange} />);

    const clearBtn = screen.getByTestId("clear-filters-btn");
    fireEvent.click(clearBtn);

    expect(mockOnChange).toHaveBeenCalledWith({
      search: "",
      platform: "all",
      status: "all",
      category: "all",
    });
  });
});
