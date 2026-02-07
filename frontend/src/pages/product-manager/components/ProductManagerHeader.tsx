import { Button, Flex, Switch, Typography } from "antd";
import { SyncOutlined } from "@ant-design/icons";

const { Title, Text } = Typography;

interface ProductManagerHeaderProps {
  autoRefresh: boolean;
  setAutoRefresh: (val: boolean) => void;
  autoSync: boolean;
  setAutoSync: (val: boolean) => void;
  syncLoading: boolean;
  onSync: () => void;
}

export function ProductManagerHeader({
  autoRefresh,
  setAutoRefresh,
  autoSync,
  setAutoSync,
  syncLoading,
  onSync,
}: ProductManagerHeaderProps) {
  return (
    <Flex justify="space-between" align="center">
      <div>
        <Title level={3} style={{ margin: 0 }}>
          Product Manager
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Platform products synced into database
        </Text>
      </div>
      <Flex align="center" gap={12}>
        <Flex align="center" gap={6}>
          <Text style={{ fontSize: 12 }}>Auto refresh</Text>
          <Switch
            checked={autoRefresh}
            onChange={setAutoRefresh}
            size="small"
          />
        </Flex>
        <Flex align="center" gap={6}>
          <Text style={{ fontSize: 12 }}>Auto sync</Text>
          <Switch checked={autoSync} onChange={setAutoSync} size="small" />
        </Flex>
        <Button
          data-testid="product-manager-sync"
          type="primary"
          icon={<SyncOutlined />}
          loading={syncLoading}
          onClick={onSync}
        >
          Sync Now
        </Button>
      </Flex>
    </Flex>
  );
}
