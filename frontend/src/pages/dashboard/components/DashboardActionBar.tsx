import { Flex, Button, Badge, Space, Tooltip } from "antd";
import {
  KeyOutlined,
  ExportOutlined,
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
    <Flex align="center" gap={16} wrap="wrap" justify="flex-end">
      <div>{getStatusBadge()}</div>
      <Space size="small" wrap>
        <Tooltip title="Manage API Tokens">
          <Button
            icon={<KeyOutlined />}
            onClick={() => openModal("token")}
            data-testid="token-modal-trigger"
          >
            Token
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
      </Space>
    </Flex>
  );
}
