import { useState, useMemo, useCallback } from "react";
import {
  Typography,
  Segmented,
  Select,
  Empty,
  Spin,
  Collapse,
  Table,
  Tag,
  theme,
  Button,
} from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import { useNotifications } from "@/contexts/NotificationContext";
import type { Notification } from "@/api/notifications";
import {
  formatRelativeTime,
  parseNotificationMessage,
  PlatformBreakdown,
} from "@/components/layout/NotificationHelpers";

const { Title, Text } = Typography;

type StatusFilter = "all" | "unread" | "read";
type CategoryFilter = "all" | "sync" | "order" | "product" | "inventory" | "system";

const typeIcons: Record<string, React.ReactNode> = {
  success: <CheckCircleOutlined style={{ color: "#52c41a" }} />,
  error: <CloseCircleOutlined style={{ color: "#ff4d4f" }} />,
  warning: <ExclamationCircleOutlined style={{ color: "#faad14" }} />,
  info: <InfoCircleOutlined style={{ color: "#1677ff" }} />,
};

export default function NotificationsPage() {
  const { token } = theme.useToken();
  const { notifications, loading, markAsRead, fetchNotifications } =
    useNotifications();

  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>("all");

  const filtered = useMemo(() => {
    let items = notifications;
    if (statusFilter === "unread") items = items.filter((n) => !n.read);
    if (statusFilter === "read") items = items.filter((n) => n.read);
    if (categoryFilter !== "all")
      items = items.filter((n) => n.category === categoryFilter);
    return items;
  }, [notifications, statusFilter, categoryFilter]);

  const handleExpand = useCallback(
    (keys: string | string[]) => {
      const expandedKeys = Array.isArray(keys) ? keys : [keys];
      // Mark as read when expanded
      expandedKeys.forEach((key) => {
        const notif = notifications.find((n) => String(n.id) === key);
        if (notif && !notif.read) {
          markAsRead(notif.id);
        }
      });
    },
    [notifications, markAsRead],
  );

  const collapseItems = filtered.map((notif) => {
    const parsed = parseNotificationMessage(notif.message);
    return {
      key: String(notif.id),
      label: (
        <div style={{ display: "flex", alignItems: "center", gap: 10, width: "100%" }}>
          <span style={{ fontSize: 18 }}>{typeIcons[notif.type] || typeIcons.info}</span>
          <div style={{ flex: 1, minWidth: 0 }}>
            <Text strong={!notif.read} style={{ display: "block" }}>
              {notif.title}
            </Text>
            {parsed.stats && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {parsed.stats.total !== undefined && `Total: ${parsed.stats.total}`}
                {parsed.stats.processed !== undefined && ` · Done: ${parsed.stats.processed}`}
                {parsed.stats.failed !== undefined && parsed.stats.failed > 0 && ` · Failed: ${parsed.stats.failed}`}
              </Text>
            )}
          </div>
          <Text type="secondary" style={{ fontSize: 12, flexShrink: 0 }}>
            {formatRelativeTime(notif.created_at)}
          </Text>
          {!notif.read && (
            <div
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                background: token.colorPrimary,
                flexShrink: 0,
              }}
            />
          )}
        </div>
      ),
      children: (
        <NotificationDetail notif={notif} parsed={parsed} />
      ),
    };
  });

  return (
    <div style={{ padding: "24px", maxWidth: 900, margin: "0 auto" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
        <Title level={3} style={{ margin: 0 }}>
          Notifications
        </Title>
        <Button icon={<ReloadOutlined />} onClick={fetchNotifications} loading={loading}>
          Refresh
        </Button>
      </div>

      {/* Filters */}
      <div style={{ display: "flex", gap: 12, marginBottom: 20, flexWrap: "wrap" }}>
        <Segmented
          value={statusFilter}
          onChange={(v) => setStatusFilter(v as StatusFilter)}
          options={[
            { label: "All", value: "all" },
            { label: "Unread", value: "unread" },
            { label: "Read", value: "read" },
          ]}
        />
        <Select
          value={categoryFilter}
          onChange={setCategoryFilter}
          style={{ width: 140 }}
          options={[
            { label: "All Categories", value: "all" },
            { label: "Sync", value: "sync" },
            { label: "Order", value: "order" },
            { label: "Product", value: "product" },
            { label: "Inventory", value: "inventory" },
            { label: "System", value: "system" },
          ]}
        />
      </div>

      {/* List */}
      {loading && filtered.length === 0 ? (
        <Spin style={{ display: "block", margin: "60px auto" }} />
      ) : filtered.length === 0 ? (
        <Empty description="No notifications" style={{ marginTop: 60 }} />
      ) : (
        <Collapse
          accordion={false}
          onChange={handleExpand}
          items={collapseItems}
          style={{ background: token.colorBgContainer }}
        />
      )}
    </div>
  );
}

/** Expanded detail view for a single notification */
function NotificationDetail({
  notif,
  parsed,
}: {
  notif: Notification;
  parsed: ReturnType<typeof parseNotificationMessage>;
}) {
  const { token } = theme.useToken();

  return (
    <div style={{ padding: "8px 0" }}>
      {/* Platform breakdown */}
      {parsed.platforms && Object.keys(parsed.platforms).length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <Text strong style={{ display: "block", marginBottom: 6 }}>
            Per Platform
          </Text>
          <PlatformBreakdown platforms={parsed.platforms} />
        </div>
      )}

      {/* Failed items */}
      {parsed.failedItems && parsed.failedItems.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <Text strong style={{ display: "block", marginBottom: 6, color: token.colorError }}>
            Failed Items ({parsed.failedItems.length})
          </Text>
          <Table
            size="small"
            pagination={parsed.failedItems.length > 10 ? { pageSize: 10 } : false}
            dataSource={parsed.failedItems.map((item, i) => ({ ...item, key: i }))}
            columns={[
              { title: "SKU", dataIndex: "sku", key: "sku", width: 120 },
              { title: "Platform", dataIndex: "platform", key: "platform", width: 100, render: (v: string) => v ? <Tag>{v}</Tag> : "-" },
              { title: "Error", dataIndex: "error", key: "error" },
            ]}
          />
        </div>
      )}

      {/* Raw message if no structured data */}
      {!parsed.platforms && !parsed.failedItems && parsed.summary && (
        <Text type="secondary">{parsed.summary}</Text>
      )}

      {/* Metadata */}
      <div style={{ marginTop: 8 }}>
        <Text type="secondary" style={{ fontSize: 11 }}>
          Category: {notif.category} · Created: {new Date(notif.created_at).toLocaleString()}
        </Text>
      </div>
    </div>
  );
}
