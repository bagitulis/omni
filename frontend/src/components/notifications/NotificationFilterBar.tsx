import { Input, Segmented, Select } from "antd";
import { SEVERITY_META } from "./severity";
import type { NotificationSeverity } from "@/api/notifications";

export type StatusFilter = "all" | "unread" | "read";

interface NotificationFilterBarProps {
  status: StatusFilter;
  onStatusChange: (s: StatusFilter) => void;
  categories: string[];
  onCategoriesChange: (c: string[]) => void;
  minSeverity: NotificationSeverity | 0;
  onMinSeverityChange: (s: NotificationSeverity | 0) => void;
  search: string;
  onSearchChange: (q: string) => void;
}

const CATEGORY_OPTIONS = [
  { label: "Sync", value: "sync" },
  { label: "Order", value: "order" },
  { label: "Product", value: "product" },
  { label: "Inventory", value: "inventory" },
  { label: "System", value: "system" },
  { label: "Auth", value: "auth" },
  { label: "Export", value: "export" },
  { label: "Security", value: "security" },
];

const SEVERITY_OPTIONS = [
  { label: "Any severity", value: 0 },
  ...(Object.entries(SEVERITY_META) as [string, (typeof SEVERITY_META)[NotificationSeverity]][])
    .map(([value, meta]) => ({ label: `${meta.label}+`, value: Number(value) })),
];

/**
 * Filter bar for the notifications page.
 * Layout: status segmented + category multi + severity select + search input.
 * Wraps to one row per control on <sm.
 */
export function NotificationFilterBar(props: NotificationFilterBarProps) {
  return (
    <div className="notification-filter-bar">
      <Segmented
        value={props.status}
        onChange={(v) => props.onStatusChange(v as StatusFilter)}
        options={[
          { label: "All", value: "all" },
          { label: "Unread", value: "unread" },
          { label: "Read", value: "read" },
        ]}
      />
      <Select
        mode="multiple"
        allowClear
        placeholder="Categories"
        value={props.categories}
        onChange={props.onCategoriesChange}
        options={CATEGORY_OPTIONS}
        className="notification-filter-bar__category"
      />
      <Select
        value={props.minSeverity}
        onChange={(v) => props.onMinSeverityChange(v as NotificationSeverity | 0)}
        options={SEVERITY_OPTIONS}
        className="notification-filter-bar__severity"
      />
      <Input.Search
        placeholder="Search title or message"
        value={props.search}
        onChange={(e) => props.onSearchChange(e.target.value)}
        allowClear
        className="notification-filter-bar__search"
      />
    </div>
  );
}
