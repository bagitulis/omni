import { useMemo, useState } from "react";
import {
  Table,
  Tag,
  Button,
  Card,
  DatePicker,
  Input,
  Select,
  Space,
  Popconfirm,
  Typography,
  theme,
} from "antd";
import { DeleteOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import type { Dayjs } from "dayjs";
import type { JobHistory } from "@/types/scriptMonitor";
import {
  formatDuration,
  formatStatus,
  formatTimestamp,
} from "../utils/historyFormatters";
import { getJobTypeLabel, getUniqueJobTypes } from "../utils/jobUtils";

const { RangePicker } = DatePicker;

interface Props {
  history: JobHistory[];
  loading?: boolean;
  onClear: () => void;
}

export function HistoryTab({ history, loading, onClear }: Props) {
  const { token } = theme.useToken();
  const [searchValue, setSearchValue] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [jobTypeFilter, setJobTypeFilter] = useState("all");
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs] | null>(null);

  const uniqueJobTypes = useMemo(() => getUniqueJobTypes(history), [history]);

  const filteredHistory = useMemo(() => {
    return history.filter((item) => {
      const normalizedSearch = searchValue.trim().toLowerCase();
      const itemJobType = item.job_type || "";
      const matchesSearch =
        !normalizedSearch ||
        item.job_id.toLowerCase().includes(normalizedSearch) ||
        itemJobType.toLowerCase().includes(normalizedSearch);

      const matchesStatus =
        statusFilter === "all" || item.status === statusFilter;
      const matchesJobType =
        jobTypeFilter === "all" || item.job_type === jobTypeFilter;

      const timestamp = item.completed_at || item.created_at;
      const recordDate = timestamp ? new Date(timestamp) : null;
      const hasValidDate =
        recordDate !== null && !Number.isNaN(recordDate.getTime());

      const matchesDateRange =
        !dateRange ||
        (hasValidDate &&
          recordDate >= dateRange[0].startOf("day").toDate() &&
          recordDate <= dateRange[1].endOf("day").toDate());

      return (
        matchesSearch && matchesStatus && matchesJobType && matchesDateRange
      );
    });
  }, [history, searchValue, statusFilter, jobTypeFilter, dateRange]);

  const columns: ColumnsType<JobHistory> = [
    {
      title: "Job ID",
      dataIndex: "job_id",
      key: "job_id",
      width: 100,
      ellipsis: true,
      sorter: (left, right) => left.job_id.localeCompare(right.job_id),
    },
    {
      title: "Type",
      dataIndex: "job_type",
      key: "job_type",
      render: (type?: string) => (
        <Tag color="blue">{getJobTypeLabel(type || "")}</Tag>
      ),
      sorter: (left, right) =>
        (left.job_type || "").localeCompare(right.job_type || ""),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => {
        const formattedStatus = formatStatus(status);
        return <Tag color={formattedStatus.color}>{formattedStatus.label}</Tag>;
      },
      sorter: (left, right) => left.status.localeCompare(right.status),
    },
    {
      title: "Duration",
      dataIndex: "duration_ms",
      key: "duration_ms",
      render: (ms?: number) => formatDuration(ms || 0),
      sorter: (left, right) =>
        (left.duration_ms || 0) - (right.duration_ms || 0),
    },
    {
      title: "Completed At",
      dataIndex: "completed_at",
      key: "completed_at",
      render: (_: string | undefined, record) =>
        formatTimestamp(record.completed_at || record.created_at),
      sorter: (left, right) => {
        const leftTime = new Date(
          left.completed_at || left.created_at,
        ).getTime();
        const rightTime = new Date(
          right.completed_at || right.created_at,
        ).getTime();
        return leftTime - rightTime;
      },
      defaultSortOrder: "descend",
    },
    {
      title: "Error",
      dataIndex: "error_message",
      key: "error_message",
      ellipsis: true,
      render: (msg?: string) => (
        <Typography.Text
          style={{ color: msg ? token.colorError : token.colorText }}
        >
          {msg || "-"}
        </Typography.Text>
      ),
    },
  ];

  return (
    <Card
      title={`Recent History (${filteredHistory.length})`}
      extra={
        <Popconfirm
          title="Clear all history entries?"
          description="This action cannot be undone."
          okText="Clear"
          cancelText="Cancel"
          okButtonProps={{ danger: true }}
          onConfirm={onClear}
        >
          <Button
            danger
            icon={<DeleteOutlined />}
            disabled={history.length === 0}
          >
            Clear History
          </Button>
        </Popconfirm>
      }
    >
      <Space wrap size={12} style={{ marginBottom: 16 }}>
        <Input
          placeholder="Search by Job ID or type"
          allowClear
          value={searchValue}
          onChange={(event) => setSearchValue(event.target.value)}
          style={{ width: 260 }}
        />

        <Select
          value={statusFilter}
          onChange={setStatusFilter}
          style={{ width: 180 }}
          options={[
            { label: "All Statuses", value: "all" },
            { label: "Completed", value: "completed" },
            { label: "Failed", value: "failed" },
            { label: "Cancelled", value: "cancelled" },
          ]}
        />

        <Select
          value={jobTypeFilter}
          onChange={setJobTypeFilter}
          style={{ width: 220 }}
          options={[
            { label: "All Job Types", value: "all" },
            ...uniqueJobTypes.map((type) => ({
              label: getJobTypeLabel(type),
              value: type,
            })),
          ]}
        />

        <RangePicker
          value={dateRange}
          onChange={(value) => setDateRange(value as [Dayjs, Dayjs] | null)}
          allowClear
        />
      </Space>

      <Table
        dataSource={filteredHistory}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          pageSizeOptions: ["10", "20", "50"],
          showTotal: (total) => `${total} items`,
        }}
      />
    </Card>
  );
}
