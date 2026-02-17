import React, {
  useRef,
  useEffect,
  useMemo,
  useCallback,
  useState,
} from "react";
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
  const [flashState, setFlashState] = useState<"idle" | "success" | "error">(
    "idle",
  );
  const saveInvokedRef = useRef(false);
  const saveSucceededRef = useRef(false);
  const skipBlurSaveRef = useRef(false);

  const wrappedOnSave = useCallback(
    async (newValue: number) => {
      saveInvokedRef.current = true;
      saveSucceededRef.current = true;

      try {
        await onSave(newValue);
      } catch (saveError) {
        saveSucceededRef.current = false;
        throw saveError;
      }
    },
    [onSave],
  );

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
    onSave: wrappedOnSave,
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

  useEffect(() => {
    if (flashState === "idle") {
      return;
    }

    const timer = setTimeout(() => {
      setFlashState("idle");
    }, 1000);

    return () => {
      clearTimeout(timer);
    };
  }, [flashState]);

  const triggerSave = useCallback(async () => {
    if (isLoading) {
      return;
    }

    saveInvokedRef.current = false;
    saveSucceededRef.current = false;
    await handleSave();

    if (saveInvokedRef.current && saveSucceededRef.current) {
      setFlashState("success");
      return;
    }

    setFlashState("error");
  }, [handleSave, isLoading]);

  const handleCancelAction = useCallback(() => {
    skipBlurSaveRef.current = true;
    handleCancel();
  }, [handleCancel]);

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      void triggerSave();
    } else if (e.key === "Escape") {
      e.preventDefault();
      handleCancelAction();
    }
  };

  const handleBlur = () => {
    if (skipBlurSaveRef.current) {
      skipBlurSaveRef.current = false;
      return;
    }

    void triggerSave();
  };

  const flashBackground = useMemo(() => {
    if (flashState === "success") {
      return token.colorSuccessBg;
    }

    if (flashState === "error") {
      return token.colorErrorBg;
    }

    return "transparent";
  }, [flashState, token.colorErrorBg, token.colorSuccessBg]);

  // Render Display Mode
  if (!isEditing && !isLoading) {
    return (
      <button
        type="button"
        className={`inline-edit-cell-display mode-${mode}`}
        onClick={!disabled ? handleEdit : undefined}
        onKeyDown={(e) => {
          if (!disabled && (e.key === "Enter" || e.key === " ")) {
            e.preventDefault();
            handleEdit();
          }
        }}
        disabled={disabled}
        data-testid="inline-edit-cell-display"
        data-flash-state={flashState}
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
          width: "100%",
          backgroundColor: flashBackground,
          textAlign: "left",
        }}
        onMouseEnter={(e) => {
          if (!disabled) {
            e.currentTarget.style.borderColor = token.colorBorder;
            if (flashState === "idle") {
              e.currentTarget.style.backgroundColor = token.colorBgTextHover;
            }
          }
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.borderColor = "transparent";
          e.currentTarget.style.backgroundColor = flashBackground;
        }}
        aria-label={`Edit ${mode}`}
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
      </button>
    );
  }

  // Render Edit Mode
  return (
    <div
      className="inline-edit-cell-edit"
      data-testid="inline-edit-cell-edit"
      data-flash-state={flashState}
      style={{
        position: "relative",
        backgroundColor: flashBackground,
        borderRadius: token.borderRadius,
        transition: "background-color 0.3s ease",
      }}
    >
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
                aria-label="Save value"
                onClick={(e) => {
                  e.stopPropagation();
                  void triggerSave();
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.stopPropagation();
                    void triggerSave();
                  }
                }}
                onMouseDown={(e) => {
                  e.preventDefault();
                }}
                tabIndex={0}
                role="button"
                style={{ cursor: "pointer", color: token.colorSuccess }}
              />
              <CloseOutlined
                aria-label="Cancel edit"
                onClick={(e) => {
                  e.stopPropagation();
                  handleCancelAction();
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.stopPropagation();
                    handleCancelAction();
                  }
                }}
                onMouseDown={(e) => {
                  e.preventDefault();
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
