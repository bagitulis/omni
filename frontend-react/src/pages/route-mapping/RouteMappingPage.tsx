import { Card, Typography, List } from "antd";

const { Title, Text } = Typography;

export function RouteMappingPage() {
  const routes = [
    { path: "/", name: "Dashboard" },
    { path: "/products", name: "Master Products" },
    { path: "/product-manager", name: "Product Manager" },
    { path: "/orders", name: "Orders" },
    { path: "/inventory", name: "Inventory" },
    { path: "/analytics", name: "Analytics" },
    { path: "/settings", name: "Settings" },
    { path: "/route-mapping", name: "Route Mapping" },
    { path: "/script-monitor", name: "Script Monitor" },
  ];

  return (
    <div className="p-6">
      <Title level={2}>Route Mapping</Title>
      <Card>
        <List
          header={
            <div>
              <strong>Application Routes</strong>
            </div>
          }
          bordered
          dataSource={routes}
          renderItem={(item) => (
            <List.Item>
              <div className="flex justify-between w-full">
                <Text strong>{item.name}</Text>
                <Text code>{item.path}</Text>
              </div>
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
}
