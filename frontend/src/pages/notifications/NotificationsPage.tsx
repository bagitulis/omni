import { ArrowLeftOutlined, ReloadOutlined, SettingOutlined } from "@ant-design/icons";
import { Alert, Button, Empty, Grid, Typography } from "antd";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import type { Notification, NotificationSeverity } from "@/api/notifications";
import { useNotifications } from "@/contexts/NotificationContext";
import { NotificationList } from "@/components/notifications/NotificationList";
import { NotificationDetailPane } from "@/components/notifications/NotificationDetailPane";
import { NotificationFilterBar, type StatusFilter } from "@/components/notifications/NotificationFilterBar";
import { NotificationBulkToolbar } from "@/components/notifications/NotificationBulkToolbar";
import { NotificationSettingsDrawer } from "@/components/notifications/NotificationSettingsDrawer";
import "@/components/layout/notifications.css";
import "./NotificationsPage.css";

const { Title, Text } = Typography;
const { useBreakpoint } = Grid;

/**
 * Notifications page — master-detail layout with:
 *  - counters row (All/Unread/High/Critical) that also acts as quick-filter,
 *  - filter bar: category multi + severity + search,
 *  - master list grouped by day with severity rail + checkbox,
 *  - detail pane on the right at ≥lg viewport, drilldown on <lg.
 */
export default function NotificationsPage() {
  const {
    notifications,
    counts,
    loading,
    fetchNotifications,
    markAsRead,
    bulkMarkRead,
    bulkDelete,
    snooze,
  } = useNotifications();

  const screens = useBreakpoint();
  const isWide = Boolean(screens.lg);

  const [status, setStatus] = useState<StatusFilter>("all");
  const [categories, setCategories] = useState<string[]>([]);
  const [minSeverity, setMinSeverity] = useState<NotificationSeverity | 0>(0);
  const [search, setSearch] = useState("");
  const [activeId, setActiveId] = useState<number | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [refreshError, setRefreshError] = useState<string | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [searchParams] = useSearchParams();

  // Deep-link: /notifications?expand=<id>
  useEffect(() => {
    const expandId = searchParams.get("expand");
    if (!expandId) return;
    setStatus("all");
    setCategories([]);
    setMinSeverity(0);
    setSearch("");
    const id = Number(expandId);
    if (!Number.isNaN(id)) {
      setActiveId(id);
      const notif = notifications.find((n) => n.id === id);
      if (notif && !notif.read) void markAsRead(id);
    }
  }, [searchParams, notifications, markAsRead]);

  const filtered = useMemo(() => {
    let items = notifications;
    if (status === "unread") items = items.filter((n) => !n.read);
    if (status === "read") items = items.filter((n) => n.read);
    if (categories.length > 0) items = items.filter((n) => categories.includes(n.category));
    if (minSeverity > 0) items = items.filter((n) => n.severity >= minSeverity);
    const q = search.trim().toLowerCase();
    if (q !== "") {
      items = items.filter(
        (n) => n.title.toLowerCase().includes(q) || n.message.toLowerCase().includes(q),
      );
    }
    return items;
  }, [notifications, status, categories, minSeverity, search]);

  const active = useMemo(
    () => filtered.find((n) => n.id === activeId) ?? null,
    [filtered, activeId],
  );

  const handleRefresh = useCallback(async () => {
    setRefreshError(null);
    try {
      await fetchNotifications({ limit: 50 });
    } catch {
      setRefreshError("Unable to refresh notifications. Try again in a moment.");
    }
  }, [fetchNotifications]);

  const handleOpen = useCallback((item: Notification) => {
    setActiveId(item.id);
    if (!item.read) void markAsRead(item.id);
  }, [markAsRead]);

  const handleSelectToggle = useCallback((id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const handleClearSelection = useCallback(() => setSelectedIds(new Set()), []);

  const handleBulkMarkRead = useCallback(async () => {
    await bulkMarkRead(Array.from(selectedIds));
    handleClearSelection();
  }, [bulkMarkRead, handleClearSelection, selectedIds]);

  const handleBulkDelete = useCallback(async () => {
    await bulkDelete(Array.from(selectedIds));
    handleClearSelection();
  }, [bulkDelete, handleClearSelection, selectedIds]);

  const handleBulkSnooze = useCallback(async () => {
    const until = new Date(Date.now() + 60 * 60 * 1000);
    for (const id of Array.from(selectedIds)) {
      await snooze(id, until);
    }
    handleClearSelection();
  }, [handleClearSelection, selectedIds, snooze]);

  const showDetail = active !== null && (isWide || activeId !== null);
  const showList = !activeId || isWide;

  return (
    <div className="notifications-page" data-testid="notifications-page">
      {/* Header */}
      <div className="notifications-page__header">
        <div>
          <Title level={3} className="notifications-page__title">
            Notifications
          </Title>
          <Text type="secondary">
            {counts.total} total · {counts.unread} unread ·{" "}
            {counts.by_severity["50"] ?? 0} critical
          </Text>
        </div>
        <div className="notifications-page__header-actions">
          <Button icon={<ReloadOutlined />} onClick={handleRefresh} loading={loading}>
            Refresh
          </Button>
          <Button icon={<SettingOutlined />} onClick={() => setSettingsOpen(true)}>
            Settings
          </Button>
        </div>
      </div>

      {refreshError && (
        <Alert
          type="error"
          showIcon
          message={refreshError}
          closable
          onClose={() => setRefreshError(null)}
          className="notifications-page__error"
        />
      )}

      <NotificationFilterBar
        status={status}
        onStatusChange={setStatus}
        categories={categories}
        onCategoriesChange={setCategories}
        minSeverity={minSeverity}
        onMinSeverityChange={setMinSeverity}
        search={search}
        onSearchChange={setSearch}
      />

      <NotificationBulkToolbar
        selectedIds={Array.from(selectedIds)}
        onMarkRead={handleBulkMarkRead}
        onDelete={handleBulkDelete}
        onSnooze={handleBulkSnooze}
        onClear={handleClearSelection}
      />

      <div
        className={`notifications-page__body${showDetail && isWide ? " notifications-page__body--two-col" : ""}`}
      >
        {showList && (
          <div className="notifications-page__list-col">
            <NotificationList
              items={filtered}
              loading={loading}
              activeId={activeId}
              selectedIds={selectedIds}
              onOpen={handleOpen}
              onSelectToggle={handleSelectToggle}
            />
          </div>
        )}

        {showDetail && active && (
          <div className="notifications-page__detail-col">
            {!isWide && (
              <Button
                type="text"
                icon={<ArrowLeftOutlined />}
                onClick={() => setActiveId(null)}
                className="notifications-page__back"
                aria-label="Back to list"
              >
                Back
              </Button>
            )}
            <NotificationDetailPane
              notif={active}
              onBack={isWide ? undefined : () => setActiveId(null)}
            />
          </div>
        )}

        {!showDetail && isWide && (
          <div className="notifications-page__detail-col notifications-page__detail-col--empty">
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="Select a notification to view details"
            />
          </div>
        )}
      </div>

      <NotificationSettingsDrawer open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </div>
  );
}
