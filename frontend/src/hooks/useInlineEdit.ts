/**
 * useInlineEdit Hook — Reusable Click-to-Edit State Machine
 *
 * Provides state management for inline editing cells (price, stock, etc.).
 * Handles: enter edit mode, change value, save, cancel, loading/error states.
 *
 * @example
 * const { value, isEditing, isLoading, error, handleEdit, handleChange, handleSave, handleCancel } =
 *   useInlineEdit({ initialValue: 100, onSave: async (newVal) => updateProduct(id, { price: newVal }) });
 */

import { useState, useCallback } from "react";

export interface UseInlineEditOptions {
  /** Initial value for the editable field */
  initialValue: number;
  /** Async save function — receives new value, returns Promise */
  onSave: (newValue: number) => Promise<void>;
  /** Optional minimum value (default: 0) */
  min?: number;
  /** Optional validation function — return true if valid, false otherwise */
  validate?: (value: number) => boolean;
}

export interface UseInlineEditReturn {
  /** Current display value (reflects initialValue until save succeeds) */
  value: number;
  /** Draft value while editing */
  editValue: string;
  /** Whether currently in edit mode */
  isEditing: boolean;
  /** Whether save operation is in progress */
  isLoading: boolean;
  /** Error message from save operation (null if no error) */
  error: string | null;
  /** Enter edit mode */
  handleEdit: () => void;
  /** Update draft value while editing */
  handleChange: (newValue: string) => void;
  /** Save the edited value (calls onSave) */
  handleSave: () => Promise<void>;
  /** Cancel edit and revert to original value */
  handleCancel: () => void;
}

export function useInlineEdit({
  initialValue,
  onSave,
  min = 0,
  validate,
}: UseInlineEditOptions): UseInlineEditReturn {
  // Persistent value shown in non-edit mode (synced to initialValue)
  const [value, setValue] = useState<number>(initialValue);

  // Draft value while editing (string to allow partial input like "12.")
  const [editValue, setEditValue] = useState<string>(String(initialValue));

  // Edit state
  const [isEditing, setIsEditing] = useState<boolean>(false);

  // Loading state (during save)
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // Error state
  const [error, setError] = useState<string | null>(null);

  /**
   * Enter edit mode
   * Reset draft value to current value + clear error state
   */
  const handleEdit = useCallback(() => {
    setEditValue(String(value));
    setIsEditing(true);
    setError(null);
  }, [value]);

  /**
   * Update draft value while editing
   * Allows partial input (e.g., user typing "12." before "5")
   */
  const handleChange = useCallback((newValue: string) => {
    setEditValue(newValue);
    setError(null);
  }, []);

  /**
   * Save the edited value
   * Validate → Call onSave → Update persistent value → Exit edit mode
   */
  const handleSave = useCallback(async () => {
    const numericValue = parseFloat(editValue);

    // Validation: check if parseable number
    if (isNaN(numericValue)) {
      setError("Invalid number");
      return;
    }

    // Validation: check minimum
    if (numericValue < min) {
      setError(`Value must be at least ${min}`);
      return;
    }

    // Validation: custom validator
    if (validate && !validate(numericValue)) {
      setError("Invalid value");
      return;
    }

    // Perform save
    setIsLoading(true);
    setError(null);

    try {
      await onSave(numericValue);
      setValue(numericValue);
      setIsEditing(false);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to save";
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  }, [editValue, min, validate, onSave]);

  /**
   * Cancel edit and revert to original value
   * Reset draft value, clear error, exit edit mode
   */
  const handleCancel = useCallback(() => {
    setEditValue(String(value));
    setIsEditing(false);
    setError(null);
  }, [value]);

  return {
    value,
    editValue,
    isEditing,
    isLoading,
    error,
    handleEdit,
    handleChange,
    handleSave,
    handleCancel,
  };
}
