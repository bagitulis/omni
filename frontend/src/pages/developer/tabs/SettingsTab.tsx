import { useCallback, useEffect, useState } from "react";
import { Alert, Descriptions, Spin, Typography } from "antd";
import { developerApi, type EnvironmentInfo } from "@/api/developer";
import { useAuthStore } from "@/stores/authStore";

const { Text } = Typography;

export default function SettingsTab() {
  const accessToken = useAuthStore((state) => state.accessToken);
  const [info, setInfo] = useState<EnvironmentInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchEnvironment = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await developerApi.getEnvironmentInfo();
      if (response.success && response.data) {
        setInfo(response.data);
      } else {
        setError(response.error || "Failed to load environment info");
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to load environment info";
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (accessToken) {
      void fetchEnvironment();
    }
  }, [accessToken, fetchEnvironment]);

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert
        type="error"
        showIcon
        message="Failed to load environment info"
        description={error}
      />
    );
  }

  if (!info) {
    return null;
  }

  const items = [
    { key: "version", label: "App Version", children: info.version },
    { key: "environment", label: "Environment", children: info.environment },
    { key: "db_driver", label: "Database Driver", children: info.db_driver },
    { key: "go_version", label: "Go Version", children: info.go_version },
    { key: "node_env", label: "Node Environment", children: info.node_env },
    { key: "build_time", label: "Build Time", children: info.build_time },
    { key: "api_base_url", label: "API Base URL", children: info.api_base_url },
    { key: "dev_login_enabled", label: "Dev Login Enabled", children: info.dev_login_enabled },
    { key: "active_tenants_count", label: "Active Tenants Count", children: String(info.active_tenants_count) },
  ];

  return (
    <>
      <Descriptions bordered column={1} items={items} />
      <Text type="secondary" style={{ display: "block", marginTop: 16 }}>
        Settings management coming in v2
      </Text>
    </>
  );
}
