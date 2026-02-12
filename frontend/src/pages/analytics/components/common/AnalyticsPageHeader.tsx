import React from "react";
import { Button, Typography, theme } from "antd";
import { SettingOutlined } from "@ant-design/icons";
import type { ReactNode } from "react";

const { Title, Text } = Typography;

interface AnalyticsPageHeaderProps {
  title: ReactNode;
  subtitle: string;
  onSettingsClick: () => void;
}

const AnalyticsPageHeader = ({
  title,
  subtitle,
  onSettingsClick,
}: AnalyticsPageHeaderProps) => {
  const { token } = theme.useToken();

  return (
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginBottom: 24,
      }}
    >
      <div>
        <Title level={2} style={{ margin: 0 }}>
          {title}
        </Title>
        <Text type="secondary">{subtitle}</Text>
      </div>
      <Button
        icon={<SettingOutlined />}
        onClick={onSettingsClick}
        style={{ borderRadius: token.borderRadius }}
      >
        Settings
      </Button>
    </div>
  );
};

export default React.memo(AnalyticsPageHeader);
export { AnalyticsPageHeader };
