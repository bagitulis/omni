import { useEffect, useRef, useState } from "react";
import { Spin, Typography, theme } from "antd";
import { SyncOutlined } from "@ant-design/icons";

interface MonitorData {
  current_job: { name: string; status: string } | null;
  recent_history: { function_name: string; executed_at: string; status: string }[];
}

function formatRelativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

/**
 * Compact indicator showing last sync time and active sync status.
 * Polls /api/script-monitor every 30s (lightweight).
 */
export function SyncStatusIndicator() {
  const { token } = theme.useToken();
  const [data, setData] = useState<MonitorData | null>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const res = await fetch("/api/script-monitor", { credentials: "include" });
        if (!res.ok) return;
        const json = await res.json();
        if (json.success) setData(json.data);
      } catch {
        // Silent fail
      }
    };

    fetchStatus();
    intervalRef.current = setInterval(fetchStatus, 30000);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, []);

  if (!data) return null;

  const isSyncing = data.current_job?.status === "running";
  const lastSync = data.recent_history?.find(
    (h) => h.function_name === "sync_products_inventory" && h.status === "success",
  );

  return (
    <div style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
      {isSyncing ? (
        <>
          <Spin size="small" indicator={<SyncOutlined spin style={{ fontSize: 12 }} />} />
          <Typography.Text style={{ fontSize: 12, color: token.colorPrimary }}>
            Syncing...
          </Typography.Text>
        </>
      ) : lastSync ? (
        <Typography.Text style={{ fontSize: 12, color: token.colorTextSecondary }}>
          Last synced: {formatRelativeTime(lastSync.executed_at)}
        </Typography.Text>
      ) : (
        <Typography.Text style={{ fontSize: 12, color: token.colorTextTertiary }}>
          Never synced
        </Typography.Text>
      )}
    </div>
  );
}
