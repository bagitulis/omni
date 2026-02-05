import { Card, Typography, theme } from "antd";

const { Text } = Typography;
const { useToken } = theme;

interface Props {
  label: string;
  value: number;
  color: string;
}

export const HealthCard = ({ label, value, color }: Props) => {
  const { token } = useToken();

  return (
    <Card
      style={{
        height: "100%",
        borderRadius: token.borderRadius,
        backgroundColor: `${color}15`,
        borderColor: color,
      }}
      styles={{ body: { padding: 16 } }}
    >
      <div style={{ textAlign: "center" }}>
        <Text style={{ fontSize: 12, color: "#666" }}>{label}</Text>
        <div
          style={{
            fontSize: 28,
            fontWeight: 600,
            color: color,
            marginTop: 8,
          }}
        >
          {value}
        </div>
      </div>
    </Card>
  );
};
