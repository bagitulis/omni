import { ApiOutlined } from "@ant-design/icons";
import { Card, Col, Empty, Row, Tag, theme, Typography, Flex } from "antd";
import type { ComponentDetail } from "@/types/routeMapping";

interface ComponentsListProps {
  filteredComponents: Record<string, ComponentDetail>;
}

export function ComponentsList({ filteredComponents }: ComponentsListProps) {
  const { token } = theme.useToken();

  if (Object.keys(filteredComponents).length === 0) {
    return (
      <div style={{ padding: 16, background: token.colorBgLayout }}>
        <Empty />
      </div>
    );
  }

  return (
    <div style={{ padding: 16, background: token.colorBgLayout }}>
      <Row gutter={[16, 16]}>
        {Object.entries(filteredComponents).map(([name, detail]) => (
          <Col span={8} key={name}>
            <Card
              title={name}
              size="small"
              extra={<Tag>{detail.routes_called?.length || 0} Calls</Tag>}
            >
              <Flex vertical gap={8}>
                {detail.path && (
                  <Typography.Text
                    type="secondary"
                    style={{
                      fontSize: 10,
                      fontFamily: "monospace",
                      marginBottom: 8,
                      display: "block",
                    }}
                    ellipsis={{ tooltip: detail.path }}
                  >
                    {detail.path}
                  </Typography.Text>
                )}
                <div style={{ maxHeight: 160, overflowY: "auto" }}>
                  {detail.routes_called?.map((route: string) => (
                    <div
                      key={route}
                      style={{
                        fontSize: 10,
                        borderBottom: `1px solid ${token.colorBorderSecondary}`,
                        paddingBlock: 4,
                        whiteSpace: "nowrap",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                      }}
                      title={route}
                    >
                      <ApiOutlined
                        style={{ marginRight: 4, color: token.colorPrimary }}
                      />
                      {route}
                    </div>
                  ))}
                </div>
              </Flex>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
}
