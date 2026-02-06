import {
  Table,
  Switch,
  Button,
  Card,
  Popconfirm,
  Tag,
  Space,
  Tooltip,
} from "antd";
import {
  PlayCircleOutlined,
  DeleteOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { AutoFunctionConfig } from "@/types/scriptMonitor";

interface Props {
  configs: AutoFunctionConfig[];
  loading?: boolean;
  onEnable: (name: string) => void;
  onDisable: (name: string) => void;
  onDelete: (name: string) => void;
  onCancelScheduled: (id: number) => void;
}

export function ConfigTab({
  configs,
  loading,
  onEnable,
  onDisable,
  onDelete,
  onCancelScheduled,
}: Props) {
  const columns: ColumnsType<AutoFunctionConfig> = [
    {
      title: "Function Name",
      dataIndex: "name",
      key: "name",
      render: (name: string) => <Tag color="blue">{name}</Tag>,
    },
    {
      title: "Status",
      dataIndex: "enabled",
      key: "enabled",
      render: (enabled: boolean, record) => (
        <Switch
          checked={enabled}
          onChange={(checked) =>
            checked ? onEnable(record.name) : onDisable(record.name)
          }
        />
      ),
    },
    {
      title: "Interval",
      dataIndex: "interval_minutes",
      key: "interval_minutes",
      render: (min: number) => `${min} min`,
    },
    {
      title: "Schedule",
      key: "schedule",
      render: (_, record) => (
        <Space direction="vertical" size="small">
          <span>Start: {record.start_time || "-"}</span>
          <span>End: {record.end_time || "-"}</span>
        </Space>
      ),
    },
    {
      title: "Execution",
      key: "execution",
      render: (_, record) => (
        <Space direction="vertical" size="small">
          <Tooltip title="Last Executed">
            <Tag icon={<PlayCircleOutlined />}>
              {record.last_executed
                ? new Date(record.last_executed).toLocaleString()
                : "-"}
            </Tag>
          </Tooltip>
          <Tooltip title="Next Scheduled">
            <Tag icon={<ClockCircleOutlined />} color="processing">
              {record.next_scheduled_execution
                ? new Date(record.next_scheduled_execution).toLocaleString()
                : "-"}
            </Tag>
          </Tooltip>
        </Space>
      ),
    },
    {
      title: "Actions",
      key: "actions",
      render: (_, record) => (
        <Space>
          {record.next_scheduled_execution && (
            <Popconfirm
              title="Cancel scheduled run?"
              onConfirm={() => onCancelScheduled(record.id)}
            >
              <Button size="small" icon={<ClockCircleOutlined />} danger>
                Skip Next
              </Button>
            </Popconfirm>
          )}
          <Popconfirm
            title="Delete configuration?"
            description="This cannot be undone."
            onConfirm={() => onDelete(record.name)}
          >
            <Button size="small" icon={<DeleteOutlined />} danger type="text" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="Auto-Function Configurations"
      extra={
        <Button type="primary" disabled>
          + Add New
        </Button>
      }
    >
      <Table
        dataSource={configs}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={false}
      />
    </Card>
  );
}
