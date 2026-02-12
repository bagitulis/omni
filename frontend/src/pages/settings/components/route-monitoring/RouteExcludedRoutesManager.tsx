import { useState } from "react";
import { Button, Card, Input, Space, Tag, Typography, theme } from "antd";
import { DeleteOutlined, PlusOutlined } from "@ant-design/icons";

const { Text } = Typography;

export function RouteExcludedRoutesManager() {
  const { token } = theme.useToken();
  const [new_route, setNewRoute] = useState("");
  const [excluded_routes, setExcludedRoutes] = useState<string[]>([]);

  const handleAddRoute = () => {
    const normalized = new_route.trim();
    if (!normalized || excluded_routes.includes(normalized)) {
      return;
    }
    setExcludedRoutes((prev) => [...prev, normalized]);
    setNewRoute("");
  };

  const handleRemoveRoute = (route: string) => {
    setExcludedRoutes((prev) => prev.filter((item) => item !== route));
  };

  return (
    <Card
      size="small"
      title="Excluded Routes"
      style={{ borderRadius: token.borderRadius }}
    >
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        <Space.Compact style={{ width: "100%" }}>
          <Input
            value={new_route}
            placeholder="Add route path (e.g. /api/orders)"
            onChange={(event) => setNewRoute(event.target.value)}
            onPressEnter={handleAddRoute}
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleAddRoute}
          >
            Add
          </Button>
        </Space.Compact>
        {excluded_routes.length === 0 ? (
          <Text type="secondary">No excluded routes configured.</Text>
        ) : (
          <Space wrap>
            {excluded_routes.map((route) => (
              <Tag
                key={route}
                color={token.colorPrimary}
                closeIcon={<DeleteOutlined />}
                onClose={() => handleRemoveRoute(route)}
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
