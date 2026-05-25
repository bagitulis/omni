import { Card, Statistic, Typography, theme } from "antd";
import type { ReactNode, CSSProperties } from "react";

const { Text } = Typography;

interface StatCardProps {
  title: string;
  value: number | string;
  icon?: ReactNode;
  loading?: boolean;
  prefix?: ReactNode;
  suffix?: ReactNode;
  valueStyle?: CSSProperties;
  precision?: number;
  color?: string;
}

/**
 * Stat display card wrapping Ant Design Statistic inside a Card.
 * Used across analytics pages for summary metrics.
 */
export function StatCard({
  title,
  value,
  icon,
  loading,
  prefix,
  suffix,
  valueStyle,
  precision,
  color,
}: StatCardProps) {
  const { token } = theme.useToken();

  const mergedValueStyle: CSSProperties = {
    fontSize: 24,
    fontWeight: 600,
    ...(color ? { color } : {}),
    ...valueStyle,
  };

  return (
    <Card
      size="small"
      loading={loading}
      style={{
        height: "100%",
        borderRadius: 3,
        border: `1px solid ${token.colorBorderSecondary}`,
      }}
      styles={{ body: { padding: "16px 20px" } }}
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
          valueStyle={mergedValueStyle}
          prefix={icon || prefix}
          suffix={suffix}
        />
      </div>
    </Card>
  );
}
