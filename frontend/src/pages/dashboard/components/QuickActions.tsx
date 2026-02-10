import { Card, Button, Row, Col, Space, Dropdown } from "antd";
import {
  ThunderboltOutlined,
  ExportOutlined,
  SyncOutlined,
  DollarOutlined,
  ShoppingOutlined,
  DownOutlined,
} from "@ant-design/icons";

import { useModalsStore } from "@/stores/modalsStore";

export function QuickActions() {
  const { openModal } = useModalsStore();

  const handleExportOrders = (platform: string) => {
    openModal("exportOrders", { platform });
  };

  const handleUpdatePrice = () => {
    openModal("price");
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
        <Space>
          <ThunderboltOutlined />
          <span>Quick Actions</span>
        </Space>
      }
    >
      <Row gutter={[12, 12]}>
        <Col span={12}>
          <Button
            type="primary"
            block
            icon={<SyncOutlined />}
            onClick={() => console.log("Sync All")}
          >
            Sync All
          </Button>
        </Col>
        <Col span={12}>
          <Dropdown menu={{ items }} trigger={["click"]}>
            <Button block icon={<ExportOutlined />}>
              <Space>
                Export Orders
                <DownOutlined />
              </Space>
            </Button>
          </Dropdown>
        </Col>
        <Col span={12}>
          <Button
            block
            icon={<DollarOutlined />}
            onClick={() => handleUpdatePrice()}
          >
            Update Prices
          </Button>
        </Col>
        <Col span={12}>
          <Button block icon={<ShoppingOutlined />}>
            Manage Stock
          </Button>
        </Col>
      </Row>
    </Card>
  );
}
