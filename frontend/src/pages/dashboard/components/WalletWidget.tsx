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
        <Space>
          <WalletOutlined />
          <span>Wallet Balance</span>
        </Space>
      }
      extra={
        <Button icon={<ExportOutlined />} type="text">
          Export
        </Button>
      }
      bodyStyle={{ padding: 24 }}
    >
      <Row gutter={[24, 24]}>
        <Col span={24}>
          <Statistic
            title="Total Balance"
            value={totalBalance}
            prefix="Rp"
            precision={0}
            valueStyle={{ color: token.colorInfo, fontWeight: 600 }}
            loading={isShopeeLoading || isTiktokLoading}
          />
        </Col>
        <Col span={12}>
          <Statistic
            title="Shopee"
            value={shopeeData?.total_balance}
            prefix="Rp"
            precision={0}
            valueStyle={{ fontSize: 16 }}
            loading={isShopeeLoading}
          />
        </Col>
        <Col span={12}>
          <Statistic
            title="TikTok"
            value={tiktokData?.total_balance}
            prefix="Rp"
            precision={0}
            valueStyle={{ fontSize: 16 }}
            loading={isTiktokLoading}
          />
        </Col>
      </Row>
      <div style={{ marginTop: 16 }}>
        <Text type="secondary" style={{ fontSize: 12 }}>
          * Balances are estimated based on last sync
        </Text>
      </div>
    </Card>
  );
}
