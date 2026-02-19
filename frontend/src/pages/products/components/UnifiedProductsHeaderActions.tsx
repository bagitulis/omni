import {
  CloudDownloadOutlined,
  PlusOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import { Button, Space, Typography } from "antd";
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
    <Space style={{ width: "100%", justifyContent: "space-between" }} wrap>
      <Typography.Title level={2} style={{ margin: 0 }}>
        Products
      </Typography.Title>
      <Space wrap>
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
      </Space>
    </Space>
  );
};
