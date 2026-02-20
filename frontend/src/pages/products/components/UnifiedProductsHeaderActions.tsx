import {
  CloudDownloadOutlined,
  PlusOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import { Button, Typography } from "antd";
import type { FC } from "react";
import type { NavigateFunction } from "react-router-dom";

interface UnifiedProductsHeaderActionsProps {
  navigate: NavigateFunction;
  syncHistoryTotal: number;
}

export const UnifiedProductsHeaderActions: FC<
  UnifiedProductsHeaderActionsProps
> = ({ navigate, syncHistoryTotal }) => {
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
      </div>
    </div>
  );
};
