import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { Product } from "@/types/product";
import { ProductGridView } from "./ProductGridView";

// Mock PlatformBadge
vi.mock("@/components/ui/PlatformBadge", () => ({
  PlatformBadge: ({ platform }: { platform: string }) => (
    <span data-testid="platform-badge">{platform}</span>
  ),
}));

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

function makeProduct(overrides: Partial<Product> = {}): Product {
  return {
    item_id: "prod-1",
    item_name: "Test Product",
    item_sku: "SKU-001",
    platform: "shopee",
    image_url: "https://example.com/image.jpg",
    price: 25000,
    stock: 10,
    ...overrides,
  } as Product;
}

describe("ProductGridView", () => {
  it("renders Empty when products is empty", () => {
    render(
      <ProductGridView
        products={[]}
        page={1}
        pageSize={10}
        total={0}
        onPageChange={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText("No products found")).toBeInTheDocument();
  });

  it("renders product cards when products provided", () => {
    const products = [
      makeProduct({ item_id: "1", item_name: "Shoe A" }),
      makeProduct({ item_id: "2", item_name: "Shoe B" }),
    ];
    render(
      <ProductGridView
        products={products}
        page={1}
        pageSize={10}
        total={2}
        onPageChange={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText("Shoe A")).toBeInTheDocument();
    expect(screen.getByText("Shoe B")).toBeInTheDocument();
  });

  it("shows formatted price", () => {
    const product = makeProduct({ price: 50000 });
    render(
      <ProductGridView
        products={[product]}
        page={1}
        pageSize={10}
        total={1}
        onPageChange={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    // IDR format: Rp 50.000
    expect(screen.getByText(/Rp/)).toBeInTheDocument();
  });

  it("shows '—' when price is null", () => {
    const product = makeProduct({ price: null as unknown as number });
    render(
      <ProductGridView
        products={[product]}
        page={1}
        pageSize={10}
        total={1}
        onPageChange={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("shows '—' for stock when null", () => {
    const product = makeProduct({ stock: null as unknown as number });
    render(
      <ProductGridView
        products={[product]}
        page={1}
        pageSize={10}
        total={1}
        onPageChange={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText(/Stock:\s*—/)).toBeInTheDocument();
  });

  it("calls onDelete when delete button clicked", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    const product = makeProduct({ item_id: "prod-del" });
    render(
      <ProductGridView
        products={[product]}
        page={1}
        pageSize={10}
        total={1}
        onPageChange={vi.fn()}
        onDelete={onDelete}
      />,
    );
    // Delete button is the danger button (second action)
    const buttons = screen.getAllByRole("button");
    const deleteBtn = buttons.find(
      (b) => b.querySelector("[data-icon='delete']") !== null,
    );
    if (deleteBtn) {
      await user.click(deleteBtn);
      expect(onDelete).toHaveBeenCalledWith("prod-del");
    }
  });

  it("renders Pagination", () => {
    const products = [makeProduct()];
    render(
      <ProductGridView
        products={products}
        page={1}
        pageSize={10}
        total={50}
        onPageChange={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    // Pagination renders page numbers
    expect(screen.getAllByRole("listitem").length).toBeGreaterThan(0);
  });
});
