import { Card, Button, Row, Col, Space, Dropdown } from "antd";
import {
  ThunderboltOutlined,
  ExportOutlined,
  SyncOutlined,
  ShoppingOutlined,
  DownOutlined,
} from "@ant-design/icons";

import { useModalsStore } from "@/stores/modalsStore";
import { logger } from "@/lib/logger";

export function QuickActions() {
  const { openModal } = useModalsStore();

  const handleExportOrders = (platform: string) => {
    openModal("exportOrders", { platform });
  };

  const items = [
    {
      key: "shopee",
      label: "Shopee",
      onClick: () => handleExportOrders("shopee"),
    },
    {
      key: "tiktok",
      label: "TikTok",
      onClick: () => handleExportOrders("tiktok"),
    },
    {
      key: "lazada",
      label: "Lazada",
      onClick: () => handleExportOrders("lazada"),
    },
  ];

  return (
    <Card
      title={
        <Space style={{ fontSize: 14 }}>
          <ThunderboltOutlined style={{ color: "#d97706" }} />
          <span style={{ fontWeight: 600 }}>Quick Actions</span>
        </Space>
      }
      style={{ borderRadius: 3, border: "1px solid #f0f0f0" }}
      bodyStyle={{ padding: "16px 24px" }}
    >
      <Row gutter={[16, 16]}>
        <Col span={12}>
          <Button
            type="primary"
            block
            icon={<SyncOutlined />}
            onClick={() => logger.debug("Sync All")}
            style={{ borderRadius: 3, height: 40, fontSize: 12, fontWeight: 500 }}
          >
            Sync All
          </Button>
        </Col>
        <Col span={12}>
          <Dropdown menu={{ items }} trigger={["click"]}>
            <Button block icon={<ExportOutlined />} style={{ borderRadius: 3, height: 40, fontSize: 12 }}>
              <Space>
                Export
                <DownOutlined style={{ fontSize: 10 }} />
              </Space>
            </Button>
          </Dropdown>
        </Col>
        <Col span={24}>
          <Button block icon={<ShoppingOutlined />} style={{ borderRadius: 3, height: 40, fontSize: 12 }}>
            Manage Stock
          </Button>
        </Col>
      </Row>
    </Card>
  );
}
