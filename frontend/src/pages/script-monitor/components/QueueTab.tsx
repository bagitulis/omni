import { Button, Card, Popconfirm, Space, Table, Tag } from "antd";
import { PauseCircleOutlined, StopOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import type { Job } from "@/types/scriptMonitor";

interface Props {
  queue: Job[];
  loading?: boolean;
  onCancel: (jobId: string) => void;
  onForceCancel: (jobId: string) => void;
}

export function QueueTab({ queue, loading, onCancel, onForceCancel }: Props) {
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
      render: (data: unknown) => JSON.stringify(data),
    },
    {
      title: "Actions",
      key: "actions",
      render: (_, record) => (
        <Space>
          <Popconfirm
            title="Cancel this job?"
            onConfirm={() => onCancel(record.id)}
          >
            <Button size="small" icon={<PauseCircleOutlined />}>
              Cancel
            </Button>
          </Popconfirm>

          <Popconfirm
            title="Force cancel this stuck job?"
            okButtonProps={{ danger: true }}
            onConfirm={() => onForceCancel(record.id)}
          >
            <Button size="small" danger icon={<StopOutlined />}>
              Force Cancel
            </Button>
          </Popconfirm>
        </Space>
      ),
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
