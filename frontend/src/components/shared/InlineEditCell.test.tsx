import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom";
import { InlineEditCell } from "./InlineEditCell";

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

describe("InlineEditCell", () => {
  let mockOnSave: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockOnSave = vi.fn().mockResolvedValue(undefined);
  });

  it("renders initial value correctly", () => {
    render(
      <InlineEditCell
        value={100}
        mode="price"
        onSave={mockOnSave}
        prefix="Rp"
      />,
    );

    expect(screen.getByText("100")).toBeInTheDocument();
    expect(screen.getByText("Rp")).toBeInTheDocument();
  });

  it("enters edit mode on click", async () => {
    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    const displayCell = screen.getByRole("button");
    fireEvent.click(displayCell);

    const input = screen.getByRole("textbox");
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue("100");
  });

  it("does not enter edit mode from focus only", () => {
    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    const displayCell = screen.getByRole("button", { name: "Edit value" });
    fireEvent.focus(displayCell);

    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("saves on Enter key", async () => {
    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    // Enter edit mode
    fireEvent.click(screen.getByRole("button"));

    const input = screen.getByRole("textbox");
    // Change value
    fireEvent.change(input, { target: { value: "150" } });

    // Press Enter
    fireEvent.keyDown(input, { key: "Enter", code: "Enter" });

    await waitFor(() => {
      expect(mockOnSave).toHaveBeenCalledWith(150);
    });

    await waitFor(() => {
      expect(screen.getByTestId("inline-edit-cell-display")).toHaveAttribute(
        "data-flash-state",
        "success",
      );
    });
  });

  it("cancels on Escape key", async () => {
    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    // Enter edit mode
    fireEvent.click(screen.getByRole("button"));

    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "200" } });

    // Press Escape
    fireEvent.keyDown(input, { key: "Escape", code: "Escape" });

    // Should revert to display mode
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.getByText("100")).toBeInTheDocument();
    expect(mockOnSave).not.toHaveBeenCalled();
  });

  it("shows validation error and error flash on invalid input", async () => {
    render(
      <InlineEditCell value={100} mode="stock" onSave={mockOnSave} min={0} />,
    );

    fireEvent.click(screen.getByRole("button"));
    const input = screen.getByRole("textbox");

    // Invalid value
    fireEvent.change(input, { target: { value: "-10" } });
    fireEvent.keyDown(input, { key: "Enter", code: "Enter" });

    expect(
      await screen.findByText("Value must be at least 0"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("inline-edit-cell-edit")).toHaveAttribute(
      "data-flash-state",
      "error",
    );
    expect(mockOnSave).not.toHaveBeenCalled();
  });

  it("saves on blur", async () => {
    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    fireEvent.click(screen.getByRole("button"));
    const input = screen.getByRole("textbox");

    fireEvent.change(input, { target: { value: "120" } });
    fireEvent.blur(input);

    await waitFor(() => {
      expect(mockOnSave).toHaveBeenCalledWith(120);
    });
  });

  it("cancels from cancel action without triggering blur save", async () => {
    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    fireEvent.click(screen.getByRole("button", { name: "Edit value" }));
    const input = screen.getByRole("textbox");

    fireEvent.change(input, { target: { value: "130" } });
    fireEvent.click(screen.getByLabelText("Cancel edit"));

    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.getByText("100")).toBeInTheDocument();
    expect(mockOnSave).not.toHaveBeenCalled();
  });

  it("shows error flash and stays in edit mode when save fails", async () => {
    mockOnSave = vi.fn().mockRejectedValue(new Error("Network error"));

    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    fireEvent.click(screen.getByRole("button", { name: "Edit value" }));
    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "110" } });
    fireEvent.keyDown(input, { key: "Enter", code: "Enter" });

    expect(await screen.findByText("Network error")).toBeInTheDocument();
    expect(screen.getByTestId("inline-edit-cell-edit")).toHaveAttribute(
      "data-flash-state",
      "error",
    );
    expect(screen.getByRole("textbox")).toBeInTheDocument();
  });

  it("handles disabled state", () => {
    render(
      <InlineEditCell value={100} mode="stock" onSave={mockOnSave} disabled />,
    );

    const displayCell = screen.getByRole("button");
    fireEvent.click(displayCell);

    // Should not enter edit mode
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("displays loading state", async () => {
    // Mock a slow save
    mockOnSave.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100)),
    );

    render(<InlineEditCell value={100} mode="stock" onSave={mockOnSave} />);

    fireEvent.click(screen.getByRole("button"));
    const input = screen.getByRole("textbox");

    fireEvent.change(input, { target: { value: "120" } });
    fireEvent.keyDown(input, { key: "Enter", code: "Enter" });

    // Spin/Loading icon should appear (Ant Design's Spin or LoadingOutlined)
    // We can check for the disabled input or loading class
    expect(input).toBeDisabled();
  });
});
