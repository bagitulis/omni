import { Card, Table, Tag, Typography, theme } from "antd";
import type { TableProps } from "antd";
import type { RouteConfig } from "@/types/routeConfig";

const { Text } = Typography;

interface RouteQueueSectionProps {
  routes: RouteConfig[];
}

interface QueueRow {
  key: number;
  route_path: string;
  queue_max_size: number;
  queue_priority: number;
}

export function RouteQueueSection({ routes }: RouteQueueSectionProps) {
  const { token } = theme.useToken();

  const queue_routes = routes.filter((route) => route.queue_enabled);
  const data_source: QueueRow[] = queue_routes.map((route) => ({
    key: route.id,
    route_path: route.route_path,
    queue_max_size: route.queue_max_size ?? 0,
    queue_priority: route.queue_priority ?? 0,
  }));

  const columns: TableProps<QueueRow>["columns"] = [
    {
      title: "Route",
      dataIndex: "route_path",
      key: "route_path",
      render: (value: string) => <Text code>{value}</Text>,
    },
    {
      title: "Queue Max Size",
      dataIndex: "queue_max_size",
      key: "queue_max_size",
      width: 140,
      sorter: (a, b) => a.queue_max_size - b.queue_max_size,
    },
    {
      title: "Queue Priority",
      dataIndex: "queue_priority",
      key: "queue_priority",
      width: 140,
      sorter: (a, b) => a.queue_priority - b.queue_priority,
      render: (value: number) => {
        const color =
          value >= 8
            ? token.colorError
            : value >= 5
              ? token.colorWarning
              : token.colorSuccess;
        return <Tag color={color}>{value}</Tag>;
      },
    },
  ];

  return (
    <Card
      size="small"
      title="Queue Overview"
      style={{ borderRadius: token.borderRadius }}
    >
      <Table<QueueRow>
        rowKey="key"
        size="small"
        columns={columns}
        dataSource={data_source}
        pagination={false}
        locale={{ emptyText: "No queue-enabled routes found" }}
      />
    </Card>
  );
}
