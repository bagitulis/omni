import { useMemo, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Col,
  Dropdown,
  Row,
  Space,
  Spin,
  Statistic,
  Table,
  message,
  theme,
} from "antd";
import type { MenuProps } from "antd";
import {
  useBulkUpdateRouteConfigs,
  useRouteConfigs,
  useUpdateRouteConfig,
} from "@/hooks/useRouteConfig";
import type { RouteConfig } from "@/types/routeConfig";
import { RouteConfigModal } from "../components/RouteConfigModal";
import {
  ROUTE_PRESETS,
  createRouteColumns,
} from "../components/RouteManagementTableColumns";

const { useToken } = theme;

export default function RouteManagementTab() {
  const { token } = useToken();
  const [editingRoute, setEditingRoute] = useState<RouteConfig | null>(null);

  const {
    data: routes = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useRouteConfigs();
  const updateRouteConfig = useUpdateRouteConfig();
  const bulkUpdateRouteConfigs = useBulkUpdateRouteConfigs();

  const isMutating =
    updateRouteConfig.isPending || bulkUpdateRouteConfigs.isPending;

  const stats = useMemo(
    () => ({
      total_routes: routes.length,
      enabled_routes: routes.filter((route) => route.enabled).length,
      cached_routes: routes.filter((route) => route.caching_enabled).length,
      queued_routes: routes.filter((route) => route.queue_enabled).length,
    }),
    [routes],
  );

  const updateBooleanField = (
    id: number,
    field: "enabled" | "caching_enabled" | "queue_enabled",
    value: boolean,
  ) => {
    updateRouteConfig.mutate({ id, data: { [field]: value } });
  };

  const applyBulkUpdate = (data: Partial<RouteConfig>) => {
    if (routes.length === 0) {
      message.warning("No route configs available");
      return;
    }

    bulkUpdateRouteConfigs.mutate({
      ids: routes.map((route) => route.id),
      data,
    });
  };

  const handlePresetSelect: MenuProps["onClick"] = ({ key }) => {
    const preset = ROUTE_PRESETS[key];
    if (!preset) {
      return;
    }
    applyBulkUpdate(preset);
  };

  const presetMenuItems: MenuProps["items"] = [
    { key: "api_heavy", label: "API-Heavy" },
    { key: "cache_heavy", label: "Cache-Heavy" },
    { key: "balanced", label: "Balanced" },
  ];

  const columns = createRouteColumns({
    isMutating,
    fontSizeSM: token.fontSizeSM,
    fontWeightStrong: token.fontWeightStrong,
    textSecondaryColor: token.colorTextSecondary,
    onToggleBoolean: updateBooleanField,
    onEdit: setEditingRoute,
  });

  if (isLoading) {
    return (
      <div
        style={{
          minHeight: 240,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Spin size="large" />
      </div>
    );
  }

  if (isError) {
    const errorMessage =
      error instanceof Error ? error.message : "Failed to load route configs";

    return (
      <Alert
        type="error"
        showIcon
        message="Unable to load route configuration"
        description={errorMessage}
        action={
          <Button size="small" onClick={() => refetch()}>
            Retry
          </Button>
        }
      />
    );
  }

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic title="Total Routes" value={stats.total_routes} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic title="Enabled Routes" value={stats.enabled_routes} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic title="Cached Routes" value={stats.cached_routes} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic title="Queued Routes" value={stats.queued_routes} />
          </Card>
        </Col>
      </Row>

      <Card
        size="small"
        title="Route Configuration"
        extra={
          <Space wrap>
            <Button
              onClick={() => applyBulkUpdate({ enabled: true })}
              disabled={isMutating || routes.length === 0}
            >
              Enable All
            </Button>
            <Button
              onClick={() => applyBulkUpdate({ enabled: false })}
              disabled={isMutating || routes.length === 0}
            >
              Disable All
            </Button>
            <Dropdown
              menu={{ items: presetMenuItems, onClick: handlePresetSelect }}
              trigger={["click"]}
              disabled={isMutating || routes.length === 0}
            >
              <Button>Apply Preset</Button>
            </Dropdown>
          </Space>
        }
        style={{ borderRadius: token.borderRadius }}
      >
        <Table<RouteConfig>
          rowKey="id"
          columns={columns}
          dataSource={routes}
          size="small"
          scroll={{ x: 1240 }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `${total} routes`,
          }}
          onRow={(record) => ({
            onClick: () => setEditingRoute(record),
          })}
        />
      </Card>

      <RouteConfigModal
        open={Boolean(editingRoute)}
        route={editingRoute}
        onClose={() => setEditingRoute(null)}
      />
    </Space>
  );
}
