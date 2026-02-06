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
  valueStyle,
  onClick,
}: TaskCardProps) {
  return (
    <Card
      size="small"
      loading={loading}
      hoverable={!!onClick}
      onClick={onClick}
      style={{ height: "100%" }}
    >
      <Statistic
        title={<Text type="secondary" style={{ fontWeight: 500 }}>{title}</Text>}
        value={value}
        precision={precision}
        valueStyle={{ fontSize: 24, fontWeight: 600, ...valueStyle }}
        prefix={icon}
        suffix={suffix}
      />
    </Card>
  );
}
