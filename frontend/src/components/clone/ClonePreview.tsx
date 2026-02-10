import React from "react";
import {
  Descriptions,
  Alert,
  Card,
  Table,
  Typography,
  Space,
  Spin,
  Badge,
  theme,
} from "antd";
import {
  WarningOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from "@ant-design/icons";
import type { ProductData, ConflictResult, Difference } from "@/types/clone";

interface ClonePreviewProps {
  sourceProduct?: ProductData;
  conflictResult?: ConflictResult;
  targetPlatform: string;
  loading?: boolean;
}

const { Text } = Typography;

export const ClonePreview: React.FC<ClonePreviewProps> = ({
  sourceProduct,
  conflictResult,
  targetPlatform,
  loading = false,
}) => {
  const { token } = theme.useToken();

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: "40px" }}>
        <Spin
          size="large"
          tip="Analyzing product and platform requirements..."
        />
      </div>
    );
  }

  if (!sourceProduct) {
    return null;
  }

  const hasConflicts = conflictResult?.has_conflict;
  const adjustments = conflictResult?.adjustments;
  const differences = conflictResult?.differences || [];

  const columns = [
    {
      title: "Field",
      dataIndex: "field",
      key: "field",
      width: "20%",
      render: (text: string) => (
        <Text strong>{text.charAt(0).toUpperCase() + text.slice(1)}</Text>
      ),
    },
    {
      title: "Source Value",
      dataIndex: "source_value",
      key: "source_value",
      width: "40%",
      render: (text: string) => <Text type="secondary">{text}</Text>,
    },
    {
      title: "Target Requirement / Conflict",
      dataIndex: "target_value",
      key: "target_value",
      width: "40%",
      render: (text: string) => <Text type="danger">{text}</Text>,
    },
  ];

  return (
    <div className="clone-preview">
      <Card
        title={
          <Space>
            <Badge status="processing" />
            <span>Source Product Overview</span>
          </Space>
        }
        size="small"
        style={{ marginBottom: 16, borderColor: token.colorBorderSecondary }}
      >
        <Descriptions column={2} size="small">
          <Descriptions.Item label="Name">
            <Text
              ellipsis={{ tooltip: sourceProduct.name }}
              style={{ maxWidth: 300 }}
            >
              {sourceProduct.name}
            </Text>
          </Descriptions.Item>
          <Descriptions.Item label="Price">
            {new Intl.NumberFormat("id-ID", {
              style: "currency",
              currency: "IDR",
            }).format(sourceProduct.price)}
          </Descriptions.Item>
          <Descriptions.Item label="Stock">
            {sourceProduct.stock}
          </Descriptions.Item>
          <Descriptions.Item label="Images">
            {sourceProduct.images.length}
          </Descriptions.Item>
          <Descriptions.Item label="Variants">
            {sourceProduct.variants?.length || 0}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {adjustments && (
        <Space direction="vertical" style={{ width: "100%", marginBottom: 16 }}>
          {adjustments.title_will_truncate && (
            <Alert
              message="Title Truncation Warning"
              description={
                <>
                  Target platform limit is {adjustments.title_limit} chars.
                  Title will be shortened to:{" "}
                  <Text code>{adjustments.adjusted_title}</Text>
                </>
              }
              type="warning"
              showIcon
              icon={<WarningOutlined />}
            />
          )}
          {adjustments.desc_will_truncate && (
            <Alert
              message="Description Truncation Warning"
              description={`Description exceeds target platform limit of ${adjustments.desc_limit} chars and will be truncated.`}
              type="warning"
              showIcon
            />
          )}
        </Space>
      )}

      {hasConflicts ? (
        <Card
          title={
            <Space>
              <ExclamationCircleOutlined
                style={{ color: token.colorWarning }}
              />
              <span>Conflicts Detected</span>
            </Space>
          }
          size="small"
          bodyStyle={{ padding: 0 }}
          style={{ borderColor: token.colorWarningBorder }}
        >
          <Table<Difference>
            dataSource={differences}
            columns={columns}
            pagination={false}
            rowKey="field"
            size="small"
          />
        </Card>
      ) : (
        <Alert
          message="No Conflicts Detected"
          description={`Product is compatible with ${targetPlatform} requirements.`}
          type="success"
          showIcon
          icon={<CheckCircleOutlined />}
        />
      )}
    </div>
  );
};
