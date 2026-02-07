import {
  Input,
  Select,
  DatePicker,
  Button,
  Space,
  Card,
  Row,
  Col,
  Flex,
  Typography,
  Switch,
  Tooltip,
} from "antd";
import {
  ReloadOutlined,
  DownloadOutlined,
  SyncOutlined,
} from "@ant-design/icons";
import { Dayjs } from "dayjs";

const { RangePicker } = DatePicker;
const { Text } = Typography;

interface OrderFiltersProps {
  onSearch: (value: string) => void;
  onPlatformChange: (value: string) => void;
  onDateChange: (dates: [Dayjs | null, Dayjs | null] | null) => void;
  onRefresh: () => void;
  onExport: () => void;
  loading?: boolean;
  autoRefresh?: boolean;
  onAutoRefreshChange?: (enabled: boolean) => void;
}

export function OrderFilters({
  onSearch,
  onPlatformChange,
  onDateChange,
  onRefresh,
  onExport,
  loading,
  autoRefresh = true,
  onAutoRefreshChange,
}: OrderFiltersProps) {
  return (
    <Card
      styles={{ body: { padding: 16 } }}
      style={{ marginBottom: 16, borderRadius: 4 }}
    >
      <Row gutter={[16, 16]} align="bottom">
        {/* Platform Select */}
        <Col xs={24} sm={12} md={4}>
          <Flex vertical gap={4}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Platform
            </Text>
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
          </Flex>
        </Col>

        {/* Date Range */}
        <Col xs={24} sm={12} md={6}>
          <Flex vertical gap={4}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Order Date
            </Text>
            <RangePicker style={{ width: "100%" }} onChange={onDateChange} />
          </Flex>
        </Col>

        {/* Search */}
        <Col xs={24} sm={12} md={6}>
          <Flex vertical gap={4}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Search
            </Text>
            <Input.Search
              placeholder="Order ID, Customer Name..."
              onSearch={onSearch}
              allowClear
            />
          </Flex>
        </Col>

        {/* Actions */}
        <Col xs={24} sm={12} md={8}>
          <Flex
            justify="flex-end"
            align="center"
            gap={16}
            style={{ height: "100%" }}
          >
            {/* Auto-Refresh Toggle */}
            <Tooltip title="Auto-refresh every 30 seconds">
              <Flex align="center" gap={6}>
                <SyncOutlined
                  spin={autoRefresh && loading}
                  style={{ color: autoRefresh ? "#52c41a" : "#999" }}
                />
                <Text type="secondary" style={{ fontSize: 11 }}>
                  Auto
                </Text>
                <Switch
                  size="small"
                  checked={autoRefresh}
                  onChange={onAutoRefreshChange}
                />
              </Flex>
            </Tooltip>
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
          </Flex>
        </Col>
      </Row>
    </Card>
  );
}
