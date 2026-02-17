import React, { useRef, useEffect } from "react";
import { Input, Typography, theme } from "antd";
import type { InputRef } from "antd";
import {
  LoadingOutlined,
  EditOutlined,
  CheckOutlined,
  CloseOutlined,
} from "@ant-design/icons";
import { useInlineEdit } from "@/hooks/useInlineEdit";
import type { InlineEditCellProps } from "@/types/shared";

const { Text } = Typography;

export const InlineEditCell: React.FC<InlineEditCellProps> = ({
  value: initialValue,
  mode,
  onSave,
  disabled = false,
  min = 0,
  prefix,
}) => {
  const {
    value,
    editValue,
    isEditing,
    isLoading,
    error,
    handleEdit,
    handleChange,
    handleSave,
    handleCancel,
  } = useInlineEdit({
    initialValue,
    onSave,
    min,
  });

  const inputRef = useRef<InputRef>(null);
  const { token } = theme.useToken();

  // Focus input when entering edit mode
  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isEditing]);

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleSave();
    } else if (e.key === "Escape") {
      handleCancel();
    }
  };

  const handleBlur = () => {
    // Optional: save on blur, or cancel.
    // For inline edit cells, saving on blur is common but can be annoying if accidental.
    // Let's stick to explicit Enter/Escape or button click for now as per "Handle Enter/Escape" requirement.
    // Actually, widespread pattern is click-outside to save or cancel.
    // The requirement says "Handle Enter/Escape interactions and blur behavior safely".
    // "Safely" usually means don't leave it in a broken state.
    // Let's implement cancel on blur for safety if not saving, or maybe save on blur?
    // Given it's a "Click-to-Edit" cell, usually blur = save or cancel.
    // Let's try to save on blur, but if it fails validation it might be tricky.
    // Safer to just cancel or keep editing?
    // Let's stick to Enter/Escape for explicit action to avoid "stuck" focus loops with validation errors.
    // But we should probably handle blur to at least exit if no changes?
    // For now, let's keep it simple: Enter to save, Escape to cancel.
    // However, if the user clicks away, it shouldn't remain in edit mode forever.
    // Let's cancel on blur if not loading.
    if (!isLoading) {
      // Check if value changed?
      // If we cancel on blur, user loses work.
      // If we save on blur, user might save accidental work.
      // Let's go with Save on blur as it's the most common "Excel-like" behavior.
      handleSave();
    }
  };

  // Render Display Mode
  if (!isEditing && !isLoading) {
    return (
      <div
        className={`inline-edit-cell-display mode-${mode}`}
        onClick={!disabled ? handleEdit : undefined}
        onKeyDown={(e) => {
          if (!disabled && (e.key === "Enter" || e.key === " ")) {
            e.preventDefault();
            handleEdit();
          }
        }}
        style={{
          cursor: disabled ? "not-allowed" : "pointer",
          padding: "4px 8px",
          minHeight: "32px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          border: `1px solid transparent`,
          borderRadius: token.borderRadius,
          transition: "all 0.2s",
        }}
        onMouseEnter={(e) => {
          if (!disabled) {
            e.currentTarget.style.borderColor = token.colorBorder;
            e.currentTarget.style.backgroundColor = token.colorBgTextHover;
          }
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.borderColor = "transparent";
          e.currentTarget.style.backgroundColor = "transparent";
        }}
        role="button"
        aria-label={`Edit ${mode}`}
        tabIndex={disabled ? -1 : 0}
        onFocus={!disabled ? handleEdit : undefined}
      >
        <Text>
          {prefix && (
            <span style={{ color: token.colorTextSecondary, marginRight: 4 }}>
              {prefix}
            </span>
          )}
          {value.toLocaleString()}
        </Text>
        {!disabled && (
          <EditOutlined
            style={{
              fontSize: 12,
              color: token.colorTextQuaternary,
              opacity: 0.5,
            }}
          />
        )}
      </div>
    );
  }

  // Render Edit Mode
  return (
    <div className="inline-edit-cell-edit" style={{ position: "relative" }}>
      <Input
        ref={inputRef}
        value={editValue}
        onChange={(e) => handleChange(e.target.value)}
        onKeyDown={onKeyDown}
        onBlur={handleBlur}
        disabled={isLoading}
        status={error ? "error" : ""}
        prefix={prefix}
        suffix={
          isLoading ? (
            <LoadingOutlined />
          ) : (
            <div style={{ display: "flex", gap: 4 }}>
              <CheckOutlined
                onClick={(e) => {
                  e.stopPropagation();
                  handleSave();
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.stopPropagation();
                    handleSave();
                  }
                }}
                tabIndex={0}
                role="button"
                style={{ cursor: "pointer", color: token.colorSuccess }}
              />
              <CloseOutlined
                onClick={(e) => {
                  e.stopPropagation();
                  handleCancel();
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.stopPropagation();
                    handleCancel();
                  }
                }}
                tabIndex={0}
                role="button"
                style={{ cursor: "pointer", color: token.colorTextSecondary }}
              />
            </div>
          )
        }
        style={{ width: "100%" }}
      />
      {error && (
        <div
          style={{
            position: "absolute",
            top: "100%",
            left: 0,
            zIndex: 10,
            fontSize: 12,
            color: token.colorError,
            backgroundColor: token.colorBgElevated,
            padding: "4px 8px",
            boxShadow: token.boxShadowSecondary,
            borderRadius: token.borderRadius,
            marginTop: 4,
          }}
        >
          {error}
        </div>
      )}
    </div>
  );
};
