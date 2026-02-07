import { Card, Col, Row, Upload, Typography, Table, theme } from "antd";
import { InboxOutlined } from "@ant-design/icons";
import type { UploadProps } from "antd";
import { AdsData, SHOPEE_ORANGE } from "./types";
import { columns } from "./columns";

const { Dragger } = Upload;
const { Text } = Typography;
const { useToken } = theme;

interface UploadTabProps {
  uploadProps: UploadProps;
  uploadedData: AdsData[];
}

export const UploadTab = ({ uploadProps, uploadedData }: UploadTabProps) => {
  const { token } = useToken();
  const hasData = uploadedData.length > 0;

  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} lg={12}>
        <Card title="Upload CSV" size="small">
          <Dragger {...uploadProps} style={{ padding: 16 }}>
            <p className="ant-upload-drag-icon">
              <InboxOutlined style={{ color: SHOPEE_ORANGE, fontSize: 48 }} />
            </p>
            <p className="ant-upload-text">Click or drag CSV file to upload</p>
            <p className="ant-upload-hint">
              Upload Shopee Ads export file to analyze performance
            </p>
          </Dragger>
          <div style={{ marginTop: 16, fontSize: 12, color: "#666" }}>
            <Text type="secondary">
              Expected CSV format: product_id, product_name, cost, revenue,
              clicks, impressions, conversions, date
            </Text>
          </div>
        </Card>
      </Col>
      <Col xs={24} lg={12}>
        <Card title="Upload Preview" size="small">
          {hasData ? (
            <Table
              columns={columns.slice(0, 4)}
              dataSource={uploadedData.slice(0, 5)}
              rowKey="product_id"
              size="small"
              pagination={false}
            />
          ) : (
            <div
              style={{
                textAlign: "center",
                padding: 40,
                color: token.colorTextSecondary,
              }}
            >
              <InboxOutlined style={{ fontSize: 32, marginBottom: 8 }} />
              <p>No data uploaded yet</p>
            </div>
          )}
        </Card>
      </Col>
    </Row>
  );
};
