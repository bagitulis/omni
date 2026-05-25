import { Typography, Space, Tag } from "antd";
import { ShopOutlined, VideoCameraOutlined } from "@ant-design/icons";
import type { ReactNode } from "react";
import type { ReportPlatform } from "@/types/analytics";

const { Title } = Typography;

interface ReportPageHeaderProps {
  title: string;
  platform: ReportPlatform;
  extra?: ReactNode;
}

const platformConfig: Record<
  ReportPlatform,
  { color: string; icon: ReactNode; label: string }
> = {
  shopee: {
    color: "#ee4d2d",
    icon: <ShopOutlined />,
    label: "Shopee",
  },
  tiktok: {
    color: "#161823",
    icon: <VideoCameraOutlined />,
    label: "TikTok",
  },
};

/**
 * Page header component with title and platform badge.
 * Provides an `extra` slot for additional actions (e.g., settings button).
 */
export function ReportPageHeader({
  title,
  platform,
  extra,
}: ReportPageHeaderProps) {
  const config = platformConfig[platform];

  return (
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginBottom: 16,
        flexWrap: "wrap",
        gap: 12,
      }}
    >
      <Space align="center">
        <Title level={4} style={{ margin: 0 }}>
          {title}
        </Title>
        <Tag
          icon={config.icon}
          color={config.color}
          style={{ fontSize: 11, lineHeight: "20px", borderRadius: 3 }}
        >
          {config.label}
        </Tag>
      </Space>
      {extra && <div>{extra}</div>}
    </div>
  );
}
