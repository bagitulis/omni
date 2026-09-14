import { useState } from "react";
import {
  Alert,
  Button,
  Card,
  Empty,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { ReloadOutlined, LinkOutlined } from "@ant-design/icons";
import {
  useExtensions,
  useGeneratePairingCode,
  useUnpairExtension,
} from "@/hooks/useExtensions";
import type { OmniExtension } from "@/api/extensions";

const { Title, Paragraph, Text } = Typography;

/** Seconds a pairing code remains valid, mirroring the backend TTL. */
const PAIRING_CODE_TTL_SECONDS = 300;

function formatTimestamp(value?: string): string {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleString();
}

/** Render a pairing code as plain characters, since it is read aloud or typed. */
function PairingCodePanel({ code }: { code: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      message.success("Pairing code copied");
    } catch {
      // The clipboard is unavailable over plain http and in some browsers; the
      // code is visible on screen either way, so this is not worth an error.
      message.warning("Could not copy automatically — read the code below");
    }
  };

  return (
    <Space direction="vertical" style={{ width: "100%" }} align="center">
      <Text
        strong
        style={{ fontSize: 28, letterSpacing: 4, fontFamily: "monospace" }}
      >
        {code}
      </Text>
      <Space>
        <Button size="small" onClick={copy}>
          {copied ? "Copied" : "Copy"}
        </Button>
        <Text type="secondary">
          Valid for {Math.round(PAIRING_CODE_TTL_SECONDS / 60)} minutes
        </Text>
      </Space>
    </Space>
  );
}

export function ExtensionsPage() {
  const { data: extensions, isLoading, isError, error, refetch } = useExtensions();
  const generateCode = useGeneratePairingCode();
  const unpair = useUnpairExtension();

  const columns: ColumnsType<OmniExtension> = [
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
          loading={unpair.isPending}
          onClick={() => {
            unpair.mutate(record.extension_id, {
              onSuccess: () => message.success("Extension unpaired"),
              onError: (err: Error) =>
                message.error(err.message || "Failed to unpair"),
            });
          }}
        >
          Unpair
        </Button>
      ),
    },
  ];

  const onGenerate = () => {
    generateCode.mutate(undefined, {
      onError: (err: Error) =>
        message.error(err.message || "Failed to generate a pairing code"),
    });
  };

  return (
    <div style={{ padding: 24 }}>
      <Space
        style={{ width: "100%", justifyContent: "space-between" }}
        align="start"
      >
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>
            Installed Extensions
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            Chrome extensions paired to this tenant. Each tenant pairs its own
            browser; an extension can only act within the tenant it was paired
            to.
          </Paragraph>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => void refetch()}>
            Refresh
          </Button>
          <Button
            type="primary"
            icon={<LinkOutlined />}
            loading={generateCode.isPending}
            onClick={onGenerate}
          >
            Pair New Extension
          </Button>
        </Space>
      </Space>

      {generateCode.data && (
        <Card style={{ marginTop: 16 }} title="Pairing code">
          <PairingCodePanel code={generateCode.data.code} />
          <Paragraph type="secondary" style={{ marginTop: 16, marginBottom: 0 }}>
            Enter this code in the Omni extension popup. It works once and
            expires shortly, so generate a new one if it lapses.
          </Paragraph>
        </Card>
      )}

      {isError && (
        <Alert
          style={{ marginTop: 16 }}
          type="error"
          showIcon
          message="Failed to load extensions"
          description={(error as Error)?.message}
        />
      )}

      <Card style={{ marginTop: 16 }}>
        {isLoading ? (
          <Spin />
        ) : (
          <Table
            rowKey="extension_id"
            columns={columns}
            dataSource={extensions ?? []}
            pagination={false}
            locale={{
              emptyText: (
                <Empty description="No extensions paired yet. Use 'Pair New Extension' to start." />
              ),
            }}
          />
        )}
      </Card>
    </div>
  );
}

export default ExtensionsPage;
