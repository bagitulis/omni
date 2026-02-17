import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { ImageGalleryPicker } from "./ImageGalleryPicker";
import apiClient from "@/api/client";
import "@testing-library/jest-dom";

// Mock matchMedia for Ant Design
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock API Client
vi.mock("@/api/client", () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

// Mock Upload component since rc-upload is hard to test in JSDOM
vi.mock("antd", async (importOriginal) => {
  const actual = await importOriginal<typeof import("antd")>();
  return {
    ...actual,
    Upload: ({
      customRequest,
      children,
    }: {
      customRequest: (options: unknown) => void;
      children: React.ReactNode;
    }) => (
      <div data-testid="upload-mock">
        <button
          type="button"
          onClick={() =>
            customRequest({
              file: new File(["(⌐□_□)"], "chucknorris.png", {
                type: "image/png",
              }),
              onSuccess: vi.fn(),
              onError: vi.fn(),
            })
          }
        >
          Mock Upload Trigger
        </button>
        {children}
      </div>
    ),
  };
});

describe("ImageGalleryPicker", () => {
  const mockOnClose = vi.fn();
  const mockOnConfirm = vi.fn();

  const mockImages = Array.from({ length: 5 }).map((_, i) => ({
    id: i + 1,
    filename: `image-${i + 1}.jpg`,
    content_hash: `hash${i + 1}`,
    local_path: `uploads/image${i + 1}`,
    mime_type: "image/jpeg",
    width: 800,
    height: 600,
    file_size: 1024,
    ref_count: 0,
    created_at: new Date().toISOString(),
  }));

  const mockResponse = {
    success: true,
    data: {
      data: mockImages,
      meta: {
        page: 1,
        pages: 1,
        total: 5,
        page_size: 20,
      },
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.get as unknown as ReturnType<typeof vi.fn>).mockResolvedValue(
      mockResponse,
    );
  });

  it("renders correctly when visible", async () => {
    render(
      <ImageGalleryPicker
        visible={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    // Title
    expect(screen.getByText("Image Gallery")).toBeInTheDocument();

    // Search input
    expect(screen.getByPlaceholderText("Search images...")).toBeInTheDocument();

    // Wait for images to load
    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeInTheDocument();
    });

    // Initial fetch call
    expect(apiClient.get).toHaveBeenCalledWith(
      "/images/gallery",
      expect.objectContaining({
        params: expect.objectContaining({ page: 1, limit: 20 }),
      }),
    );
  });

  it("handles search with debounce", async () => {
    vi.useFakeTimers();
    render(
      <ImageGalleryPicker
        visible={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    const input = screen.getByPlaceholderText("Search images...");
    fireEvent.change(input, { target: { value: "test" } });

    // Should not call immediately (debounce)
    expect(apiClient.get).toHaveBeenCalledTimes(1); // Initial load only

    // Fast-forward timers
    vi.advanceTimersByTime(500);

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalledTimes(2);
      expect(apiClient.get).toHaveBeenLastCalledWith(
        "/images/gallery",
        expect.objectContaining({
          params: expect.objectContaining({ search: "test" }),
        }),
      );
    });

    vi.useRealTimers();
  });

  it("handles selection and deselection", async () => {
    render(
      <ImageGalleryPicker
        visible={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeInTheDocument();
    });

    const image1 = screen
      .getByAltText("image-1.jpg")
      .closest("div[role='button']");

    // Select
    fireEvent.click(image1!);
    expect(image1).toHaveStyle({ border: "3px solid #ff6b2c" });
    expect(screen.getByText("1 image selected")).toBeInTheDocument();

    // Deselect
    fireEvent.click(image1!);
    expect(image1).not.toHaveStyle({ border: "3px solid #ff6b2c" });
    expect(screen.getByText("0 images selected")).toBeInTheDocument();
  });

  it("respects maxImages limit", async () => {
    render(
      <ImageGalleryPicker
        visible={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
        maxImages={2}
      />,
    );

    await waitFor(() => {
      expect(screen.getAllByRole("button").length).toBeGreaterThan(0);
    });

    const items = screen.getAllByRole("button").filter(
      (el) => el.querySelector("img"), // Filter to get only image items, not toolbar buttons
    );

    // Select 2 images
    fireEvent.click(items[0]);
    fireEvent.click(items[1]);

    expect(screen.getByText("2 images selected")).toBeInTheDocument();

    // Try selecting 3rd
    fireEvent.click(items[2]);

    // Should still be 2
    expect(screen.getByText("2 images selected")).toBeInTheDocument();
    // Warning toast would appear (not easily testable with just screen, usually mocked message)
  });

  it("handles upload successfully", async () => {
    (apiClient.post as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
      data: {},
    });

    render(
      <ImageGalleryPicker
        visible={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    const uploadBtn = screen.getByText("Mock Upload Trigger");
    fireEvent.click(uploadBtn);

    await waitFor(() => {
      expect(apiClient.post).toHaveBeenCalledWith(
        "/images/upload",
        expect.any(FormData),
        expect.any(Object),
      );
    });

    // Should refetch gallery after upload
    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalledTimes(2); // Initial + After upload
    });
  });

  it("confirms selection", async () => {
    render(
      <ImageGalleryPicker
        visible={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeInTheDocument();
    });

    const items = screen
      .getAllByRole("button")
      .filter((el) => el.querySelector("img"));
    fireEvent.click(items[0]);

    const confirmBtn = screen.getByText("Confirm");
    fireEvent.click(confirmBtn);

    expect(mockOnConfirm).toHaveBeenCalledWith([mockImages[0]]);
    expect(mockOnClose).toHaveBeenCalled();
  });

  it("keyboard navigation works", async () => {
    render(
      <ImageGalleryPicker
        visible={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeInTheDocument();
    });

    const items = screen
      .getAllByRole("button")
      .filter((el) => el.querySelector("img"));
    items[0].focus();

    // Press Enter to select
    fireEvent.keyDown(items[0], { key: "Enter", code: "Enter" });

    expect(items[0]).toHaveStyle({ border: "3px solid #ff6b2c" });
  });
});
