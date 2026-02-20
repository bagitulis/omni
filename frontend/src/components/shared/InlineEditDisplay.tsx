import React, { useMemo } from "react";
import { Typography, theme } from "antd";
import { EditOutlined } from "@ant-design/icons";

const { Text } = Typography;

type FlashState = "idle" | "success" | "error";

interface InlineEditDisplayProps {
  value: number;
  prefix?: React.ReactNode;
  disabled: boolean;
  flashState: FlashState;
  onEdit: () => void;
}

/**
 * Read-only display mode for InlineEditCell.
 * Extracted per SRP — display rendering is a distinct concern from edit interaction.
 */
export function InlineEditDisplay({
  value,
  prefix,
  disabled,
  flashState,
  onEdit,
}: InlineEditDisplayProps) {
  const { token } = theme.useToken();

  const flashBackground = useMemo(() => {
    if (flashState === "success") return token.colorSuccessBg;
    if (flashState === "error") return token.colorErrorBg;
    return "transparent";
  }, [flashState, token.colorErrorBg, token.colorSuccessBg]);

  return (
    <button
      type="button"
      className="inline-edit-cell-display"
      onClick={!disabled ? onEdit : undefined}
      onKeyDown={(e) => {
        if (!disabled && (e.key === "Enter" || e.key === " ")) {
          e.preventDefault();
          onEdit();
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
        border: "1px solid transparent",
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
      aria-label="Edit value"
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
