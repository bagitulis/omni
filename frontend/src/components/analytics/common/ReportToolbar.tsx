import { Space, Button, Modal } from "antd";
import {
  SyncOutlined,
  DeleteOutlined,
  ReloadOutlined,
  DownloadOutlined,
} from "@ant-design/icons";

interface ReportToolbarProps {
  onSync: () => void;
  onDelete: () => void;
  onRepopulate: () => void;
  onExportCSV: () => void;
  syncing?: boolean;
  hasData?: boolean;
}

/**
 * Action buttons row for the report page.
 * Includes Sync, Delete Sync, Repopulate Items, and Export CSV.
 */
export function ReportToolbar({
  onSync,
  onDelete,
  onRepopulate,
  onExportCSV,
  syncing,
  hasData,
}: ReportToolbarProps) {
  const handleDelete = () => {
    Modal.confirm({
      title: "Delete Sync Data",
      content:
        "Are you sure you want to delete the sync data for this period? This action cannot be undone.",
      okText: "Delete",
      okType: "danger",
      cancelText: "Cancel",
      onOk: onDelete,
    });
  };

  return (
    <Space wrap>
      <Button
        type="primary"
        icon={<SyncOutlined />}
        onClick={onSync}
        loading={syncing}
        disabled={syncing}
        size="small"
        aria-label={syncing ? "Sync report data in progress" : "Sync report data"}
      >
        Sync
      </Button>
      <Button
        danger
        icon={<DeleteOutlined />}
        onClick={handleDelete}
        size="small"
        aria-label="Delete synced report data"
      >
        Delete Sync
      </Button>
      <Button
        icon={<ReloadOutlined />}
        onClick={onRepopulate}
        size="small"
        aria-label="Repopulate report items"
      >
        Repopulate Items
      </Button>
      <Button
        icon={<DownloadOutlined />}
        onClick={onExportCSV}
        disabled={!hasData}
        size="small"
        aria-label="Export report data as CSV"
      >
        Export CSV
      </Button>
    </Space>
  );
}
