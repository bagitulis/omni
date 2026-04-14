import { Card, Statistic, Typography, GlobalToken, Tag } from "antd";
import { RocketOutlined } from "@ant-design/icons";

const { Text } = Typography;

interface ScaleUpAnalysisProps {
  targetRoas: number;
  maxSafeBudget: number;
  scaleFactor: number;
  optimalLabel: string;
  token: GlobalToken;
  formatCurrency: (value: number) => string;
  formatRoas: (value: number) => string;
}

export const ScaleUpAnalysis = ({
  targetRoas,
  maxSafeBudget,
  scaleFactor,
  optimalLabel,
  token,
  formatCurrency,
  formatRoas,
}: ScaleUpAnalysisProps) => {

  return (
    <Card
      size="small"
      title={
        <span>
          <RocketOutlined style={{ marginRight: 8, color: token.colorSuccess }} />
          Scale-Up Analysis
        </span>
      }
      style={{
        borderRadius: token.borderRadius,
        borderColor: token.colorSuccessBorder,
        background: `${token.colorSuccess}08`,
      }}
    >
      <Statistic
        title={optimalLabel}
        value={maxSafeBudget}
        precision={0}
        formatter={(value) => formatCurrency(Number(value))}
        suffix="/day"
        valueStyle={{ color: token.colorSuccess, fontSize: 20 }}
      />
      <div
        style={{
          marginTop: 12,
          display: "flex",
          gap: 8,
          flexWrap: "wrap",
          alignItems: "center",
        }}
      >
        <Tag color="green">{scaleFactor}x current spend</Tag>
        <Text type="secondary" style={{ fontSize: 12 }}>
          ROAS stays ≥ {formatRoas(targetRoas)} at this budget
        </Text>
      </div>
    </Card>
  );
};
