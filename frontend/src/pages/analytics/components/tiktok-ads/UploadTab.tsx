import { Card, Col, Row, Upload, Typography, Alert, Spin, theme } from "antd";
import {
  InboxOutlined,
  CheckCircleOutlined,
  FileExcelOutlined,
} from "@ant-design/icons";

const { Dragger } = Upload;
const { Text } = Typography;
const { useToken } = theme;

interface UploadResult {
  totalRows: number;
  period: {
    start: string;
    end: string;
    label: string;
  };
}

interface UploadTabProps {
  uploadProps: import("antd").UploadProps;
  uploading: boolean;
  lastResult: UploadResult | null;
}

export const UploadTab = ({
  uploadProps,
  uploading,
  lastResult,
}: UploadTabProps) => {
  const { token } = useToken();

  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} lg={14}>
        <Card title="Upload TikTok Ads Data" size="small">
          <Spin spinning={uploading} tip="Uploading & processing...">
            <Dragger {...uploadProps} style={{ padding: 16 }}>
              <p className="ant-upload-drag-icon">
                <FileExcelOutlined
                  style={{ color: "#52c41a", fontSize: 48 }}
                />
              </p>
              <p className="ant-upload-text">
                Klik atau drag file Excel (.xlsx) ke sini
              </p>
              <p className="ant-upload-hint">
                Upload file export TikTok Ads — format: &quot;creative data for
                product campaigns YYYY-MM-DD ~ YYYY-MM-DD.xlsx&quot;
              </p>
            </Dragger>
          </Spin>
          <div
            style={{
              marginTop: 16,
              fontSize: 12,
              color: token.colorTextSecondary,
            }}
          >
            <Text type="secondary">
              Sistem akan otomatis mendeteksi periode dari nama file dan mencegah
              upload duplikat.
            </Text>
          </div>
        </Card>
      </Col>
      <Col xs={24} lg={10}>
        <Card title="Upload Status" size="small">
          {lastResult ? (
            <Alert
              type="success"
              icon={<CheckCircleOutlined />}
              showIcon
              message={`Upload berhasil`}
              description={
                <div>
                  <div>
                    Periode: <strong>{lastResult.period?.label}</strong>
                  </div>
                  <div>
                    Total baris: <strong>{lastResult.totalRows}</strong>
                  </div>
                </div>
              }
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
              <p>Belum ada upload terbaru</p>
            </div>
          )}
        </Card>
      </Col>
    </Row>
  );
};
