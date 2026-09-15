import { CheckOutlined, ClockCircleOutlined, DeleteOutlined } from "@ant-design/icons";
import { Button, Popconfirm, Typography } from "antd";

const { Text } = Typography;

interface NotificationBulkToolbarProps {
  selectedIds: number[];
  onMarkRead: () => void;
  onDelete: () => void;
  onSnooze: () => void;
  onClear: () => void;
}

/**
 * Sticky action bar shown when at least one notification is selected.
 * Layout: count on the left, actions on the right. Destructive delete is
 * gated behind a Popconfirm.
 */
export function NotificationBulkToolbar(props: NotificationBulkToolbarProps) {
  const n = props.selectedIds.length;
  if (n === 0) return null;
  return (
    <div className="notification-bulk-toolbar" role="toolbar" aria-label="Bulk actions">
      <Text strong>{n} selected</Text>
      <div className="notification-bulk-toolbar__actions">
        <Button icon={<CheckOutlined />} onClick={props.onMarkRead}>
          Mark read
        </Button>
        <Button icon={<ClockCircleOutlined />} onClick={props.onSnooze}>
          Snooze 1h
        </Button>
        <Popconfirm
          title={`Delete ${n} notification${n === 1 ? "" : "s"}?`}
          okText="Delete"
          okButtonProps={{ danger: true }}
          cancelText="Cancel"
          onConfirm={props.onDelete}
        >
          <Button danger icon={<DeleteOutlined />}>
            Delete
          </Button>
        </Popconfirm>
        <Button type="text" onClick={props.onClear}>
          Clear
        </Button>
      </div>
    </div>
  );
}
