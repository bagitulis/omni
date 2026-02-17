import React from "react";
import { Tooltip, theme } from "antd";
import {
  CheckCircleFilled,
  CloseCircleFilled,
  ExclamationCircleFilled,
  LoadingOutlined,
  DisconnectOutlined,
} from "@ant-design/icons";
import type { PlatformIndicatorData, Platform } from "@/types/shared";

export interface PlatformIndicatorProps {
  data: PlatformIndicatorData;
  size?: "small" | "default";
  onClick?: (platform: Platform) => void;
  showLabel?: boolean; // Kept for backward compatibility
}

const PLATFORM_CONFIG: Record<
  Platform,
  { label: string; color: string; icon: string }
> = {
  shopee: { label: "Shopee", color: "#ee4d2d", icon: "S" },
  tiktok: { label: "TikTok", color: "#000000", icon: "T" },
  lazada: { label: "Lazada", color: "#0f146d", icon: "L" },
};

export function PlatformIndicator({
  data,
  size = "default",
  onClick,
}: PlatformIndicatorProps) {
  const { token } = theme.useToken();
  const config = PLATFORM_CONFIG[data.platform];

  const sizePx = size === "small" ? 24 : 32;
  const fontSize = size === "small" ? 10 : 12;
  const badgeSize = size === "small" ? 10 : 12;

  // Visual State Logic
  const containerStyle: React.CSSProperties = {
    position: "relative",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    width: sizePx,
    height: sizePx,
    borderRadius: "6px", // <= 3-6px
    border: `1px solid ${token.colorBorder}`, // default border
    backgroundColor: token.colorBgContainer,
    transition: "all 0.2s ease-in-out",
    cursor: onClick ? "pointer" : "default",
    overflow: "visible",
  };

  let badgeIcon: React.ReactNode = null;
  const badgeStyle: React.CSSProperties = {
    position: "absolute",
    bottom: -4,
    right: -4,
    backgroundColor: token.colorBgContainer,
    borderRadius: "50%",
    width: badgeSize + 2,
    height: badgeSize + 2,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontSize: badgeSize,
    lineHeight: 1,
    zIndex: 1,
  };

  let tooltipText = config.label;
  const platformIconStyle: React.CSSProperties = {
    color: config.color,
    fontSize: fontSize + 4,
    fontWeight: 700,
  };

  // 1. NOT_LINKED: opacity 0.2 + grayscale(100%)
  if (!data.linked) {
    containerStyle.opacity = 0.2;
    containerStyle.filter = "grayscale(100%)";
    tooltipText = `${config.label}: Not Linked`;
    badgeIcon = (
      <DisconnectOutlined
        style={{ fontSize: badgeSize, color: token.colorTextDisabled }}
      />
    );
  } else {
    // Linked base state
    containerStyle.backgroundColor = token.colorBgContainer;

    if (data.sync_state === "error") {
      // 6. ERROR: red border + shake animation + error tooltip
      containerStyle.borderColor = token.colorError;
      containerStyle.animation = "shake 0.4s ease-in-out";
      tooltipText = `Error: ${data.error_message || "Sync failed"}`;
      badgeIcon = <CloseCircleFilled style={{ color: token.colorError }} />;
    } else if (data.sync_state === "syncing") {
      // 4. SYNCING: blue border + spinner overlay
      containerStyle.borderColor = token.colorPrimary;
      tooltipText = "Syncing...";
      badgeIcon = <LoadingOutlined style={{ color: token.colorPrimary }} />;
    } else if (data.has_update) {
      // 3. HAS_UPDATE: amber border + exclamation badge
      containerStyle.borderColor = token.colorWarning;
      tooltipText = "Update Available";
      badgeIcon = (
        <ExclamationCircleFilled style={{ color: token.colorWarning }} />
      );
    } else if (data.sync_state === "success") {
      // 5. SUCCESS: green border + popIn animation
      containerStyle.borderColor = token.colorSuccess;
      containerStyle.animation =
        "popIn 0.3s cubic-bezier(0.18, 0.89, 0.32, 1.28)";
      tooltipText = "Synced";
      badgeIcon = <CheckCircleFilled style={{ color: token.colorSuccess }} />;
    } else {
      // 2. LINKED idle: green border (implied by success usually, but if just linked without success/error...)
      // Fallback for "idle" but linked
      containerStyle.borderColor = token.colorSuccess;
      badgeIcon = <CheckCircleFilled style={{ color: token.colorSuccess }} />;
    }
  }

  return (
    <Tooltip title={tooltipText}>
      {/* Interactive elements must be focusable and have semantics */}
      <div
        role={onClick ? "button" : undefined}
        tabIndex={onClick ? 0 : -1}
        style={containerStyle}
        onClick={() => onClick?.(data.platform)}
        onKeyDown={(e) => {
          if (onClick && (e.key === "Enter" || e.key === " ")) {
            e.preventDefault();
            onClick(data.platform);
          }
        }}
        data-testid={`platform-indicator-${data.platform}`}
      >
        <span style={platformIconStyle}>{config.icon}</span>
        {badgeIcon && <div style={badgeStyle}>{badgeIcon}</div>}
      </div>
      <style>{`
        @keyframes popIn {
          0% { transform: scale(0.8); opacity: 0; }
          100% { transform: scale(1); opacity: 1; }
        }
        @keyframes shake {
          0%, 100% { transform: translateX(0); }
          10%, 30%, 50%, 70%, 90% { transform: translateX(-2px); }
          20%, 40%, 60%, 80% { transform: translateX(2px); }
        }
      `}</style>
    </Tooltip>
  );
}
