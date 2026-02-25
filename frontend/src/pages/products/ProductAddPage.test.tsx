import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ProductAddPage from "./ProductAddPage";

// ---------------------------------------------------------------------------
// Browser API stubs required by Ant Design
// ---------------------------------------------------------------------------
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

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: () => vi.fn(),
    Link: ({ children }: { children: React.ReactNode }) => <a>{children}</a>,
  };
});

const mockCreateProduct = vi.fn();
vi.mock("@/api/products", () => ({
  createProduct: (data: unknown) => mockCreateProduct(data),
}));

vi.mock("../../components/forms/ProductBasicForm", () => ({
  ProductBasicForm: ({ onFinish }: { onFinish: (val: unknown) => void }) => (
    <div data-testid="basic-form">
      Basic Form
      <button onClick={() => onFinish({ item_name: "New Product" })}>
        Next
      </button>
    </div>
  ),
}));

vi.mock("../../components/forms/ProductCategoryForm", () => ({
  ProductCategoryForm: ({
    onFinish,
    onBack,
  }: {
    onFinish: (val: unknown) => void;
    onBack: () => void;
  }) => (
    <div data-testid="category-form">
      Category Form
      <button onClick={() => onFinish({ category: "Electronics" })}>
        Next
      </button>
      <button onClick={onBack}>Back</button>
    </div>
  ),
}));

vi.mock("../../components/forms/ProductMediaForm", () => ({
  ProductMediaForm: ({
    onFinish,
    onBack,
  }: {
    onFinish: (val: unknown) => void;
    onBack: () => void;
  }) => (
    <div data-testid="media-form">
      Media Form
      <button onClick={() => onFinish({ images: [] })}>Next</button>
      <button onClick={onBack}>Back</button>
    </div>
  ),
}));

vi.mock("../../components/forms/ProductPricingForm", () => ({
  ProductPricingForm: ({
    onFinish,
    onBack,
  }: {
    onFinish: (val: unknown) => void;
    onBack: () => void;
  }) => (
    <div data-testid="pricing-form">
      Pricing Form
      <button onClick={() => onFinish({ price: 100 })}>Submit</button>
      <button onClick={onBack}>Back</button>
    </div>
  ),
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("ProductAddPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders initial step (Basic Info)", () => {
    render(<ProductAddPage />);
    expect(screen.getByText("Add New Product")).toBeInTheDocument();
    expect(screen.getByTestId("basic-form")).toBeInTheDocument();
  });

  it("navigates through steps and submits", async () => {
    mockCreateProduct.mockResolvedValue({ id: "123" });

    render(<ProductAddPage />);

    // Step 1: Basic
    fireEvent.click(screen.getByText("Next"));
    await waitFor(() =>
      expect(screen.getByTestId("category-form")).toBeInTheDocument(),
    );

    // Step 2: Category
    fireEvent.click(screen.getByText("Next"));
    await waitFor(() =>
      expect(screen.getByTestId("media-form")).toBeInTheDocument(),
    );

    // Step 3: Media
    fireEvent.click(screen.getByText("Next"));
    await waitFor(() =>
      expect(screen.getByTestId("pricing-form")).toBeInTheDocument(),
    );

    // Step 4: Pricing (Submit)
    fireEvent.click(screen.getByText("Submit"));

    await waitFor(() => {
      expect(mockCreateProduct).toHaveBeenCalled();
    });
  });

  it("navigates back", async () => {
    render(<ProductAddPage />);

    // Go to step 2
    fireEvent.click(screen.getByText("Next"));
    await waitFor(() =>
      expect(screen.getByTestId("category-form")).toBeInTheDocument(),
    );

    // Go back to step 1
    fireEvent.click(screen.getByText("Back"));
    await waitFor(() =>
      expect(screen.getByTestId("basic-form")).toBeInTheDocument(),
    );
  });
});
