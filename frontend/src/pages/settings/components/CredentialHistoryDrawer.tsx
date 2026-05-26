import { Drawer, Table } from "antd";
import type { CredentialAuditEvent, CredentialPlatformSummary } from "@/api/credentials";

interface CredentialHistoryDrawerProps {
  open: boolean;
  platform: CredentialPlatformSummary | null;
  events: CredentialAuditEvent[];
  loading: boolean;
  platformNames: Record<string, string>;
  onClose: () => void;
}

export function CredentialHistoryDrawer({
  open,
  platform,
  events,
  loading,
  platformNames,
  onClose,
}: CredentialHistoryDrawerProps) {
  const title = platform
    ? `Credential History - ${platformNames[platform.platform] || platform.platform}`
    : "Credential History";

  return (
    <Drawer open={open} title={title} width={560} onClose={onClose}>
      <Table
        rowKey={(record) => `${record.event_type}-${record.created_at}`}
        size="small"
        loading={loading}
        dataSource={events}
        pagination={false}
        columns={[
          {
            title: "Time",
            dataIndex: "created_at",
            render: (value: string) => new Date(value).toLocaleString(),
          },
          { title: "Event", dataIndex: "event_type" },
          {
            title: "Code",
            dataIndex: "code",
            render: (value: string | undefined) => value || "-",
          },
        ]}
      />
    </Drawer>
  );
}
