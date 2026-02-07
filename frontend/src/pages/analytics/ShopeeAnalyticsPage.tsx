import { useState } from "react";
import { Card, Typography, DatePicker, Space, Empty } from "antd";
import { ShopOutlined } from "@ant-design/icons";
import dayjs from "dayjs";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

// Platform brand color
const SHOPEE_ORANGE = "#ee4d2d";

export const ShopeeAnalyticsPage = () => {
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    dayjs().subtract(14, "day"),
    dayjs(),
  ]);

  return (
    <div style={{ padding: 24, maxWidth: 1600, margin: "0 auto" }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 24,
          flexWrap: "wrap",
          gap: 16,
        }}
      >
        <div>
          <Title level={2} style={{ margin: 0, marginBottom: 4 }}>
            <span style={{ color: SHOPEE_ORANGE }}>Shopee</span> Analytics
          </Title>
          <Text type="secondary">
            Performance insights for your Shopee store
          </Text>
        </div>
        <Space wrap>
          <RangePicker
            value={dateRange}
            onChange={(dates) => dates && setDateRange([dates[0]!, dates[1]!])}
            style={{ width: 260 }}
          />
        </Space>
      </div>

      <Card>
        <Empty
          image={
            <ShopOutlined style={{ fontSize: 64, color: SHOPEE_ORANGE }} />
          }
          description={
            <span>
              Shopee Analytics Coming Soon
              <br />
              <Text type="secondary">
                Connect your Shopee store to view revenue, orders, and product
                performance.
              </Text>
            </span>
          }
        >
          <Text type="secondary" style={{ fontSize: 12 }}>
            Selected period: {dateRange[0].format("MMM DD, YYYY")} -{" "}
            {dateRange[1].format("MMM DD, YYYY")}
          </Text>
        </Empty>
      </Card>
    </div>
  );
};

export default ShopeeAnalyticsPage;
