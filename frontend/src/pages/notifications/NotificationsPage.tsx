import { useState, useMemo, useCallback, useEffect } from "react";
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
  Alert,
} from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { useSearchParams } from "react-router-dom";
import { useNotifications } from "@/contexts/NotificationContext";
import type { Notification } from "@/api/notifications";
import {
  formatRelativeTime,
  parseNotificationMessage,
  PlatformBreakdown,
  getNotificationTypeConfig,
  NotificationStats,
} from "@/components/layout/NotificationHelpers";
import "@/components/layout/notifications.css";

const { Title, Text } = Typography;

type StatusFilter = "all" | "unread" | "read";
type CategoryFilter = "all" | "sync" | "order" | "product" | "inventory" | "system";

export default function NotificationsPage() {
  const { token } = theme.useToken();
  const { notifications, loading, markAsRead, fetchNotifications } =
    useNotifications();

  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>("all");
  const [expandedKeys, setExpandedKeys] = useState<string[]>([]);
  const [refreshError, setRefreshError] = useState<string | null>(null);
  const [searchParams] = useSearchParams();

  // Auto-expand notification from query param (e.g., /notifications?expand=123)
  useEffect(() => {
    const expandId = searchParams.get("expand");
    if (expandId) {
      // Reset filters so the target notification is visible
      setStatusFilter("all");
      setCategoryFilter("all");
      setExpandedKeys([expandId]);
      // Mark as read
      const notif = notifications.find((n) => String(n.id) === expandId);
      if (notif && !notif.read) {
        markAsRead(notif.id);
      }
    }
  }, [markAsRead, notifications, searchParams]);

  const filtered = useMemo(() => {
    let items = notifications;
    if (statusFilter === "unread") items = items.filter((n) => !n.read);
    if (statusFilter === "read") items = items.filter((n) => n.read);
    if (categoryFilter !== "all")
      items = items.filter((n) => n.category === categoryFilter);
    return items;
  }, [notifications, statusFilter, categoryFilter]);

  const handleRefresh = useCallback(async () => {
    setRefreshError(null);
    try {
      await fetchNotifications();
    } catch {
      setRefreshError("Unable to refresh notifications. Try again in a moment.");
    }
  }, [fetchNotifications]);

  const handleExpand = useCallback(
    (keys: string | string[]) => {
      const newKeys = Array.isArray(keys) ? keys : [keys];
      // Find newly expanded keys (not previously in expandedKeys)
      const newlyExpanded = newKeys.filter((k) => !expandedKeys.includes(k));
      // Mark newly expanded as read
      newlyExpanded.forEach((key) => {
        const notif = notifications.find((n) => String(n.id) === key);
        if (notif && !notif.read) {
          markAsRead(notif.id);
        }
      });
      setExpandedKeys(newKeys);
    },
    [notifications, markAsRead, expandedKeys],
  );

  const collapseItems = filtered.map((notif) => {
    const parsed = parseNotificationMessage(notif.message);
    const typeConfig = getNotificationTypeConfig(notif.type, token);
    return {
      key: String(notif.id),
      label: (
        <div className="notification-page-item__label">
          <span className="notification-page-item__icon" style={{ color: typeConfig.color }}>{typeConfig.icon}</span>
          <div className="notification-page-item__main">
            <Text strong={!notif.read} ellipsis className="notification-page-item__title">
              {notif.title}
            </Text>
            {parsed.stats && (
              <Text type="secondary" ellipsis className="notification-page-item__stats">
                {parsed.stats.total !== undefined && `Total: ${parsed.stats.total}`}
                {parsed.stats.processed !== undefined && ` · Done: ${parsed.stats.processed}`}
                {parsed.stats.failed !== undefined && parsed.stats.failed > 0 && ` · Failed: ${parsed.stats.failed}`}
              </Text>
            )}
          </div>
          <Text type="secondary" className="notification-page-item__time">
            {formatRelativeTime(notif.created_at)}
          </Text>
          {!notif.read && (
            <div className="notification-page-item__unread-dot" />
          )}
        </div>
      ),
      children: (
        <NotificationDetail notif={notif} parsed={parsed} />
      ),
    };
  });

  return (
    <div className="notification-page" data-testid="notifications-page">
      <div className="notification-page__header">
        <Title level={3} className="notification-page__title">
          Notifications
        </Title>
        <Button icon={<ReloadOutlined />} onClick={handleRefresh} loading={loading}>
          Refresh
        </Button>
      </div>

      {refreshError && (
        <Alert type="error" showIcon message={refreshError} closable onClose={() => setRefreshError(null)} />
      )}

      {/* Filters */}
      <div className="notification-page__filters">
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
          className="notification-page__category-filter"
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
        <div className="notification-page__state">
          <Spin />
        </div>
      ) : filtered.length === 0 ? (
        <Empty description="No notifications match the current filters" className="notification-page__state" />
      ) : (
        <Collapse
          accordion={false}
          activeKey={expandedKeys}
          onChange={handleExpand}
          items={collapseItems}
          className="notification-page__collapse"
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
    <div className="notification-detail">
      {/* Platform breakdown */}
      {parsed.platforms && Object.keys(parsed.platforms).length > 0 && (
        <div className="notification-detail__section">
          <Text strong className="notification-detail__heading">
            Per Platform
          </Text>
          <PlatformBreakdown platforms={parsed.platforms} />
        </div>
      )}

      {/* Failed items */}
      {parsed.failedItems && parsed.failedItems.length > 0 && (
        <div className="notification-detail__section">
          <Text strong className="notification-detail__heading" style={{ color: token.colorError }}>
            Failed Items ({parsed.failedItems.length})
          </Text>
          <Table
            size="small"
            pagination={parsed.failedItems.length > 10 ? { pageSize: 10 } : false}
            dataSource={parsed.failedItems.map((item, i) => ({ ...item, key: i }))}
            className="notification-detail__table"
            scroll={{ x: 520 }}
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
        <Text type="secondary" className="notification-detail__summary">{parsed.summary}</Text>
      )}

      {parsed.stats && (
        <NotificationStats stats={parsed.stats} failed={notif.type === "error"} platforms={parsed.platforms} />
      )}

      {/* Metadata */}
      <div className="notification-detail__meta">
        <Text type="secondary" className="notification-detail__meta">
          Category: {notif.category} · Created: {new Date(notif.created_at).toLocaleString()}
        </Text>
      </div>
    </div>
  );
}
