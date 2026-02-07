import { useState } from "react";
import { Card, Typography, DatePicker, Space, Button, Empty } from "antd";
import { DownloadOutlined, VideoCameraOutlined } from "@ant-design/icons";
import dayjs from "dayjs";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

// Platform brand colors
const TIKTOK_BLACK = "#000000";

const TiktokAnalyticsPage = () => {
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    dayjs().subtract(7, "days"),
    dayjs(),
  ]);

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: 24,
        }}
      >
        <Space direction="vertical" size={0}>
          <Title level={2} style={{ margin: 0 }}>
            TikTok Analytics
          </Title>
          <Text type="secondary">
            Performance metrics and engagement insights
          </Text>
        </Space>
        <Space>
          <RangePicker
            value={dateRange}
            onChange={(dates) => dates && setDateRange([dates[0]!, dates[1]!])}
            style={{ width: 260 }}
          />
          <Button icon={<DownloadOutlined />} disabled>
            Export
          </Button>
        </Space>
      </div>

      <Card>
        <Empty
          image={
            <VideoCameraOutlined
              style={{ fontSize: 64, color: TIKTOK_BLACK }}
            />
          }
          description={
            <span>
              TikTok Analytics Coming Soon
              <br />
              <Text type="secondary">
                Connect your TikTok Shop to view revenue, video performance, and
                engagement metrics.
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

export default TiktokAnalyticsPage;
