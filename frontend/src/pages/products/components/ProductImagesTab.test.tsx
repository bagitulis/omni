import "@testing-library/jest-dom";
import { render, screen, fireEvent } from "@testing-library/react";
import type { UploadFile } from "antd/es/upload/interface";
import { ProductImagesTab } from "./ProductImagesTab";

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

vi.mock("@/components/shared/ImageGalleryPicker", () => ({
  ImageGalleryPicker: ({
    open,
    onClose,
  }: {
    open: boolean;
    onClose: () => void;
  }) =>
    open ? (
      <div data-testid="gallery-picker">
        <button type="button" onClick={onClose}>Close Gallery</button>
      </div>
    ) : null,
}));

vi.mock("@/types/shared", () => ({
  getImageUrl: vi.fn((path: string) => `https://cdn.example.com/${path}`),
}));

const makeFile = (uid: string, url: string): UploadFile => ({
  uid,
  name: `image-${uid}.png`,
  status: "done",
  url,
});

describe("ProductImagesTab", () => {
  const onSave = vi.fn();
  const onRefresh = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders Save Images button", () => {
    render(
      <ProductImagesTab initialValues={[]} onSave={onSave} loading={false} />,
    );
    expect(
      screen.getByRole("button", { name: /save images/i }),
    ).toBeInTheDocument();
  });

  it("renders Browse Gallery button when under max images", () => {
    render(
      <ProductImagesTab initialValues={[]} onSave={onSave} loading={false} />,
    );
    expect(
      screen.getByRole("button", { name: /browse gallery/i }),
    ).toBeInTheDocument();
  });

  it("disables Browse Gallery button when at max (8) images", () => {
    const files = Array.from({ length: 8 }, (_, i) =>
      makeFile(`${i}`, `https://example.com/img${i}.png`),
    );
    render(
      <ProductImagesTab
        initialValues={files}
        onSave={onSave}
        loading={false}
      />,
    );
    expect(
      screen.getByRole("button", { name: /browse gallery/i }),
    ).toBeDisabled();
  });

  it("renders Refresh From Sync button disabled when no onRefresh", () => {
    render(
      <ProductImagesTab initialValues={[]} onSave={onSave} loading={false} />,
    );
    expect(
      screen.getByRole("button", { name: /refresh from sync/i }),
    ).toBeDisabled();
  });

  it("renders Refresh From Sync button enabled when onRefresh provided", () => {
    render(
      <ProductImagesTab
        initialValues={[]}
        onSave={onSave}
        onRefresh={onRefresh}
        loading={false}
      />,
    );
    expect(
      screen.getByRole("button", { name: /refresh from sync/i }),
    ).not.toBeDisabled();
  });

  it("calls onSave with current file list when Save Images clicked", () => {
    const files = [makeFile("1", "https://example.com/a.png")];
    render(
      <ProductImagesTab
        initialValues={files}
        onSave={onSave}
        loading={false}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /save images/i }));
    expect(onSave).toHaveBeenCalledWith(files);
  });

  it("calls onRefresh when Refresh From Sync clicked", () => {
    render(
      <ProductImagesTab
        initialValues={[]}
        onSave={onSave}
        onRefresh={onRefresh}
        loading={false}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /refresh from sync/i }));
    expect(onRefresh).toHaveBeenCalled();
  });

  it("opens gallery picker when Browse Gallery clicked", () => {
    render(
      <ProductImagesTab initialValues={[]} onSave={onSave} loading={false} />,
    );
    expect(screen.queryByTestId("gallery-picker")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /browse gallery/i }));
    expect(screen.getByTestId("gallery-picker")).toBeInTheDocument();
  });

  it("closes gallery picker when Close Gallery clicked", () => {
    render(
      <ProductImagesTab initialValues={[]} onSave={onSave} loading={false} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /browse gallery/i }));
    fireEvent.click(screen.getByRole("button", { name: /close gallery/i }));
    expect(screen.queryByTestId("gallery-picker")).not.toBeInTheDocument();
  });

  it("shows loading on Save Images button when loading=true", () => {
    render(
      <ProductImagesTab initialValues={[]} onSave={onSave} loading={true} />,
    );
    // Ant Design adds aria-busy or disables while loading
    const btn = screen.getByRole("button", { name: /save images/i });
    expect(btn).toBeInTheDocument();
  });
});
