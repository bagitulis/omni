import React from "react";
import { Button, Divider, Grid, Space, Tooltip, Typography, theme } from "antd";
import {
  CloseOutlined,
  CloudSyncOutlined,
  CopyOutlined,
  DeleteOutlined,
  DollarOutlined,
  FieldNumberOutlined,
  ShopOutlined,
  SyncOutlined,
} from "@ant-design/icons";
import type { BatchActionType } from "@/types/shared";

export interface UnifiedBatchBarProps {
  selectedCount: number;
  onAction: (actionKey: BatchActionType) => void;
  onClearSelection: () => void;
  disabledActions?: Partial<Record<BatchActionType, string>>;
}

interface UnifiedBatchAction {
  key: BatchActionType;
  label: string;
  icon: React.ReactNode;
  danger?: boolean;
  tooltip: string;
}

const UNIFIED_BATCH_ACTIONS: UnifiedBatchAction[] = [
  {
    key: "sync_stock",
    label: "Sync Stock",
    icon: <SyncOutlined />,
    tooltip: "Sync stock to selected marketplaces",
  },
  {
    key: "sync_marketplace",
    label: "Sync Marketplace",
    icon: <CloudSyncOutlined />,
    tooltip: "Refresh price/stock from Shopee, TikTok, Lazada",
  },
  {
    key: "update_price",
    label: "Update Price",
    icon: <DollarOutlined />,
    tooltip: "Update price on all linked marketplaces",
  },
  {
    key: "wholesale",
    label: "Wholesale",
    icon: <ShopOutlined />,
    tooltip: "Set wholesale tiers (Shopee only)",
  },
  {
    key: "mpq",
    label: "MPQ",
    icon: <FieldNumberOutlined />,
    tooltip: "Set minimum purchase quantity (Shopee + TikTok)",
  },
  {
    key: "clone",
    label: "Clone",
    icon: <CopyOutlined />,
    tooltip: "Clone products to other platforms",
  },
  {
    key: "delete_wholesale",
    label: "Del Wholesale",
    icon: <DeleteOutlined />,
    danger: true,
    tooltip: "Delete wholesale configs (Shopee only)",
  },
  {
    key: "delete_products",
    label: "Delete",
    icon: <DeleteOutlined />,
    danger: true,
    tooltip: "Delete selected products",
  },
];

export const UnifiedBatchBar: React.FC<UnifiedBatchBarProps> = ({
  selectedCount,
  onAction,
  onClearSelection,
  disabledActions = {},
}) => {
  const { token } = theme.useToken();
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  if (selectedCount <= 0) {
    return null;
  }

  return (
    <div
      style={{
        position: "fixed",
        bottom: isMobile ? 0 : 24,
        left: isMobile ? 0 : "50%",
        transform: isMobile ? "none" : "translateX(-50%)",
        width: isMobile ? "100%" : "auto",
        zIndex: 1000,
        backgroundColor: token.colorBgElevated,
        borderRadius: isMobile ? 0 : 3,
        borderTop: `1px solid ${token.colorBorderSecondary}`,
        borderBottom: isMobile
          ? "none"
          : `1px solid ${token.colorBorderSecondary}`,
        borderLeft: isMobile
          ? "none"
          : `1px solid ${token.colorBorderSecondary}`,
        borderRight: isMobile
          ? "none"
          : `1px solid ${token.colorBorderSecondary}`,
        padding: "12px 16px",
        maxWidth: isMobile ? "100%" : "calc(100vw - 48px)",
        overflowX: "auto",
        display: "flex",
        alignItems: "center",
        gap: 12,
        boxShadow: isMobile
          ? "0 -2px 8px rgba(0,0,0,0.15)"
          : token.boxShadowSecondary,
      }}
      role="toolbar"
      aria-label="Batch actions"
    >
      <Typography.Text
        strong
        style={{ color: token.colorPrimary, whiteSpace: "nowrap" }}
      >
        {selectedCount} selected
      </Typography.Text>

      <Divider type="vertical" style={{ height: 24, margin: 0 }} />

      <Space size={8} wrap>
        {UNIFIED_BATCH_ACTIONS.map((action) => {
          const isDisabled = Object.prototype.hasOwnProperty.call(
            disabledActions,
            action.key,
          );
          const disabledReason = disabledActions[action.key];

          const actionButton = (
            <Button
              size="small"
              icon={action.icon}
              danger={action.danger}
              disabled={isDisabled}
              onClick={() => onAction(action.key)}
              aria-label={action.label}
            >
              {action.label}
            </Button>
          );

          return (
            <Tooltip
              key={action.key}
              title={
                isDisabled
                  ? (disabledReason ?? "This action is unavailable")
                  : action.tooltip
              }
            >
              <span>{actionButton}</span>
            </Tooltip>
          );
        })}
      </Space>

      <Button
        type="text"
        size="small"
        icon={<CloseOutlined />}
        onClick={onClearSelection}
        aria-label="Clear selection"
        style={{ marginLeft: "auto" }}
      >
        Clear
      </Button>
    </div>
  );
};
