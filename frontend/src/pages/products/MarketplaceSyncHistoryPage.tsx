import {
  ArrowLeftOutlined,
  ReloadOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Button,
  Card,
  Col,
  DatePicker,
  Input,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Typography,
  theme,
} from "antd";
import type { Dayjs } from "dayjs";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useMarketplaceSyncHistory } from "@/hooks/useMarketplaceSyncHistory";
import type {
  MarketplaceSyncHistoryFilter,
  Platform,
  SyncOperation,
  SyncResultStatus,
} from "@/types/shared";
import {
  createMarketplaceSyncHistoryColumns,
  isSyncHistoryRowExpandable,
  renderSyncHistoryExpandedRow,
} from "./utils/marketplaceSyncHistoryPageUtils";

export default function MarketplaceSyncHistoryPage() {
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const columns = useMemo(() => createMarketplaceSyncHistoryColumns(), []);

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [platform, setPlatform] = useState<Platform | undefined>(undefined);
  const [operation, setOperation] = useState<SyncOperation | undefined>(
    undefined,
  );
  const [status, setStatus] = useState<SyncResultStatus | undefined>(undefined);
  const [skuSearch, setSkuSearch] = useState("");
  const [dateRange, setDateRange] = useState<
    [Dayjs | null, Dayjs | null] | null
  >(null);

  const filter = useMemo<MarketplaceSyncHistoryFilter>(
    () => ({
      page,
      page_size: pageSize,
      platform,
      operation,
      status,
      sku_search: skuSearch.trim() || undefined,
      date_from: dateRange?.[0]?.format("YYYY-MM-DD"),
      date_to: dateRange?.[1]?.format("YYYY-MM-DD"),
    }),
    [dateRange, operation, page, pageSize, platform, skuSearch, status],
  );

  const { data, isLoading, isError, error, refetch } =
    useMarketplaceSyncHistory(filter);
  const entries = useMemo(() => data?.entries ?? [], [data?.entries]);

  const summary = useMemo(() => {
    const successCount = entries.filter(
      (entry) => entry.status === "success",
    ).length;
    const failedCount = entries.filter(
      (entry) => entry.status === "failed",
    ).length;

    return {
      total: entries.length,
      failed: failedCount,
      success_rate:
        entries.length === 0 ? 0 : (successCount / entries.length) * 100,
    };
  }, [entries]);

  return (
    <div style={{ padding: 24 }}>
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <Space
          style={{ justifyContent: "space-between", width: "100%" }}
          align="center"
        >
          <Space align="center">
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate("/products")}
            >
              Back to Products
            </Button>
            <Typography.Title level={2} style={{ margin: 0 }}>
              Marketplace Sync History
            </Typography.Title>
          </Space>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            Refresh
          </Button>
        </Space>

        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8} md={8} lg={6}>
            <Card size="small">
              <Statistic title="Total Operations" value={summary.total} />
            </Card>
          </Col>
          <Col xs={24} sm={8} md={8} lg={6}>
            <Card size="small">
              <Statistic
                title="Success Rate"
                value={summary.success_rate}
                precision={1}
                suffix="%"
              />
            </Card>
          </Col>
          <Col xs={24} sm={8} md={8} lg={6}>
            <Card size="small">
              <Statistic
                title="Failed Operations"
                value={summary.failed}
                valueStyle={{
                  color: summary.failed > 0 ? token.colorError : undefined,
                }}
              />
            </Card>
          </Col>
        </Row>

        <Card>
          <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <Row gutter={[12, 12]}>
              <Col xs={24} sm={12} md={8} lg={6}>
                <DatePicker.RangePicker
                  style={{ width: "100%" }}
                  value={dateRange}
                  onChange={(nextRange) => {
                    setDateRange(nextRange);
                    setPage(1);
                  }}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={4}>
                <Select
                  allowClear
                  placeholder="Platform"
                  style={{ width: "100%" }}
                  value={platform}
                  onChange={(value) => {
                    setPlatform(value);
                    setPage(1);
                  }}
                  options={[
                    { label: "Shopee", value: "shopee" },
                    { label: "TikTok", value: "tiktok" },
                    { label: "Lazada", value: "lazada" },
                  ]}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={5}>
                <Select
                  allowClear
                  placeholder="Operation"
                  style={{ width: "100%" }}
                  value={operation}
                  onChange={(value) => {
                    setOperation(value);
                    setPage(1);
                  }}
                  options={[
                    { label: "Stock Update", value: "stock_update" },
                    { label: "Price Update", value: "price_update" },
                    { label: "Wholesale Update", value: "wholesale_update" },
                    { label: "MPQ Update", value: "mpq_update" },
                    { label: "Clone", value: "clone" },
                  ]}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={4}>
                <Select
                  allowClear
                  placeholder="Status"
                  style={{ width: "100%" }}
                  value={status}
                  onChange={(value) => {
                    setStatus(value);
                    setPage(1);
                  }}
                  options={[
                    { label: "Success", value: "success" },
                    { label: "Partial", value: "partial" },
                    { label: "Failed", value: "failed" },
                  ]}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={5}>
                <Input
                  allowClear
                  prefix={<SearchOutlined />}
                  placeholder="Search SKU"
                  value={skuSearch}
                  onChange={(event) => {
                    setSkuSearch(event.target.value);
                    setPage(1);
                  }}
                />
              </Col>
            </Row>

            {isError && (
              <Alert
                type="error"
                showIcon
                message="Failed to load marketplace sync history"
                description={
                  error instanceof Error ? error.message : "Unknown error"
                }
              />
            )}

            <Table
              rowKey="id"
              columns={columns}
              dataSource={entries}
              loading={isLoading}
              scroll={{ x: 1100 }}
              expandable={{
                rowExpandable: isSyncHistoryRowExpandable,
                expandedRowRender: renderSyncHistoryExpandedRow,
              }}
              pagination={{
                current: page,
                pageSize,
                total: data?.total ?? 0,
                showSizeChanger: true,
                onChange: (nextPage, nextPageSize) => {
                  setPage(nextPage);
                  setPageSize(nextPageSize);
                },
              }}
            />
          </Space>
        </Card>
      </Space>
    </div>
  );
}
