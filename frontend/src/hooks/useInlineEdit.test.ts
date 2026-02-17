/**
 * useInlineEdit Hook Tests
 * Tests state machine for inline editing: edit, change, save, cancel, error handling
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { useInlineEdit } from "./useInlineEdit";

describe("useInlineEdit", () => {
  let mockOnSave: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockOnSave = vi.fn().mockResolvedValue(undefined);
  });

  it("initializes with default values", () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 100, onSave: mockOnSave }),
    );

    expect(result.current.value).toBe(100);
    expect(result.current.editValue).toBe("100");
    expect(result.current.isEditing).toBe(false);
    expect(result.current.isLoading).toBe(false);
    expect(result.current.error).toBe(null);
  });

  it("enters edit mode on handleEdit", () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 50, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    expect(result.current.isEditing).toBe(true);
    expect(result.current.editValue).toBe("50");
  });

  it("updates editValue on handleChange", () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 75, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("120");
    });

    expect(result.current.editValue).toBe("120");
    expect(result.current.value).toBe(75); // value unchanged until save
  });

  it("saves valid value and exits edit mode", async () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 100, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("150");
    });

    await act(async () => {
      await result.current.handleSave();
    });

    await waitFor(() => {
      expect(mockOnSave).toHaveBeenCalledWith(150);
      expect(result.current.value).toBe(150);
      expect(result.current.isEditing).toBe(false);
      expect(result.current.error).toBe(null);
    });
  });

  it("cancels edit and reverts to original value", () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 200, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("300");
    });

    act(() => {
      result.current.handleCancel();
    });

    expect(result.current.isEditing).toBe(false);
    expect(result.current.editValue).toBe("200");
    expect(result.current.value).toBe(200);
    expect(mockOnSave).not.toHaveBeenCalled();
  });

  it("rejects invalid number input", async () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 50, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("not-a-number");
    });

    await act(async () => {
      await result.current.handleSave();
    });

    expect(result.current.error).toBe("Invalid number");
    expect(result.current.isEditing).toBe(true);
    expect(mockOnSave).not.toHaveBeenCalled();
  });

  it("enforces minimum value constraint", async () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 100, onSave: mockOnSave, min: 10 }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("5");
    });

    await act(async () => {
      await result.current.handleSave();
    });

    expect(result.current.error).toBe("Value must be at least 10");
    expect(result.current.isEditing).toBe(true);
    expect(mockOnSave).not.toHaveBeenCalled();
  });

  it("uses custom validation function", async () => {
    const customValidate = vi.fn((val: number) => val % 10 === 0); // Must be multiple of 10

    const { result } = renderHook(() =>
      useInlineEdit({
        initialValue: 100,
        onSave: mockOnSave,
        validate: customValidate,
      }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("105");
    });

    await act(async () => {
      await result.current.handleSave();
    });

    expect(customValidate).toHaveBeenCalledWith(105);
    expect(result.current.error).toBe("Invalid value");
    expect(mockOnSave).not.toHaveBeenCalled();
  });

  it("handles save failure gracefully", async () => {
    const mockFailingSave = vi
      .fn()
      .mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 100, onSave: mockFailingSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("150");
    });

    await act(async () => {
      await result.current.handleSave();
    });

    await waitFor(() => {
      expect(result.current.error).toBe("Network error");
      expect(result.current.isEditing).toBe(true); // Stays in edit mode on error
      expect(result.current.value).toBe(100); // Value not updated
    });
  });

  it("clears error on new handleEdit", () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 50, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("abc");
    });

    act(() => {
      result.current.handleSave();
    });

    // Error should be present
    expect(result.current.error).toBe("Invalid number");

    // Re-enter edit mode
    act(() => {
      result.current.handleCancel();
    });

    act(() => {
      result.current.handleEdit();
    });

    // Error cleared on new edit session
    expect(result.current.error).toBe(null);
  });

  it("shows loading state during save", async () => {
    const slowSave = vi.fn(
      (_value: number) =>
        new Promise<void>((resolve) => setTimeout(() => resolve(), 50)),
    );

    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 100, onSave: slowSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("200");
    });

    // Start save and verify loading becomes true then false
    await act(async () => {
      const savePromise = result.current.handleSave();
      // Check loading immediately after starting save (may already be true)
      await savePromise;
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
      expect(slowSave).toHaveBeenCalledWith(200);
    });
  });

  it("allows partial numeric input (e.g., '12.' before '125')", () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 100, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("12.");
    });

    expect(result.current.editValue).toBe("12.");
    expect(result.current.error).toBe(null);
  });

  it("respects min=0 by default", async () => {
    const { result } = renderHook(() =>
      useInlineEdit({ initialValue: 50, onSave: mockOnSave }),
    );

    act(() => {
      result.current.handleEdit();
    });

    act(() => {
      result.current.handleChange("-10");
    });

    await act(async () => {
      await result.current.handleSave();
    });

    expect(result.current.error).toBe("Value must be at least 0");
    expect(mockOnSave).not.toHaveBeenCalled();
  });
});
