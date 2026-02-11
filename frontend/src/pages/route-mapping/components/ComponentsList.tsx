import { Card, Col, Empty, Row, Tag } from "antd";
import { ApiOutlined } from "@ant-design/icons";
import { ComponentDetail } from "@/types/routeMapping";

interface ComponentsListProps {
  filteredComponents: Record<string, ComponentDetail>;
}

export function ComponentsList({ filteredComponents }: ComponentsListProps) {
  if (Object.keys(filteredComponents).length === 0) {
    return (
      <div className="p-4 bg-gray-50">
        <Empty />
      </div>
    );
  }

  return (
    <div className="p-4 bg-gray-50">
      <Row gutter={[16, 16]}>
        {Object.entries(filteredComponents).map(([name, detail]) => (
          <Col span={8} key={name}>
            <Card
              title={name}
              size="small"
              extra={<Tag>{detail.routes_called?.length || 0} Calls</Tag>}
            >
              <div className="space-y-2">
                {detail.path && (
                  <div
                    className="text-xs text-gray-500 font-mono mb-2 truncate"
                    title={detail.path}
                  >
                    {detail.path}
                  </div>
                )}
                <div className="max-h-40 overflow-y-auto">
                  {detail.routes_called?.map((route: string) => (
                    <div
                      key={route}
                      className="text-xs border-b border-gray-100 py-1 last:border-0 truncate"
                      title={route}
                    >
                      <ApiOutlined className="mr-1 text-blue-500" />
                      {route}
                    </div>
                  ))}
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
}
