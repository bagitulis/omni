import { Button, Switch, Tag, Typography } from "antd";
import type { TableProps } from "antd";
import type { RouteConfig } from "@/types/routeConfig";

const { Text } = Typography;

export const ROUTE_PRESETS: Record<string, Partial<RouteConfig>> = {
  api_heavy: {
    cache_ttl: 60,
    max_concurrent: 10,
    timeout: 30000,
  },
  cache_heavy: {
    caching_enabled: true,
    cache_ttl: 600,
    cache_strategy: "aggressive",
  },
  balanced: {
    caching_enabled: true,
    cache_ttl: 300,
    max_concurrent: 5,
    timeout: 15000,
  },
};

const METHOD_COLORS: Record<string, string> = {
  GET: "#52c41a",
  POST: "#1890ff",
  PUT: "#faad14",
  DELETE: "#f5222d",
};

interface CreateRouteColumnsParams {
  isMutating: boolean;
  fontSizeSM: number;
  fontWeightStrong: number;
  textSecondaryColor: string;
  onToggleBoolean: (
    id: number,
    field: "enabled" | "caching_enabled" | "queue_enabled",
    value: boolean,
  ) => void;
  onEdit: (route: RouteConfig) => void;
}

export function createRouteColumns({
  isMutating,
  fontSizeSM,
  fontWeightStrong,
  textSecondaryColor,
  onToggleBoolean,
  onEdit,
}: CreateRouteColumnsParams): TableProps<RouteConfig>["columns"] {
  return [
    {
      title: "Route Path",
      dataIndex: "route_path",
      key: "route_path",
      width: 240,
      render: (routePath: string) => (
        <Text code style={{ fontSize: fontSizeSM }}>
          {routePath}
        </Text>
      ),
    },
    {
      title: "Method",
      dataIndex: "route_method",
      key: "route_method",
      width: 100,
      render: (method: string) => {
        const color = METHOD_COLORS[method] || textSecondaryColor;
        return (
          <Tag
            color={color}
            style={{ marginInlineEnd: 0, fontWeight: fontWeightStrong }}
          >
            {method}
          </Tag>
        );
      },
    },
    {
      title: "Description",
      dataIndex: "description",
      key: "description",
      ellipsis: true,
      render: (description: string) => (
        <Text type="secondary" style={{ fontSize: fontSizeSM }}>
          {description || "-"}
        </Text>
      ),
    },
    {
      title: "Enabled",
      dataIndex: "enabled",
      key: "enabled",
      width: 90,
      render: (enabled: boolean, record) => (
        <Switch
          checked={enabled}
          size="small"
          disabled={isMutating}
          onClick={(checked, event) => {
            event?.stopPropagation();
            onToggleBoolean(record.id, "enabled", checked);
          }}
        />
      ),
    },
    {
      title: "Caching",
      dataIndex: "caching_enabled",
      key: "caching_enabled",
      width: 90,
      render: (cachingEnabled: boolean, record) => (
        <Switch
          checked={cachingEnabled}
          size="small"
          disabled={isMutating}
          onClick={(checked, event) => {
            event?.stopPropagation();
            onToggleBoolean(record.id, "caching_enabled", checked);
          }}
        />
      ),
    },
    {
      title: "Cache TTL",
      dataIndex: "cache_ttl",
      key: "cache_ttl",
      width: 110,
      render: (cacheTtl: number, record) =>
        record.caching_enabled ? `${cacheTtl}s` : "-",
    },
    {
      title: "Queue",
      dataIndex: "queue_enabled",
      key: "queue_enabled",
      width: 80,
      render: (queueEnabled: boolean, record) => (
        <Switch
          checked={queueEnabled}
          size="small"
          disabled={isMutating}
          onClick={(checked, event) => {
            event?.stopPropagation();
            onToggleBoolean(record.id, "queue_enabled", checked);
          }}
        />
      ),
    },
    {
      title: "Timeout",
      dataIndex: "timeout",
      key: "timeout",
      width: 110,
      render: (timeout: number) => `${timeout}ms`,
    },
    {
      title: "Actions",
      key: "actions",
      width: 90,
      fixed: "right",
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          style={{ padding: 0 }}
          onClick={(event) => {
            event.stopPropagation();
            onEdit(record);
          }}
        >
          Edit
        </Button>
      ),
    },
  ];
}
