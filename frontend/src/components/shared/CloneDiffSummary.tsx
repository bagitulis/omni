import type { FC } from "react";
import { Card, Space, Tag, Alert, theme } from "antd";
import {
  WarningOutlined,
  ArrowRightOutlined,
  CheckCircleOutlined,
} from "@ant-design/icons";

export interface CloneDifference {
  field: string;
  source_value: string;
  target_value: string;
  reason: string; // e.g., "Title truncated to 120 chars for Shopee"
}

interface DiffSummarySectionProps {
  differences: CloneDifference[];
}

// Difference highlight color (from Vue — yellow background)
const DIFF_HIGHLIGHT = "#fef3c7"; // Amber-50

export const DiffSummarySection: FC<DiffSummarySectionProps> = ({
  differences,
}) => {
  const { token } = theme.useToken();

  return (
    <>
      {/* Differences summary */}
      {differences.length > 0 && (
        <Card
          size="small"
          title={
            <Space>
              <WarningOutlined style={{ color: "#faad14" }} />
              <span>
                {differences.length} difference
                {differences.length > 1 ? "s" : ""} detected
              </span>
            </Space>
          }
          style={{ marginTop: 16 }}
          styles={{ body: { padding: 12 } }}
        >
          {differences.map((diff, idx) => (
            <div
              key={`${diff.field}-${diff.target_value}`}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                padding: "4px 0",
                borderBottom:
                  idx < differences.length - 1
                    ? `1px solid ${token.colorBorderSecondary}`
                    : "none",
              }}
            >
              <Tag>{diff.field}</Tag>
              <span style={{ color: token.colorTextSecondary, fontSize: 12 }}>
                {diff.source_value}
              </span>
              <ArrowRightOutlined
                style={{ fontSize: 10, color: token.colorTextQuaternary }}
              />
              <span
                style={{
                  backgroundColor: DIFF_HIGHLIGHT,
                  padding: "2px 4px",
                  borderRadius: 3,
                  fontSize: 12,
                }}
              >
                {diff.target_value}
              </span>
              <span
                style={{
                  color: "#d48806",
                  fontSize: 11,
                  marginLeft: "auto",
                }}
              >
                {diff.reason}
              </span>
            </div>
          ))}
        </Card>
      )}

      {/* No differences */}
      {differences.length === 0 && (
        <Alert
          message="No differences detected — clone will be identical to source."
          type="success"
          showIcon
          icon={<CheckCircleOutlined />}
          style={{ marginTop: 16 }}
        />
      )}
    </>
  );
};
