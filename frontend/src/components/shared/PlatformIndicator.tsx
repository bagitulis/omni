import React from "react";
import { Tooltip } from "antd";
import type { PlatformIndicatorData, Platform } from "@/types/shared";

export interface PlatformIndicatorProps {
  data: PlatformIndicatorData;
  size?: "small" | "default";
  onClick?: (platform: Platform) => void;
}

const PLATFORM_ICONS: Record<Platform, { emoji: string; label: string }> = {
  shopee: { emoji: "🟠", label: "Shopee" },
  tiktok: { emoji: "⬛", label: "TikTok" },
  lazada: { emoji: "🔵", label: "Lazada" },
};

type VisualState =
  | "not_linked"
  | "linked"
  | "has_update"
  | "syncing"
  | "success"
  | "error";

const STATE_BORDER_COLORS: Record<
  Exclude<VisualState, "not_linked">,
  string
> = {
  linked: "#52c41a",
  has_update: "#faad14",
  syncing: "#1890ff",
  success: "#52c41a",
  error: "#ff4d4f",
};

function getVisualState(data: PlatformIndicatorData): VisualState {
  if (data.sync_state === "syncing") {
    return "syncing";
  }
  if (data.sync_state === "success") {
    return "success";
  }
  if (data.sync_state === "error") {
    return "error";
  }
  if (!data.linked) {
    return "not_linked";
  }
  if (data.has_update) {
    return "has_update";
  }
  return "linked";
}

function getStateStyles(state: VisualState): React.CSSProperties {
  switch (state) {
    case "syncing":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${STATE_BORDER_COLORS.syncing}`,
      };
    case "success":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${STATE_BORDER_COLORS.success}`,
        animation: "popIn 0.3s ease",
      };
    case "error":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${STATE_BORDER_COLORS.error}`,
        animation: "shake 0.5s ease",
      };
    case "not_linked":
      return {
        opacity: 0.2,
        filter: "grayscale(100%)",
        border: "2px solid transparent",
      };
    case "has_update":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${STATE_BORDER_COLORS.has_update}`,
      };
    case "linked":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${STATE_BORDER_COLORS.linked}`,
      };
    default:
      return {
        opacity: 1,
        border: "2px solid transparent",
      };
  }
}

function getTooltipTitle(
  state: VisualState,
  data: PlatformIndicatorData,
  label: string,
): string {
  if (state === "error") {
    return `${label}: ${data.error_message ?? "Sync error"}`;
  }
  if (state === "syncing") {
    return `${label}: Syncing...`;
  }
  if (state === "success") {
    return `${label}: Sync successful`;
  }
  if (state === "not_linked") {
    return `${label}: Not linked`;
  }
  if (state === "has_update") {
    return `${label}: Has pending updates`;
  }

  const syncedAtText = data.last_synced_at
    ? ` (last sync: ${new Date(data.last_synced_at).toLocaleString()})`
    : "";

  return `${label}: Linked${syncedAtText}`;
}

function getBadge(state: VisualState): React.ReactNode {
  if (state === "linked") {
    return (
      <span
        style={{
          position: "absolute",
          bottom: -2,
          right: -2,
          width: 10,
          height: 10,
          borderRadius: "50%",
          backgroundColor: "#52c41a",
          border: "1px solid #ffffff",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: 6,
          color: "#ffffff",
          fontWeight: 700,
        }}
      >
        ✓
      </span>
    );
  }

  if (state === "has_update") {
    return (
      <span
        style={{
          position: "absolute",
          bottom: -2,
          right: -2,
          width: 10,
          height: 10,
          borderRadius: "50%",
          backgroundColor: "#faad14",
          border: "1px solid #ffffff",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: 6,
          color: "#ffffff",
          fontWeight: 700,
        }}
      >
        !
      </span>
    );
  }

  return null;
}

export function PlatformIndicator({
  data,
  size = "default",
  onClick,
}: PlatformIndicatorProps) {
  const { emoji, label } = PLATFORM_ICONS[data.platform];
  const dimensions = size === "small" ? 24 : 32;
  const visualState = getVisualState(data);
  const isInteractive = data.linked && Boolean(onClick);

  const containerStyle: React.CSSProperties = {
    position: "relative",
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    width: dimensions,
    height: dimensions,
    borderRadius: 3,
    cursor: isInteractive ? "pointer" : "default",
    transition: "all 0.2s ease",
    fontSize: size === "small" ? 14 : 18,
    lineHeight: 1,
    padding: 0,
    background: "transparent",
    ...getStateStyles(visualState),
  };

  const tooltipText = getTooltipTitle(visualState, data, label);
  const badge = getBadge(visualState);

  const handleClick = () => {
    if (isInteractive) {
      onClick?.(data.platform);
    }
  };

  return (
    <Tooltip title={tooltipText}>
      <span style={{ display: "inline-flex" }}>
        <button
          type="button"
          tabIndex={isInteractive ? 0 : -1}
          disabled={!isInteractive}
          style={containerStyle}
          onClick={handleClick}
          data-testid={`platform-indicator-${data.platform}`}
          data-state={visualState}
          aria-label={`${label} platform indicator`}
        >
          {emoji}
          {visualState === "syncing" && (
            <span
              style={{
                position: "absolute",
                inset: 0,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                backgroundColor: "rgba(255,255,255,0.7)",
                borderRadius: 3,
              }}
            >
              <span className="platform-spinner" />
            </span>
          )}
          {badge}
        </button>
      </span>
    </Tooltip>
  );
}
