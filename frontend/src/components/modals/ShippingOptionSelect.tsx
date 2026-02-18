import { Card, Typography, Row, Col, theme } from "antd";
import { CloudDownloadOutlined, FileTextOutlined } from "@ant-design/icons";

const { Title, Text } = Typography;

interface ShippingOptionSelectProps {
  onSelect: (option: "wallet" | "file") => void;
}

export function ShippingOptionSelect({ onSelect }: ShippingOptionSelectProps) {
  const { token } = theme.useToken();

  return (
    <Row gutter={16} style={{ paddingTop: 32, paddingBottom: 32 }}>
      <Col span={12}>
        <Card
          hoverable
          onClick={() => onSelect("wallet")}
          style={{ textAlign: "center", cursor: "pointer" }}
        >
          <CloudDownloadOutlined
            style={{
              fontSize: 36,
              color: token.colorPrimary,
              marginBottom: 16,
            }}
          />
          <Title level={4}>Get from Wallet</Title>
          <Text type="secondary">
            Export shipping fees directly from platform wallet to Google Sheets
          </Text>
        </Card>
      </Col>

      <Col span={12}>
        <Card
          hoverable
          onClick={() => onSelect("file")}
          style={{ textAlign: "center", cursor: "pointer" }}
        >
          <FileTextOutlined
            style={{
              fontSize: 36,
              color: token.colorSuccess,
              marginBottom: 16,
            }}
          />
          <Title level={4}>Process Shipping File</Title>
          <Text type="secondary">
            Load and process downloaded shipping files from local storage
          </Text>
        </Card>
      </Col>
    </Row>
  );
}
