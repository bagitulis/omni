import { Tabs, Spin, Alert } from "antd";
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
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Script Monitor</h1>
      <Tabs defaultActiveKey="current" items={items} destroyInactiveTabPane />
    </div>
  );
}

export default ScriptMonitorPage;
