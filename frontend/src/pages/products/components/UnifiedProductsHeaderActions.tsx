import {
  CloudDownloadOutlined,
  ImportOutlined,
  PlusOutlined,
  UploadOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import { Badge, Button, Tooltip, Typography } from "antd";
import type { FC } from "react";
import type { NavigateFunction } from "react-router-dom";

interface UnifiedProductsHeaderActionsProps {
  navigate: NavigateFunction;
  syncHistoryTotal: number;
  priceDriftCount?: number;
  onSyncDriftedPrices?: () => void;
  onPullFromMarketplace?: () => void;
  pullLoading?: boolean;
}

export const UnifiedProductsHeaderActions: FC<
  UnifiedProductsHeaderActionsProps
> = ({ navigate, syncHistoryTotal, priceDriftCount = 0, onSyncDriftedPrices, onPullFromMarketplace, pullLoading = false }) => {
  return (
    <div
      style={{
        display: "flex",
        flexWrap: "wrap",
        gap: 16,
        alignItems: "center",
        justifyContent: "space-between",
        width: "100%",
      }}
    >
      <Typography.Title level={4} style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>
        Products
      </Typography.Title>
      <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
        <Button
          icon={<PlusOutlined />}
          type="primary"
          onClick={() => navigate("/master-products/add")}
        >
          Add Product
        </Button>
        <Button
          icon={<UploadOutlined />}
          onClick={() => navigate("/master-products/import")}
        >
          Import
        </Button>
        <Button
          icon={<CloudDownloadOutlined />}
          onClick={() => navigate("/products/sync-history")}
        >
          Sync History ({syncHistoryTotal})
        </Button>
        <Button
          icon={<ImportOutlined />}
          onClick={onPullFromMarketplace}
          loading={pullLoading}
        >
          Pull from Marketplace
        </Button>
        {priceDriftCount > 0 && (
          <Tooltip title={`${priceDriftCount} prices out of sync with inventory`}>
            <Badge count={priceDriftCount} size="small" offset={[4, 0]}>
              <Button
                icon={<WarningOutlined />}
                type="default"
                style={{ color: "#fa8c16" }}
                onClick={onSyncDriftedPrices}
              >
                Price Drift
              </Button>
            </Badge>
          </Tooltip>
        )}
      </div>
    </div>
  );
};
