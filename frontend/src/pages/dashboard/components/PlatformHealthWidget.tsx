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
  const { data: lazadaStatus, isLoading: isLazadaLoading } =
    useSyncStatus("lazada");

  const platforms = [
    { name: "Shopee", data: shopeeStatus, loading: isShopeeLoading },
    { name: "TikTok", data: tiktokStatus, loading: isTiktokLoading },
    { name: "Lazada", data: lazadaStatus, loading: isLazadaLoading },
  ];

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "connected":
        return <CheckCircleFilled style={{ color: token.colorSuccess, fontSize: 18 }} />;
      case "disconnected":
        return <CloseCircleFilled style={{ color: token.colorError, fontSize: 18 }} />;
      case "syncing":
        return <SyncOutlined spin style={{ color: token.colorPrimary, fontSize: 18 }} />;
      case "error":
        return <CloseCircleFilled style={{ color: token.colorError, fontSize: 18 }} />;
      default:
        return <ApiOutlined style={{ fontSize: 18 }} />;
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
        <Space style={{ fontSize: 14 }}>
          <ApiOutlined style={{ color: token.colorPrimary }} />
          <Text strong style={{ fontSize: 14 }}>Platform Health</Text>
        </Space>
      }
      bodyStyle={{ padding: 0 }}
      style={{ borderRadius: 3, border: "1px solid #f0f0f0" }}
    >
      <List
        dataSource={platforms}
        renderItem={(item) => (
          <List.Item style={{ padding: "16px 24px" }}>
            <List.Item.Meta
              avatar={
                <div style={{ display: "flex", alignItems: "center", height: "100%", paddingRight: 8 }}>
                  {getStatusIcon(item.data?.status || "disconnected")}
                </div>
              }
              title={
                <Space style={{ marginBottom: 4 }}>
                  <Text strong style={{ fontSize: 14 }}>{item.name}</Text>
                  <Tag
                    color={getStatusColor(item.data?.status || "disconnected")}
                    bordered={false}
                    style={{ fontSize: 10, lineHeight: "16px" }}
                  >
                    {(item.data?.status || "disconnected").toUpperCase()}
                  </Tag>
                </Space>
              }
              description={
                item.data?.last_sync ? (
                  <Space size={4}>
                    <ClockCircleOutlined style={{ fontSize: 12 }} />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      Synced: {new Date(item.data.last_sync).toLocaleString()}
                    </Text>
                  </Space>
                ) : (
                  <Text type="secondary" style={{ fontSize: 12 }}>
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
