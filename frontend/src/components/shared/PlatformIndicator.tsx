import { Tooltip, Space, theme } from "antd";

import {
  CheckCircleFilled,
  CloseCircleFilled,
  ExclamationCircleFilled,
  LoadingOutlined,
  DisconnectOutlined,
  LinkOutlined,
} from "@ant-design/icons";
import type { PlatformIndicatorData, Platform } from "@/types/shared";

interface PlatformIndicatorProps {
  data: PlatformIndicatorData;
  showLabel?: boolean;
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
  showLabel = false,
}: PlatformIndicatorProps) {
  const { token } = theme.useToken();
  const config = PLATFORM_CONFIG[data.platform];

  // Determine visual state
  let statusColor = token.colorTextDisabled; // Default/Not Linked
  let statusIcon = <DisconnectOutlined />;
  let tooltipText = `${config.label}: Not Linked`;

  if (data.linked) {
    switch (data.sync_state) {
      case "success":
        statusColor = config.color; // Brand color when healthy
        statusIcon = <CheckCircleFilled />;
        tooltipText = `${config.label}: Synced`;
        break;
      case "syncing":
        statusColor = token.colorPrimary;
        statusIcon = <LoadingOutlined />;
        tooltipText = `${config.label}: Syncing...`;
        break;
      case "error":
        statusColor = token.colorError;
        statusIcon = <CloseCircleFilled />;
        tooltipText = `${config.label}: Error - ${data.error_message || "Unknown error"}`;
        break;
      default:
        statusColor = config.color;
        statusIcon = <LinkOutlined />;
        tooltipText = `${config.label}: Linked`;
    }
  }

  // Handle specific "pending" link status if needed, though derivePlatformStatus maps it to syncing
  if (data.has_update && data.sync_state !== "syncing") {
    statusIcon = <ExclamationCircleFilled />;
    statusColor = token.colorWarning;
    tooltipText = `${config.label}: Update Available`;
  }

  return (
    <Tooltip title={tooltipText}>
      <Space
        align="center"
        size={4}
        style={{
          cursor: "help",
          opacity: data.linked ? 1 : 0.5,
          transition: "all 0.2s",
        }}
      >
        {/* Platform Icon/Badge */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: 24,
            height: 24,
            borderRadius: "50%",
            backgroundColor: data.linked
              ? config.color
              : token.colorFillSecondary,
            color: data.linked ? "#fff" : token.colorTextDisabled,
            fontSize: 12,
            fontWeight: 600,
            border: `1px solid ${data.linked ? "transparent" : token.colorBorder}`,
          }}
        >
          {config.icon}
        </div>

        {/* Status Indicator (Small overlay or side icon) */}
        <div style={{ color: statusColor, fontSize: 14, lineHeight: 1 }}>
          {statusIcon}
        </div>

        {/* Optional Label */}
        {showLabel && (
          <span
            style={{
              color: data.linked ? token.colorText : token.colorTextDisabled,
              fontSize: token.fontSizeSM,
            }}
          >
            {config.label}
          </span>
        )}
      </Space>
    </Tooltip>
  );
}
