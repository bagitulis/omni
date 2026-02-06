import { Table, Tag, Card } from "antd";
import type { ColumnsType } from "antd/es/table";
import { Job } from "@/types/scriptMonitor";

interface Props {
  queue: Job[];
  loading?: boolean;
}

export function QueueTab({ queue, loading }: Props) {
  const columns: ColumnsType<Job> = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      width: 100,
      ellipsis: true,
    },
    {
      title: "Type",
      dataIndex: "type",
      key: "type",
      render: (type: string) => <Tag color="blue">{type}</Tag>,
    },
    {
      title: "Priority",
      dataIndex: "priority",
      key: "priority",
      render: (priority: string) => (
        <Tag color={priority === "high" ? "red" : "orange"}>
          {priority.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: "Created At",
      dataIndex: "created_at",
      key: "created_at",
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: "Data",
      dataIndex: "data",
      key: "data",
      ellipsis: true,
      render: (data: any) => JSON.stringify(data),
    },
  ];

  return (
    <Card title={`Pending Queue (${queue.length})`}>
      <Table
        dataSource={queue}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  );
}
