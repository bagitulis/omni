import { CheckOutlined, CloseOutlined, DeleteOutlined, ExportOutlined } from "@ant-design/icons";
import { Button, Space, Table, Tag, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import type { Notification } from "@/api/notifications";
import { useNotifications } from "@/contexts/NotificationContext";
import { isSafeActionURL } from "@/lib/notificationSecurity";
import {
  NotificationMetadata,
  NotificationStats,
  PlatformBreakdown,
  parseNotificationMessage,
} from "@/components/layout/NotificationHelpers";
import { severityLabel } from "./severity";

const { Title, Text } = Typography;

interface NotificationDetailPaneProps {
  notif: Notification;
  onBack?: () => void;
}

/**
 * Right-side detail view (or full-screen on <lg). Includes:
 *   - Header with severity + close/back.
 *   - Structured message (platform breakdown, failed items table).
 *   - Raw fallback if not structured.
 *   - Metadata + created_at + dedup badge.
 *   - Sticky action bar: mark read, delete, snooze, open action_url.
 */
export function NotificationDetailPane({ notif, onBack }: NotificationDetailPaneProps) {
  const parsed = parseNotificationMessage(notif.message);
  const { markAsRead, deleteNotification, snooze } = useNotifications();
  const navigate = useNavigate();

  const handleOpenAction = () => {
    if (notif.action_url && isSafeActionURL(notif.action_url)) navigate(notif.action_url);
  };

  const handleSnooze1h = () => {
    const until = new Date(Date.now() + 60 * 60 * 1000);
    void snooze(notif.id, until);
    onBack?.();
  };

  return (
    <section className="notification-detail-pane" aria-label={`Details for ${notif.title}`}>
      <header className="notification-detail-pane__header">
        <div>
          <Tag color="default">{severityLabel(notif.severity)}</Tag>
          <Title level={4} className="notification-detail-pane__title">
            {notif.title}
          </Title>
          <Text type="secondary" className="notification-detail-pane__timestamp">
            {new Date(notif.created_at).toLocaleString()}
            {notif.dedup_count > 1 ? ` · ×${notif.dedup_count}` : ""}
          </Text>
        </div>
        {onBack && (
          <Button type="text" icon={<CloseOutlined />} aria-label="Close detail" onClick={onBack} />
        )}
      </header>

      <div className="notification-detail-pane__body">
        {parsed.platforms && Object.keys(parsed.platforms).length > 0 && (
          <section className="notification-detail-pane__section">
            <Text strong className="notification-detail-pane__section-title">
              Per Platform
            </Text>
            <PlatformBreakdown platforms={parsed.platforms} />
          </section>
        )}

        {parsed.failedItems && parsed.failedItems.length > 0 && (
          <section className="notification-detail-pane__section">
            <Text strong type="danger" className="notification-detail-pane__section-title">
              Failed Items ({parsed.failedItems.length})
            </Text>
            <Table
              size="small"
              pagination={parsed.failedItems.length > 10 ? { pageSize: 10 } : false}
              dataSource={parsed.failedItems.map((item, i) => ({ ...item, key: i }))}
              className="notification-detail-pane__table"
              scroll={{ x: 520 }}
              columns={[
                { title: "SKU", dataIndex: "sku", key: "sku", width: 120 },
                {
                  title: "Platform",
                  dataIndex: "platform",
                  key: "platform",
                  width: 100,
                  render: (v: string) => (v ? <Tag>{v}</Tag> : "-"),
                },
                { title: "Error", dataIndex: "error", key: "error" },
              ]}
            />
          </section>
        )}

        {!parsed.platforms && !parsed.failedItems && parsed.summary && (
          <Text type="secondary" className="notification-detail-pane__summary">
            {parsed.summary}
          </Text>
        )}

        {parsed.stats && (
          <NotificationStats stats={parsed.stats} failed={notif.type === "error"} platforms={parsed.platforms} />
        )}

        <NotificationMetadata metadata={notif.metadata} />

        <footer className="notification-detail-pane__meta">
          <Text type="secondary">
            Category: {notif.category} · Source: {notif.source ?? "system"}
          </Text>
        </footer>
      </div>

      <div className="notification-detail-pane__actions" role="toolbar" aria-label="Notification actions">
        <Space wrap>
          {!notif.read && (
            <Button icon={<CheckOutlined />} onClick={() => markAsRead(notif.id)}>
              Mark read
            </Button>
          )}
          <Button icon={<ExportOutlined />} onClick={handleOpenAction} disabled={!notif.action_url || !isSafeActionURL(notif.action_url)}>
            Open action
          </Button>
          <Button onClick={handleSnooze1h}>Snooze 1h</Button>
          <Button danger icon={<DeleteOutlined />} onClick={() => { void deleteNotification(notif.id); onBack?.(); }}>
            Delete
          </Button>
        </Space>
      </div>
    </section>
  );
}
