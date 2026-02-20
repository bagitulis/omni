import { Card, Statistic, Typography } from "antd";
import React, { ReactNode } from "react";

const { Text } = Typography;

interface TaskCardProps {
  title: string;
  value: number | string;
  icon?: ReactNode;
  loading?: boolean;
  precision?: number;
  suffix?: ReactNode;
  subtitle?: string;
  valueStyle?: React.CSSProperties;
  onClick?: () => void;
}

export function TaskCard({
  title,
  value,
  icon,
  loading,
  precision,
  suffix,
  subtitle,
  valueStyle,
  onClick,
}: TaskCardProps) {
  return (
    <Card
      size="small"
      loading={loading}
      hoverable={!!onClick}
      onClick={onClick}
      style={{
        height: "100%",
        borderRadius: 3,
        border: "1px solid #f0f0f0",
        transition: "transform 0.2s, box-shadow 0.2s",
      }}
      bodyStyle={{ padding: "16px 20px" }}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        <Statistic
          title={
            <Text type="secondary" style={{ fontSize: 12, fontWeight: 500 }}>
              {title}
            </Text>
          }
          value={value}
          precision={precision}
          valueStyle={{ fontSize: 24, fontWeight: 600, ...valueStyle }}
          prefix={icon}
          suffix={suffix}
        />
        {subtitle && (
          <Text type="secondary" style={{ fontSize: 10, marginTop: -4 }}>
            {subtitle}
          </Text>
        )}
      </div>
    </Card>
  );
}
