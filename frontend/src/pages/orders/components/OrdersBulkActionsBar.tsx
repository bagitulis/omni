import { Button, Card, Divider, Flex, Space, Typography } from "antd";
import { PrinterOutlined, SendOutlined } from "@ant-design/icons";

const { Text } = Typography;

interface OrdersBulkActionsBarProps {
  selectedCount: number;
  onBulkShip: () => void;
  onBulkPrint: () => void;
  onBulkCancel: () => void;
  onClearSelection: () => void;
  isShipping: boolean;
  isPrinting: boolean;
  isCancelling: boolean;
}

export function OrdersBulkActionsBar({
  selectedCount,
  onBulkShip,
  onBulkPrint,
  onBulkCancel,
  onClearSelection,
  isShipping,
  isPrinting,
  isCancelling,
}: OrdersBulkActionsBarProps) {
  if (selectedCount <= 0) return null;

  return (
    <Card
      size="small"
      style={{
        backgroundColor: "#f0f9ff",
        border: "1px solid #bae6fd",
        borderRadius: 4,
      }}
    >
      <Flex justify="space-between" align="center">
        <Space split={<Divider type="vertical" />}>
          <Text strong style={{ color: "#0369a1" }}>
            {selectedCount} orders selected
          </Text>
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={onBulkShip}
            loading={isShipping}
          >
            Bulk Ship
          </Button>
          <Button
            icon={<PrinterOutlined />}
            onClick={onBulkPrint}
            loading={isPrinting}
          >
            Bulk Print Labels
          </Button>
          <Button danger onClick={onBulkCancel} loading={isCancelling}>
            Bulk Cancel
          </Button>
        </Space>
        <Button type="text" onClick={onClearSelection}>
          Clear Selection
        </Button>
      </Flex>
    </Card>
  );
}
