import { Table, Tag, Button, Card } from "antd";
import { DeleteOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { JobHistory } from "@/types/scriptMonitor";

interface Props {
  history: JobHistory[];
  loading?: boolean;
  onClear: () => void;
}

export function HistoryTab({ history, loading, onClear }: Props) {
  const columns: ColumnsType<JobHistory> = [
    {
      title: "Job ID",
      dataIndex: "job_id",
      key: "job_id",
      width: 100,
      ellipsis: true,
    },
    {
      title: "Type",
      dataIndex: "job_type",
      key: "job_type",
      render: (type: string) => <Tag color="blue">{type}</Tag>,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => {
        let color = "default";
        if (status === "completed") color = "success";
        if (status === "failed") color = "error";
        if (status === "cancelled") color = "warning";
        return <Tag color={color}>{status.toUpperCase()}</Tag>;
      },
    },
    {
      title: "Duration",
      dataIndex: "duration_ms",
      key: "duration_ms",
      render: (ms: number) => (ms ? `${(ms / 1000).toFixed(2)}s` : "-"),
    },
    {
      title: "Completed At",
      dataIndex: "completed_at",
      key: "completed_at",
      render: (date: string) => (date ? new Date(date).toLocaleString() : "-"),
    },
    {
      title: "Error",
      dataIndex: "error_message",
      key: "error_message",
      ellipsis: true,
      render: (msg: string) => (
        <span style={{ color: "red" }}>{msg || "-"}</span>
      ),
    },
  ];

  return (
    <Card
      title={`Recent History (${history.length})`}
      extra={
        <Button
          danger
          icon={<DeleteOutlined />}
          onClick={onClear}
          disabled={history.length === 0}
        >
          Clear History
        </Button>
      }
    >
      <Table
        dataSource={history}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  );
}
