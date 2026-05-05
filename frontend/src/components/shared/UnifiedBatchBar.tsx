import React from "react";
import { Button, Divider, Grid, Space, Tooltip, Typography, theme } from "antd";
import {
  CloseOutlined,
  CopyOutlined,
  DeleteOutlined,
  ExportOutlined,
  ImportOutlined,
  ShopOutlined,
} from "@ant-design/icons";
import type { BatchActionType } from "@/types/shared";

export interface UnifiedBatchBarProps {
  selectedCount: number;
  onAction: (actionKey: BatchActionType) => void;
  onClearSelection: () => void;
  disabledActions?: Partial<Record<BatchActionType, string>>;
  loadingActions?: Partial<Record<BatchActionType, boolean>>;
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
    key: "push_to_marketplace",
    label: "Push to Marketplace",
    icon: <ExportOutlined />,
    tooltip: "Push stock and prices to Shopee, TikTok, Lazada",
  },
  {
    key: "sync_marketplace",
    label: "Pull from Marketplace",
    icon: <ImportOutlined />,
    tooltip: "Refresh product data from marketplaces",
  },
  {
    key: "bulk_pricing",
    label: "Bulk Pricing",
    icon: <ShopOutlined />,
    tooltip: "Wholesale, MPQ, Reset, and Settings",
  },
  {
    key: "clone",
    label: "Clone",
    icon: <CopyOutlined />,
    tooltip: "Clone products to other platforms",
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
  loadingActions = {},
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
          const isLoading = !!loadingActions[action.key];

          const actionButton = (
            <Button
              size="small"
              icon={action.icon}
              danger={action.danger}
              disabled={isDisabled || isLoading}
              loading={isLoading}
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
