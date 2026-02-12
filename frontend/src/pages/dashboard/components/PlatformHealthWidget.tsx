import { Card, List, Tag, Typography, Space, theme } from "antd";
import {
  ApiOutlined,
  CheckCircleFilled,
  CloseCircleFilled,
  SyncOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons";

import { useSyncStatus } from "@/hooks/useDashboardWidgets";

const { Text } = Typography;

export function PlatformHealthWidget() {
  const { token } = theme.useToken();
  const { data: shopeeStatus, isLoading: isShopeeLoading } =
    useSyncStatus("shopee");
  const { data: tiktokStatus, isLoading: isTiktokLoading } =
    useSyncStatus("tiktok");

  const platforms = [
    { name: "Shopee", data: shopeeStatus, loading: isShopeeLoading },
    { name: "TikTok", data: tiktokStatus, loading: isTiktokLoading },
    {
      name: "Lazada",
      data: { status: "disconnected", last_sync: null } as any,
      loading: false,
    },
  ];

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "connected":
        return <CheckCircleFilled style={{ color: token.colorSuccess }} />;
      case "disconnected":
        return <CloseCircleFilled style={{ color: token.colorError }} />;
      case "syncing":
        return <SyncOutlined spin style={{ color: token.colorPrimary }} />;
      case "error":
        return <CloseCircleFilled style={{ color: token.colorError }} />;
      default:
        return <ApiOutlined />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "connected":
        return "success";
      case "disconnected":
        return "error";
      case "syncing":
        return "processing";
      default:
        return "default";
    }
  };

  return (
    <Card
      title={
        <Space>
          <ApiOutlined />
          <span>Platform Health</span>
        </Space>
      }
      bodyStyle={{ padding: 0 }}
    >
      <List
        dataSource={platforms}
        renderItem={(item) => (
          <List.Item style={{ padding: "12px 24px" }}>
            <List.Item.Meta
              avatar={getStatusIcon(item.data?.status || "disconnected")}
              title={
                <Space>
                  <Text strong>{item.name}</Text>
                  <Tag
                    color={getStatusColor(item.data?.status || "disconnected")}
                    bordered={false}
                  >
                    {(item.data?.status || "disconnected").toUpperCase()}
                  </Tag>
                </Space>
              }
              description={
                item.data?.last_sync ? (
                  <Space size={4}>
                    <ClockCircleOutlined style={{ fontSize: 10 }} />
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      Synced: {new Date(item.data.last_sync).toLocaleString()}
                    </Text>
                  </Space>
                ) : (
                  <Text type="secondary" style={{ fontSize: 11 }}>
                    Not connected
                  </Text>
                )
              }
            />
          </List.Item>
        )}
      />
    </Card>
  );
}
