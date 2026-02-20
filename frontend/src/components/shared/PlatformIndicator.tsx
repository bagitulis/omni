import React from "react";
import { theme, Tooltip } from "antd";
import type { GlobalToken } from "antd/es/theme/interface";
import type { PlatformIndicatorData, Platform } from "@/types/shared";

// Import platform icons directly
import shopeeIcon from "@/assets/icons/shopee.svg";
import tiktokIcon from "@/assets/icons/tiktok.webp";
import lazadaIcon from "@/assets/icons/lazada.webp";

export interface PlatformIndicatorProps {
  data: PlatformIndicatorData;
  size?: "small" | "default";
  onClick?: (platform: Platform) => void;
}

const PLATFORM_ICONS: Record<Platform, { icon: string; label: string }> = {
  shopee: { icon: shopeeIcon, label: "Shopee" },
  tiktok: { icon: tiktokIcon, label: "TikTok" },
  lazada: { icon: lazadaIcon, label: "Lazada" },
};

type VisualState =
  | "not_linked"
  | "linked"
  | "has_update"
  | "syncing"
  | "success"
  | "error";

function getStateBorderColors(
  token: GlobalToken,
): Record<VisualState, string> {
  return {
    linked: token.colorSuccess,
    has_update: token.colorWarning,
    syncing: token.colorPrimary,
    success: token.colorSuccess,
    error: token.colorError,
    not_linked: token.colorError,
  };
}

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

function getStateStyles(
  state: VisualState,
  token: GlobalToken,
): React.CSSProperties {
  const colors = getStateBorderColors(token);
  switch (state) {
    case "syncing":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${colors.syncing}`,
      };
    case "success":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${colors.success}`,
        animation: "popIn 0.3s ease",
      };
    case "error":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${colors.error}`,
        animation: "shake 0.5s ease",
      };
    case "not_linked":
      return {
        opacity: 0.45,
        filter: "grayscale(60%)",
        border: `2px solid ${colors.not_linked}`,
      };
    case "has_update":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${colors.has_update}`,
      };
    case "linked":
      return {
        opacity: 1,
        filter: "none",
        border: `2px solid ${colors.linked}`,
      };
    default:
      return { opacity: 1, border: "2px solid transparent" };
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

function getBadge(state: VisualState, token: GlobalToken): React.ReactNode {
  const badgeBase: React.CSSProperties = {
    position: "absolute",
    bottom: -2,
    right: -2,
    width: 10,
    height: 10,
    borderRadius: "50%",
    border: `1px solid ${token.colorBgContainer}`,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontSize: 6,
    color: token.colorBgContainer,
    fontWeight: 700,
  };

  if (state === "linked") {
    return (
      <span style={{ ...badgeBase, backgroundColor: token.colorSuccess }}>
        ✓
      </span>
    );
  }

  if (state === "has_update") {
    return (
      <span style={{ ...badgeBase, backgroundColor: token.colorWarning }}>
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
  const { token } = theme.useToken();
  const { icon, label } = PLATFORM_ICONS[data.platform];
  const dimensions = size === "small" ? 24 : 32;
  const iconSize = size === "small" ? 16 : 20;
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
    ...getStateStyles(visualState, token),
  };

  const tooltipText = getTooltipTitle(visualState, data, label);
  const badge = getBadge(visualState, token);

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
          <img
            src={icon}
            alt={label}
            width={iconSize}
            height={iconSize}
            style={{ objectFit: "contain", pointerEvents: "none" }}
          />
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
