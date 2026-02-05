import { Typography } from "antd";

const { Text } = Typography;

export function StatCard({
  label,
  value,
  color,
}: {
  label: string;
  value: number;
  color: string;
}) {
  return (
    <div>
      <Text strong style={{ fontSize: 12 }}>
        {label}
      </Text>
      <div style={{ fontSize: 24, fontWeight: 600, color, marginTop: 8 }}>
        {value}
      </div>
    </div>
  );
}
