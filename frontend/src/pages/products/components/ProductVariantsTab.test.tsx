import "@testing-library/jest-dom";
import { render, screen, fireEvent } from "@testing-library/react";
import type { ProductSku } from "../types";
import { ProductVariantsTab } from "./ProductVariantsTab";

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

const mockSkus: ProductSku[] = [
  {
    key: "sku-1",
    seller_sku: "SKU001",
    variant_name: "Red",
    stock: 10,
    price: 50000,
  },
  {
    key: "sku-2",
    seller_sku: "SKU002",
    variant_name: "Blue",
    stock: 5,
    price: 60000,
  },
];

describe("ProductVariantsTab", () => {
  const onSave = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders table column headers", () => {
    render(
      <ProductVariantsTab
        initialValues={mockSkus}
        onSave={onSave}
        loading={false}
      />,
    );
    expect(screen.getByText("Variant Name")).toBeInTheDocument();
    expect(screen.getByText("Seller SKU")).toBeInTheDocument();
    expect(screen.getByText("Stock")).toBeInTheDocument();
    expect(screen.getByText("Price")).toBeInTheDocument();
    expect(screen.getByText("Action")).toBeInTheDocument();
  });

  it("renders variant data in input fields", () => {
    render(
      <ProductVariantsTab
        initialValues={mockSkus}
        onSave={onSave}
        loading={false}
      />,
    );
    const inputs = screen.getAllByRole("textbox");
    const values = inputs.map((i) => (i as HTMLInputElement).value);
    expect(values).toContain("Red");
    expect(values).toContain("SKU001");
  });

  it("renders Add Variant button", () => {
    render(
      <ProductVariantsTab
        initialValues={mockSkus}
        onSave={onSave}
        loading={false}
      />,
    );
    expect(
      screen.getByRole("button", { name: /add variant/i }),
    ).toBeInTheDocument();
  });

  it("renders Save Variants button", () => {
    render(
      <ProductVariantsTab
        initialValues={mockSkus}
        onSave={onSave}
        loading={false}
      />,
    );
    expect(
      screen.getByRole("button", { name: /save variants/i }),
    ).toBeInTheDocument();
  });

  it("adds a new empty row when Add Variant clicked", () => {
    render(
      <ProductVariantsTab
        initialValues={mockSkus}
        onSave={onSave}
        loading={false}
      />,
    );
    const textInputsBefore = screen.getAllByRole("textbox").length;
    fireEvent.click(screen.getByRole("button", { name: /add variant/i }));
    const textInputsAfter = screen.getAllByRole("textbox").length;
    expect(textInputsAfter).toBeGreaterThan(textInputsBefore);
  });

  it("calls onSave when Save Variants clicked", () => {
    render(
      <ProductVariantsTab
        initialValues={mockSkus}
        onSave={onSave}
        loading={false}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /save variants/i }));
    expect(onSave).toHaveBeenCalledWith(mockSkus);
  });

  it("renders with empty initial values", () => {
    render(
      <ProductVariantsTab initialValues={[]} onSave={onSave} loading={false} />,
    );
    expect(
      screen.getByRole("button", { name: /add variant/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /save variants/i }),
    ).toBeInTheDocument();
  });
});
