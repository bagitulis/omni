import {
  Card,
  Statistic,
  Button,
  Space,
  Typography,
  Row,
  Col,
  theme,
} from "antd";
import { ExportOutlined, WalletOutlined } from "@ant-design/icons";
import { useWalletData } from "@/hooks/useDashboardWidgets";

const { Text } = Typography;
const { useToken } = theme;

export function WalletWidget() {
  const { token } = useToken();
  const { data: shopeeData, isLoading: isShopeeLoading } =
    useWalletData("shopee");
  const { data: tiktokData, isLoading: isTiktokLoading } =
    useWalletData("tiktok");

  const totalBalance =
    (shopeeData?.total_balance || 0) + (tiktokData?.total_balance || 0);

  return (
    <Card
      title={
        <Space style={{ fontSize: 14 }}>
          <WalletOutlined style={{ color: token.colorPrimary }} />
          <Text strong style={{ fontSize: 14 }}>Wallet Balance</Text>
        </Space>
      }
      extra={
        <Button icon={<ExportOutlined />} type="text" size="small" style={{ fontSize: 12 }}>
          Export
        </Button>
      }
      bodyStyle={{ padding: "16px 24px" }}
      style={{ borderRadius: 3, border: `1px solid ${token.colorBorderSecondary}` }}
    >
      <Row gutter={[24, 24]}>
        <Col span={24}>
          <Statistic
            title={<Text type="secondary" style={{ fontSize: 12 }}>Total Balance</Text>}
            value={totalBalance}
            prefix={<span style={{ fontWeight: 500 }}>Rp</span>}
            precision={0}
            valueStyle={{ color: token.colorPrimary, fontWeight: 700, fontSize: 24 }}
            loading={isShopeeLoading || isTiktokLoading}
          />
        </Col>
        <Col span={12}>
          <Statistic
            title={<Text type="secondary" style={{ fontSize: 12 }}>Shopee</Text>}
            value={shopeeData?.total_balance}
            prefix="Rp"
            precision={0}
            valueStyle={{ fontSize: 16, fontWeight: 500 }}
            loading={isShopeeLoading}
          />
        </Col>
        <Col span={12}>
          <Statistic
            title={<Text type="secondary" style={{ fontSize: 12 }}>TikTok</Text>}
            value={tiktokData?.total_balance}
            prefix="Rp"
            precision={0}
            valueStyle={{ fontSize: 16, fontWeight: 500 }}
            loading={isTiktokLoading}
          />
        </Col>
      </Row>
      <div style={{ marginTop: 24, borderTop: `1px solid ${token.colorBorderSecondary}`, paddingTop: 12 }}>
        <Text type="secondary" style={{ fontSize: 10 }}>
          * Balances are estimated based on last sync
        </Text>
      </div>
    </Card>
  );
}
