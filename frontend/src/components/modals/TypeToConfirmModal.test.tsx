import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { TypeToConfirmModal } from "./TypeToConfirmModal";

function renderModal(props: Partial<Parameters<typeof TypeToConfirmModal>[0]> = {}) {
  const defaultProps = {
    open: true,
    title: "Delete Item",
    description: "This action cannot be undone.",
    confirmText: "DELETE",
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
    ...props,
  };
  return { ...render(<TypeToConfirmModal {...defaultProps} />), props: defaultProps };
}

describe("TypeToConfirmModal", () => {
  it("renders with disabled confirm button initially", () => {
    renderModal();
    const okButton = screen.getByRole("button", { name: /confirm/i });
    expect(okButton).toBeDisabled();
  });

  it("keeps button disabled when input does not match", () => {
    renderModal({ confirmText: "DELETE" });
    const input = screen.getByPlaceholderText('Type "DELETE" to confirm');
    fireEvent.change(input, { target: { value: "delete" } });
    const okButton = screen.getByRole("button", { name: /confirm/i });
    expect(okButton).toBeDisabled();
  });

  it("enables button when input exactly matches confirmText", () => {
    renderModal({ confirmText: "DELETE" });
    const input = screen.getByPlaceholderText('Type "DELETE" to confirm');
    fireEvent.change(input, { target: { value: "DELETE" } });
    const okButton = screen.getByRole("button", { name: /confirm/i });
    expect(okButton).not.toBeDisabled();
  });

  it("calls onConfirm when confirm button is clicked", () => {
    const onConfirm = vi.fn();
    renderModal({ confirmText: "DELETE", onConfirm });
    const input = screen.getByPlaceholderText('Type "DELETE" to confirm');
    fireEvent.change(input, { target: { value: "DELETE" } });
    const okButton = screen.getByRole("button", { name: /confirm/i });
    fireEvent.click(okButton);
    expect(onConfirm).toHaveBeenCalledOnce();
  });

  it("calls onCancel when cancel button is clicked", () => {
    const onCancel = vi.fn();
    renderModal({ onCancel });
    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    fireEvent.click(cancelButton);
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it("resets input when modal closes and reopens", () => {
    const { rerender, props } = renderModal({ confirmText: "DELETE" });
    const input = screen.getByPlaceholderText('Type "DELETE" to confirm');
    fireEvent.change(input, { target: { value: "DELETE" } });

    // Close modal
    rerender(<TypeToConfirmModal {...props} open={false} />);

    // Reopen modal
    rerender(<TypeToConfirmModal {...props} open={true} />);
    const newInput = screen.getByPlaceholderText('Type "DELETE" to confirm');
    expect(newInput).toHaveValue("");
  });

  it("shows danger button style when danger prop is true", () => {
    renderModal({ danger: true, confirmText: "REMOVE" });
    const input = screen.getByPlaceholderText('Type "REMOVE" to confirm');
    fireEvent.change(input, { target: { value: "REMOVE" } });
    const okButton = screen.getByRole("button", { name: /confirm/i });
    expect(okButton).toHaveClass("ant-btn-dangerous");
  });
});
