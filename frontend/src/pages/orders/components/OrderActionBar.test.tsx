import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { OrderActionBar } from "./OrderActionBar";

// Mock the underlying OrderFilters component to keep test focused on the wrapper
vi.mock("@/components/forms/OrderFilters", () => ({
  OrderFilters: (props: Record<string, unknown>) => (
    <div data-testid="order-filters" data-loading={String(props.loading)} />
  ),
}));

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

describe("OrderActionBar", () => {
  function makeProps() {
    return {
      onSearch: vi.fn(),
      onPlatformChange: vi.fn(),
      onDateChange: vi.fn(),
      onRefresh: vi.fn(),
      onExport: vi.fn(),
      loading: false,
      autoRefresh: false,
      onAutoRefreshChange: vi.fn(),
    };
  }

  it("renders OrderFilters with passed props", () => {
    const props = makeProps();
    render(<OrderActionBar {...props} />);
    expect(screen.getByTestId("order-filters")).toBeInTheDocument();
  });

  it("passes loading prop through to OrderFilters", () => {
    const props = makeProps();
    props.loading = true;
    render(<OrderActionBar {...props} />);
    expect(screen.getByTestId("order-filters")).toHaveAttribute(
      "data-loading",
      "true",
    );
  });
});
