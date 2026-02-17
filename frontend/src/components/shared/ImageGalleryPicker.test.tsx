// Mock API Client FIRST before any imports
vi.mock("@/api/client", () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

import {
  render,
  screen,
  fireEvent,
  waitFor,
  act,
} from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { ImageGalleryPicker } from "./ImageGalleryPicker";
import apiClient from "@/api/client";

// Suppress known AntD compatibility warning BEFORE component renders
const originalWarn = console.warn.bind(console);
const originalError = console.error.bind(console);

// Override both warn and error to catch AntD compatibility messages
console.warn = (message?: unknown, ...optionalParams: unknown[]) => {
  const msg = String(message);
  if (msg.includes("[antd: compatible] antd v5 support React is 16 ~ 18")) {
    return; // Silently drop this specific warning
  }
  originalWarn(message, ...optionalParams);
};

console.error = (message?: unknown, ...optionalParams: unknown[]) => {
  const msg = String(message);
  if (msg.includes("[antd: compatible] antd v5 support React is 16 ~ 18")) {
    return; // Silently drop this specific warning
  }
  originalError(message, ...optionalParams);
};

// Mock matchMedia for Ant Design
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
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

// Stub getComputedStyle to prevent JSDOM not-implemented warnings (Ant Design uses it)
Object.defineProperty(window, "getComputedStyle", {
  writable: true,
  value: vi.fn().mockImplementation(() => ({
    getPropertyValue: vi.fn().mockReturnValue(""),
  })),
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

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  it("renders correctly when open and fetches images", async () => {
    render(
      <ImageGalleryPicker
        open={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    // Title
    expect(screen.getByText("Image Gallery")).toBeTruthy();

    // Search input
    expect(screen.getByPlaceholderText("Search images...")).toBeTruthy();

    // Wait for images to load
    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeTruthy();
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
    // Use real timers with explicit wait to avoid fake-timer + waitFor hang
    render(
      <ImageGalleryPicker
        open={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    const input = screen.getByPlaceholderText("Search images...");

    // Initial load happens first
    expect(apiClient.get).toHaveBeenCalledTimes(1);

    fireEvent.change(input, { target: { value: "test" } });

    // Should not call immediately (debounce 500ms)
    expect(apiClient.get).toHaveBeenCalledTimes(1);

    // Wait for debounce to trigger (500ms + buffer) - wrapped in act
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 550));
    });

    // Now debounced search should have fired
    expect(apiClient.get).toHaveBeenCalledTimes(2);
    expect(apiClient.get).toHaveBeenLastCalledWith(
      "/images/gallery",
      expect.objectContaining({
        params: expect.objectContaining({ search: "test" }),
      }),
    );
  });

  it("handles selection and deselection", async () => {
    render(
      <ImageGalleryPicker
        open={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeTruthy();
    });

    const image1 = screen
      .getByAltText("image-1.jpg")
      .closest("div[role='button']");

    // Select - assert via counter text (robust behavioral check)
    fireEvent.click(image1!);
    expect(screen.getByText("1 / 8 selected")).toBeTruthy();

    // Deselect - assert via counter text
    fireEvent.click(image1!);
    expect(screen.getByText("0 / 8 selected")).toBeTruthy();
  });

  it("respects maxSelect limit", async () => {
    render(
      <ImageGalleryPicker
        open={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
        maxSelect={2}
      />,
    );

    await waitFor(() => {
      expect(screen.getAllByRole("button").length).toBeGreaterThan(0);
    });

    const items = screen
      .getAllByRole("button")
      .filter((el) => el.querySelector("img"));

    // Select 2 images
    fireEvent.click(items[0]);
    fireEvent.click(items[1]);

    expect(screen.getByText("2 / 2 selected")).toBeTruthy();

    // Try selecting 3rd - should be blocked
    fireEvent.click(items[2]);

    // Should still be 2 (3rd selection blocked)
    expect(screen.getByText("2 / 2 selected")).toBeTruthy();
  });

  it("calls onConfirm and onClose when confirm button clicked", async () => {
    render(
      <ImageGalleryPicker
        open={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeTruthy();
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

  it("keyboard navigation works with Enter key", async () => {
    render(
      <ImageGalleryPicker
        open={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />,
    );

    await waitFor(() => {
      expect(screen.getByAltText("image-1.jpg")).toBeTruthy();
    });

    const items = screen
      .getAllByRole("button")
      .filter((el) => el.querySelector("img"));
    items[0].focus();

    // Press Enter to select - assert via counter text change
    fireEvent.keyDown(items[0], { key: "Enter", code: "Enter" });

    expect(screen.getByText("1 / 8 selected")).toBeTruthy();
  });
});
