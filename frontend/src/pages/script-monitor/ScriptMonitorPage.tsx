import { Tabs, Spin, Alert, Typography } from "antd";
import { useSearchParams } from "react-router-dom";
import { useScriptMonitor } from "@/hooks/useScriptMonitor";
import { CurrentJobTab } from "./components/CurrentJobTab";
import { QueueTab } from "./components/QueueTab";
import { HistoryTab } from "./components/HistoryTab";
import { ConfigTab } from "./components/ConfigTab";
import {
  CodeOutlined,
  UnorderedListOutlined,
  HistoryOutlined,
  SettingOutlined,
} from "@ant-design/icons";

export function ScriptMonitorPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get("tab");
  const tabParamByKey: Record<string, string> = {
    current: "current",
    queue: "queue",
    history: "history",
    config: "auto-functions",
  };
  const keyByTabParam: Record<string, string> = {
    current: "current",
    queue: "queue",
    history: "history",
    "auto-functions": "config",
    config: "config",
  };
  const activeTab = keyByTabParam[rawTab || ""] || "current";
  const {
    monitorData,
    isLoadingMonitor,
    autoFunctions,
    isLoadingAutoFunctions,
    cancelJob,
    forceCancelJob,
    clearHistory,
    enableAutoFunction,
    disableAutoFunction,
    runAutoFunction,
    createAutoFunction,
    updateAutoFunction,
    deleteAutoFunction,
    cancelScheduled,
  } = useScriptMonitor();

  if (isLoadingMonitor && !monitorData) {
    return (
      <div style={{ padding: 24, textAlign: "center" }}>
        <Spin size="large" tip="Loading monitor data..." />
      </div>
    );
  }

  if (!monitorData && !isLoadingMonitor) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          type="error"
          message="Connection Error"
          description="Failed to load script monitor data. Please check your connection."
          showIcon
        />
      </div>
    );
  }

  const items = [
    {
      key: "current",
      label: (
        <span>
          <CodeOutlined />
          Current Job
        </span>
      ),
      children: (
        <CurrentJobTab
          job={monitorData?.current_job || null}
          onCancel={cancelJob}
          onForceCancel={forceCancelJob}
        />
      ),
    },
    {
      key: "queue",
      label: (
        <span>
          <UnorderedListOutlined />
          Queue{" "}
          {monitorData?.total_pending ? `(${monitorData.total_pending})` : ""}
        </span>
      ),
      children: (
        <QueueTab
          queue={monitorData?.pending_queue || []}
          loading={isLoadingMonitor}
          onCancel={cancelJob}
          onForceCancel={forceCancelJob}
        />
      ),
    },
    {
      key: "history",
      label: (
        <span>
          <HistoryOutlined />
          History
        </span>
      ),
      children: (
        <HistoryTab
          history={monitorData?.recent_history || []}
          loading={isLoadingMonitor}
          onClear={clearHistory}
        />
      ),
    },
    {
      key: "config",
      label: (
        <span>
          <SettingOutlined />
          Auto-Functions
        </span>
      ),
      children: (
        <ConfigTab
          configs={autoFunctions || []}
          loading={isLoadingAutoFunctions}
          onEnable={enableAutoFunction}
          onDisable={disableAutoFunction}
          onRun={runAutoFunction}
          onDelete={deleteAutoFunction}
          onCancelScheduled={cancelScheduled}
          onCreate={(config, onSuccess) =>
            createAutoFunction(config, {
              onSuccess,
            })
          }
          onUpdate={(payload, onSuccess) =>
            updateAutoFunction(payload, {
              onSuccess,
            })
          }
        />
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Typography.Title level={2} style={{ marginBottom: 24, marginTop: 0 }}>
        Script Monitor
      </Typography.Title>
      <Tabs
        activeKey={activeTab}
        onChange={(key) => {
          const next = new URLSearchParams(searchParams);
          next.set("tab", tabParamByKey[key] || "current");
          setSearchParams(next);
        }}
        items={items}
        destroyInactiveTabPane
      />
    </div>
  );
}

export default ScriptMonitorPage;
