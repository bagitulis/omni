import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ProductEditPage from "./ProductEditPage";

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
    useParams: () => ({ id: "123" }),
    Link: ({ children }: { children: React.ReactNode }) => <a>{children}</a>,
  };
});

const mockGetProductById = vi.fn();
const mockUpdateProduct = vi.fn();

vi.mock("../../api/products", () => ({
  getProductById: (id: string) => mockGetProductById(id),
  updateProduct: (id: string, data: Record<string, unknown>) =>
    mockUpdateProduct(id, data),
  refreshProductImages: vi.fn(),
  syncProduct: vi.fn(),
}));

vi.mock("./utils/productEditMapper", () => ({
  mapMasterProductToProductData: (data: Record<string, unknown> & { skus?: unknown[], images?: unknown[] }) => ({
    ...data,
    skus: data.skus || [],
    images: data.images || [],
    platforms: [],
  }),
}));

vi.mock("../../components/forms/ProductBasicForm", () => ({
  ProductBasicForm: ({ onFinish }: { onFinish: (data: unknown) => void }) => (
    <button
      onClick={() =>
        onFinish({ item_name: "Updated Name", description: "Desc" })
      }
    >
      Save Basic Info
    </button>
  ),
}));

vi.mock("./components/ProductVariantsTab", () => ({
  ProductVariantsTab: () => <div data-testid="variants-tab" />,
}));
vi.mock("./components/ProductImagesTab", () => ({
  ProductImagesTab: () => <div data-testid="images-tab" />,
}));
vi.mock("./components/ProductSyncTab", () => ({
  ProductSyncTab: () => <div data-testid="sync-tab" />,
}));
vi.mock("../../components/shared/SkuMappingPanel", () => ({
  SkuMappingPanel: () => <div data-testid="mapping-panel" />,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("ProductEditPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state initially", () => {
    mockGetProductById.mockReturnValue(new Promise(() => {})); // Never resolves
    render(<ProductEditPage />);
    expect(screen.getByText("Loading product...")).toBeInTheDocument();
  });

  it("renders product data after loading", async () => {
    mockGetProductById.mockResolvedValue({
      id: "123",
      title: "Test Product",
      description: "Test Description",
      images: [],
      skus: [],
    });

    render(<ProductEditPage />);

    await waitFor(() => {
      expect(
        screen.getByText("Edit Product: Test Product"),
      ).toBeInTheDocument();
      expect(screen.getByText("Product ID: 123")).toBeInTheDocument();
    });
  });

  it("handles basic info save", async () => {
    mockGetProductById.mockResolvedValue({
      id: "123",
      title: "Test Product",
      description: "Test Description",
      images: [],
      skus: [],
    });
    mockUpdateProduct.mockResolvedValue({});

    render(<ProductEditPage />);

    await waitFor(() => {
      expect(screen.getByText("Save Basic Info")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText("Save Basic Info"));

    await waitFor(() => {
      expect(mockUpdateProduct).toHaveBeenCalledWith("123", {
        title: "Updated Name",
        description: "Desc",
      });
    });
  });

  it("renders error state", async () => {
    mockGetProductById.mockRejectedValue(new Error("Failed to load"));

    render(<ProductEditPage />);

    await waitFor(() => {
      expect(screen.getByText("Failed to load product")).toBeInTheDocument();
      expect(screen.getByText("Failed to load")).toBeInTheDocument();
    });
  });
});
