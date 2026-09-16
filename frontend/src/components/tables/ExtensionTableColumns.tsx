import { Button, Space, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { OmniExtension } from "@/api/extensions";

const { Text } = Typography;

export interface GetExtensionTableColumnsProps {
  onUnpair: (extensionId: string) => void;
  unpairing: boolean;
}

function formatTimestamp(value?: string): string {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleString();
}

export function getExtensionTableColumns({
  onUnpair,
  unpairing,
}: GetExtensionTableColumnsProps): ColumnsType<OmniExtension> {
  return [
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 130,
      render: (status: string) => (
        <Tag color={status === "connected" ? "green" : "default"}>
          {status === "connected" ? "Connected" : "Offline"}
        </Tag>
      ),
    },
    {
      title: "Extension",
      dataIndex: "extension_id",
      key: "extension_id",
      render: (id: string) => (
        <Space direction="vertical" size={0}>
          <Text copyable={{ text: id }} style={{ fontFamily: "monospace" }}>
            {id.slice(0, 16)}…
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {id}
          </Text>
        </Space>
      ),
    },
    {
      title: "Capabilities",
      dataIndex: "capabilities",
      key: "capabilities",
      render: (caps: string[]) =>
        caps.length > 0 ? (
          <Space wrap size={4}>
            {caps.map((c) => (
              <Tag key={c}>{c}</Tag>
            ))}
          </Space>
        ) : (
          <Text type="secondary">none</Text>
        ),
    },
    {
      title: "Last seen",
      dataIndex: "last_seen",
      key: "last_seen",
      width: 180,
      render: formatTimestamp,
    },
    {
      title: "Paired",
      dataIndex: "paired_at",
      key: "paired_at",
      width: 180,
      render: formatTimestamp,
    },
    {
      title: "",
      key: "actions",
      width: 110,
      render: (_, record) => (
        <Button
          danger
          size="small"
          loading={unpairing}
          onClick={() => onUnpair(record.extension_id)}
        >
          Unpair
        </Button>
      ),
    },
  ];
}
