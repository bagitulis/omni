import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ProductImportPage from "./ProductImportPage";

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
vi.mock("antd", async (importOriginal) => {
  const actual = await importOriginal<Record<string, unknown>>();
  return {
    ...(actual as Record<string, unknown>),
  };
});
vi.mock("@/components/AntStaticApi", () => ({
  message: { success: vi.fn(), error: vi.fn(), warning: vi.fn(), info: vi.fn() },
  modal: { confirm: vi.fn(({ onOk }: { onOk?: () => void }) => onOk?.()) },
  notification: {},
}));
vi.mock("react-router-dom", () => ({
  useNavigate: () => vi.fn(),
}));

const mockMutatePreview = vi.fn();
const mockMutateImport = vi.fn();
const mockMutateAutoMap = vi.fn();

vi.mock("@/hooks/useProductImport", () => ({
  useImportPreview: () => ({
    mutateAsync: mockMutatePreview,
    isPending: false,
  }),
  useImportProducts: () => ({
    mutateAsync: mockMutateImport,
    isPending: false,
  }),
  useAutoMapSkus: () => ({
    mutateAsync: mockMutateAutoMap,
    isPending: false,
  }),
}));

vi.mock("@/api/products", () => ({
  downloadImportTemplate: vi.fn(),
}));

vi.mock("@/components/forms/ImportUploader", () => ({
  ImportUploader: ({ onFileSelect }: { onFileSelect: (file: File) => void }) => (
    <button onClick={() => onFileSelect(new File(["content"], "test.csv"))}>
      Upload File
    </button>
  ),
}));

vi.mock("@/components/tables/ImportPreviewTable", () => ({
  ImportPreviewTable: () => <div data-testid="import-preview-table" />,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("ProductImportPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders upload step initially", () => {
    render(<ProductImportPage />);
    expect(screen.getByText("Import Products")).toBeInTheDocument();
    expect(screen.getByText("Download Template")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Upload File" })).toBeInTheDocument();
  });

  it("advances to preview step on successful upload", async () => {
    mockMutatePreview.mockResolvedValueOnce({
      total_rows: 10,
      valid_rows: 10,
      invalid_rows: 0,
      rows: [],
    });

    render(<ProductImportPage />);

    const uploadBtn = screen.getByRole("button", { name: "Upload File" });
    fireEvent.click(uploadBtn);

    await waitFor(() => {
      expect(mockMutatePreview).toHaveBeenCalled();
      expect(screen.getByText("Total Rows")).toBeInTheDocument();
      expect(screen.getByTestId("import-preview-table")).toBeInTheDocument();
    });
  });

  it("handles confirm import", async () => {
    mockMutatePreview.mockResolvedValueOnce({
      total_rows: 10,
      valid_rows: 10,
      invalid_rows: 0,
      rows: [],
    });

    render(<ProductImportPage />);

    // Upload first
    fireEvent.click(screen.getByRole("button", { name: "Upload File" }));
    await waitFor(() =>
      expect(screen.getByText("Confirm & Import")).toBeInTheDocument(),
    );

    // Click confirm - this triggers Modal.confirm which auto-executes onOk via mock
    fireEvent.click(screen.getByText("Confirm & Import"));

    await waitFor(() => {
      expect(mockMutateImport).toHaveBeenCalled();
      expect(screen.getByText("Import Complete!")).toBeInTheDocument();
    });
  });
});
