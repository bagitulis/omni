import { useState } from "react";
import { Button, Card, Input, Space, Tag, Typography, theme } from "antd";
import { PlusOutlined } from "@ant-design/icons";

const { Text } = Typography;

const PRESET_ROUTES = ["/api/orders", "/api/products", "/api/inventory"];

export function RouteFlowControls() {
  const { token } = theme.useToken();
  const [input_route, setInputRoute] = useState("");
  const [direct_routes, setDirectRoutes] = useState<string[]>([]);

  const addRoute = (route_value: string) => {
    const normalized = route_value.trim();
    if (!normalized || direct_routes.includes(normalized)) {
      return;
    }
    setDirectRoutes((prev) => [...prev, normalized]);
  };

  const removeRoute = (route_value: string) => {
    setDirectRoutes((prev) => prev.filter((route) => route !== route_value));
  };

  const togglePreset = (preset_route: string) => {
    if (direct_routes.includes(preset_route)) {
      removeRoute(preset_route);
      return;
    }
    addRoute(preset_route);
  };

  return (
    <Card
      size="small"
      title="Route Flow Controls"
      style={{ borderRadius: token.borderRadius }}
    >
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        <Text type="secondary">
          Direct routes bypass queue ordering while preserving route monitoring
          visibility.
        </Text>
        <Space.Compact style={{ width: "100%" }}>
          <Input
            value={input_route}
            placeholder="Add direct route"
            onChange={(event) => setInputRoute(event.target.value)}
            onPressEnter={() => {
              addRoute(input_route);
              setInputRoute("");
            }}
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              addRoute(input_route);
              setInputRoute("");
            }}
          >
            Add
          </Button>
        </Space.Compact>
        <Space wrap>
          {PRESET_ROUTES.map((preset) => {
            const is_active = direct_routes.includes(preset);
            return (
              <Button
                key={preset}
                size="small"
                type={is_active ? "primary" : "default"}
                onClick={() => togglePreset(preset)}
              >
                {preset}
              </Button>
            );
          })}
        </Space>
        {direct_routes.length === 0 ? (
          <Text type="secondary">No direct routes selected.</Text>
        ) : (
          <Space wrap>
            {direct_routes.map((route) => (
              <Tag
                key={route}
                color={token.colorPrimary}
                closable
                onClose={() => removeRoute(route)}
              >
                {route}
              </Tag>
            ))}
          </Space>
        )}
      </Space>
    </Card>
  );
}
