import { Flex, Button, Badge, Space, Tooltip } from "antd";
import {
  KeyOutlined,
  DollarOutlined,
  ExportOutlined,
  WalletOutlined,
  CarOutlined,
} from "@ant-design/icons";
import { useModalsStore } from "@/stores/modalsStore";
import { useConnectionStatus } from "@/hooks/useConnectionStatus";

export function DashboardActionBar() {
  const { openModal } = useModalsStore();
  const { connectionStatus } = useConnectionStatus();

  const getStatusBadge = () => {
    switch (connectionStatus) {
      case "connected":
        return <Badge status="success" text="Connected" />;
      case "connecting":
        return <Badge status="processing" text="Connecting" />;
      case "error":
        return <Badge status="error" text="Error" />;
      case "disconnected":
        return <Badge status="default" text="Disconnected" />;
      default:
        return <Badge status="default" text="Unknown" />;
    }
  };

  return (
    <Flex justify="space-between" align="center" style={{ width: "100%" }}>
      <Space size="middle">
        <Tooltip title="Manage API Tokens">
          <Button
            icon={<KeyOutlined />}
            onClick={() => openModal("token")}
            data-testid="token-modal-trigger"
          >
            Token
          </Button>
        </Tooltip>
        <Tooltip title="Update Product Prices">
          <Button
            icon={<DollarOutlined />}
            onClick={() => openModal("price")}
            data-testid="price-modal-trigger"
          >
            Price
          </Button>
        </Tooltip>
        <Tooltip title="Export Order Data">
          <Button
            icon={<ExportOutlined />}
            onClick={() => openModal("exportOrders")}
            data-testid="export-modal-trigger"
          >
            Export
          </Button>
        </Tooltip>
        <Tooltip title="Wallet & Balance">
          <Button
            icon={<WalletOutlined />}
            onClick={() => openModal("wallet")}
            data-testid="wallet-modal-trigger"
          >
            Wallet
          </Button>
        </Tooltip>
        <Tooltip title="Shipping Configuration">
          <Button
            icon={<CarOutlined />}
            onClick={() => openModal("shipping")}
            data-testid="shipping-modal-trigger"
          >
            Shipping
          </Button>
        </Tooltip>
      </Space>
      <div>{getStatusBadge()}</div>
    </Flex>
  );
}
