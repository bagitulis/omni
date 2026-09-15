import { Empty, Skeleton, Typography } from "antd";
import { useMemo } from "react";
import type { Notification } from "@/api/notifications";
import { NotificationListItem } from "./NotificationListItem";

const { Text } = Typography;

interface NotificationListProps {
  items: Notification[];
  loading: boolean;
  activeId: number | null;
  selectedIds: Set<number>;
  onOpen: (item: Notification) => void;
  onSelectToggle: (id: number) => void;
}

/** Group notifications into buckets by day for scannability. */
function groupByDay(items: Notification[]): Array<{ label: string; items: Notification[] }> {
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const startOfYesterday = startOfToday - 86400_000;
  const startOfWeek = startOfToday - 6 * 86400_000;

  const today: Notification[] = [];
  const yesterday: Notification[] = [];
  const thisWeek: Notification[] = [];
  const older: Notification[] = [];
  for (const it of items) {
    const t = new Date(it.created_at).getTime();
    if (t >= startOfToday) today.push(it);
    else if (t >= startOfYesterday) yesterday.push(it);
    else if (t >= startOfWeek) thisWeek.push(it);
    else older.push(it);
  }
  return [
    { label: "Today", items: today },
    { label: "Yesterday", items: yesterday },
    { label: "Last 7 days", items: thisWeek },
    { label: "Older", items: older },
  ].filter((g) => g.items.length > 0);
}

export function NotificationList({
  items,
  loading,
  activeId,
  selectedIds,
  onOpen,
  onSelectToggle,
}: NotificationListProps) {
  const groups = useMemo(() => groupByDay(items), [items]);

  if (loading) {
    return (
      <div className="notification-list__state">
        <Skeleton active paragraph={{ rows: 5 }} />
      </div>
    );
  }
  if (items.length === 0) {
    return (
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description="No notifications"
        className="notification-list__state"
      />
    );
  }
  return (
    <div className="notification-list" role="list">
      {groups.map((g) => (
        <div key={g.label} className="notification-list__group">
          <Text type="secondary" className="notification-list__group-label">
            {g.label}
          </Text>
          {g.items.map((it) => (
            <NotificationListItem
              key={it.id}
              item={it}
              active={activeId === it.id}
              selected={selectedIds.has(it.id)}
              onSelectToggle={onSelectToggle}
              onOpen={onOpen}
            />
          ))}
        </div>
      ))}
    </div>
  );
}
