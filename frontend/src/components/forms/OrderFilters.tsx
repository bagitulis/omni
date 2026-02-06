import { Input, Select, DatePicker, Button, Space, Card, Row, Col } from "antd";
import { ReloadOutlined, DownloadOutlined } from "@ant-design/icons";
import { Dayjs } from "dayjs";

const { RangePicker } = DatePicker;

interface OrderFiltersProps {
  onSearch: (value: string) => void;
  onPlatformChange: (value: string) => void;
  onDateChange: (dates: [Dayjs | null, Dayjs | null] | null) => void;
  onRefresh: () => void;
  onExport: () => void;
  loading?: boolean;
}

export function OrderFilters({
  onSearch,
  onPlatformChange,
  onDateChange,
  onRefresh,
  onExport,
  loading,
}: OrderFiltersProps) {
  return (
    <Card
      styles={{ body: { padding: "16px" } }}
      className="mb-4 rounded-sm border-slate-200 shadow-sm"
    >
      <Row gutter={[16, 16]} align="middle">
        {/* Platform Select */}
        <Col xs={24} sm={12} md={4}>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-slate-500 font-medium">Platform</span>
            <Select
              defaultValue="all"
              style={{ width: "100%" }}
              onChange={onPlatformChange}
              options={[
                { value: "all", label: "All Platforms" },
                { value: "shopee", label: "Shopee" },
                { value: "tiktok", label: "TikTok" },
                { value: "lazada", label: "Lazada" },
              ]}
            />
          </div>
        </Col>

        {/* Date Range */}
        <Col xs={24} sm={12} md={6}>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-slate-500 font-medium">
              Order Date
            </span>
            <RangePicker style={{ width: "100%" }} onChange={onDateChange} />
          </div>
        </Col>

        {/* Search */}
        <Col xs={24} sm={12} md={6}>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-slate-500 font-medium">Search</span>
            <Input.Search
              placeholder="Order ID, Customer Name..."
              onSearch={onSearch}
              allowClear
            />
          </div>
        </Col>

        {/* Actions */}
        <Col
          xs={24}
          sm={12}
          md={8}
          className="flex justify-end items-end h-full mt-auto"
        >
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={onRefresh}
              loading={loading}
            >
              Refresh
            </Button>
            <Button icon={<DownloadOutlined />} onClick={onExport}>
              Export
            </Button>
          </Space>
        </Col>
      </Row>
    </Card>
  );
}
