import {
  Card,
  Typography,
  Tabs,
  Input,
  Table,
  Tag,
  Statistic,
  Row,
  Col,
  Button,
  Spin,
  Alert,
  Tooltip,
  Empty,
} from "antd";
import {
  SearchOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  DisconnectOutlined,
  ApiOutlined,
  AppstoreOutlined,
} from "@ant-design/icons";
import { useRouteMapping } from "@/hooks/useRouteMapping";
import { ViewMode } from "@/types/routeMapping";

const { Title, Text } = Typography;

export function RouteMappingPage() {
  const {
    data,
    isLoading,
    error,
    refetch,
    isFetching,
    viewMode,
    setViewMode,
    searchQuery,
    setSearchQuery,
    categoryStats,
    filteredComponents,
    filteredLists,
  } = useRouteMapping();

  const handleTabChange = (key: string) => {
    setViewMode(key as ViewMode);
  };

  const columnsConnected = [
    {
      title: "Endpoint",
      dataIndex: "endpoint",
      key: "endpoint",
      render: (text: string) => (
        <Text code copyable>
          {text}
        </Text>
      ),
    },
    {
      title: "Method",
      dataIndex: "method",
      key: "method",
      width: 100,
      render: (method: string) => (
        <Tag
          color={
            method === "GET" ? "blue" : method === "POST" ? "green" : "orange"
          }
        >
          {method}
        </Tag>
      ),
    },
    {
      title: "Category",
      dataIndex: "category",
      key: "category",
      render: (cat: string) => <Tag>{cat}</Tag>,
    },
    {
      title: "Type",
      dataIndex: "is_dynamic",
      key: "is_dynamic",
      width: 100,
      render: (dynamic: boolean) => (
        <Tag color={dynamic ? "purple" : "cyan"}>
          {dynamic ? "Dynamic" : "Static"}
        </Tag>
      ),
    },
  ];

  const columnsFrontendOnly = [
    {
      title: "Endpoint",
      dataIndex: "endpoint",
      key: "endpoint",
      render: (text: string) => (
        <Text code copyable>
          {text}
        </Text>
      ),
    },
    {
      title: "Used In Components",
      dataIndex: "components",
      key: "components",
      render: (components: string[]) => (
        <>
          {components.map((c) => (
            <Tag key={c} color="geekblue">
              {c}
            </Tag>
          ))}
        </>
      ),
    },
    {
      title: "Status",
      key: "status",
      render: () => (
        <Tag color="error" icon={<DisconnectOutlined />}>
          Disconnected
        </Tag>
      ),
    },
  ];

  const columnsBackendOnly = [
    {
      title: "Endpoint",
      dataIndex: "endpoint",
      key: "endpoint",
      render: (text: string) => (
        <Text code copyable>
          {text}
        </Text>
      ),
    },
    {
      title: "Method",
      dataIndex: "method",
      key: "method",
      render: (method: string) => <Tag>{method}</Tag>,
    },
    {
      title: "Category",
      dataIndex: "category",
      key: "category",
      render: (cat: string) => <Tag>{cat}</Tag>,
    },
  ];

  const columnsUnused = [
    {
      title: "Endpoint",
      dataIndex: "endpoint",
      key: "endpoint",
      render: (text: string) => (
        <Text code copyable>
          {text}
        </Text>
      ),
    },
    {
      title: "Category",
      dataIndex: "category",
      key: "category",
      render: (cat: string) => <Tag>{cat}</Tag>,
    },
  ];

  if (error) {
    return (
      <div className="p-6">
        <Alert
          message="Error Loading Route Mapping"
          description={(error as Error).message}
          type="error"
          showIcon
          action={
            <Button size="small" onClick={() => refetch()}>
              Retry
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <Title level={2} style={{ margin: 0 }}>
            Route Mapping
          </Title>
          <Text type="secondary">
            Visualize and analyze API route connections
          </Text>
        </div>
        <div className="flex gap-2">
          <Input
            placeholder="Search routes or components..."
            prefix={<SearchOutlined />}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ width: 300 }}
            allowClear
          />
          <Tooltip title="Refresh Data">
            <Button
              icon={<ReloadOutlined spin={isFetching} />}
              onClick={() => refetch()}
              loading={isFetching}
            />
          </Tooltip>
        </div>
      </div>

      {isLoading && !data ? (
        <div className="flex justify-center p-12">
          <Spin size="large" tip="Analyzing Routes..." />
        </div>
      ) : (
        <>
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
              <Card bordered={false} bodyStyle={{ padding: "16px" }}>
                <Statistic
                  title="Total Routes"
                  value={data?.total_routes || 0}
                  prefix={<ApiOutlined style={{ color: "#1890ff" }} />}
                  valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
              <Card bordered={false} bodyStyle={{ padding: "16px" }}>
                <Statistic
                  title="Connection Rate"
                  value={data?.connection_rate || "0%"}
                  prefix={<CheckCircleOutlined style={{ color: "#52c41a" }} />}
                  valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
              <Card bordered={false} bodyStyle={{ padding: "16px" }}>
                <Statistic
                  title="Frontend Only"
                  value={categoryStats.frontend_only}
                  prefix={<DisconnectOutlined style={{ color: "#ff4d4f" }} />}
                  valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
              <Card bordered={false} bodyStyle={{ padding: "16px" }}>
                <Statistic
                  title="Backend Only"
                  value={categoryStats.backend_only}
                  prefix={<ApiOutlined style={{ color: "#faad14" }} />}
                  valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
              <Card bordered={false} bodyStyle={{ padding: "16px" }}>
                <Statistic
                  title="Components"
                  value={data?.total_components || 0}
                  prefix={<AppstoreOutlined style={{ color: "#722ed1" }} />}
                  valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
                />
              </Card>
            </Col>
          </Row>

          <Card
            bordered={false}
            bodyStyle={{ padding: "0" }}
            className="overflow-hidden"
          >
            <Tabs
              activeKey={viewMode}
              onChange={handleTabChange}
              type="card"
              size="large"
              tabBarStyle={{
                margin: 0,
                padding: "8px 8px 0 8px",
                background: "#fafafa",
              }}
              items={[
                {
                  key: "categories",
                  label: `Connected (${categoryStats.connected})`,
                  children: (
                    <Table
                      dataSource={filteredLists.connected}
                      columns={columnsConnected}
                      rowKey="endpoint"
                      pagination={{ pageSize: 20 }}
                      size="middle"
                    />
                  ),
                },
                {
                  key: "component",
                  label: `Components (${Object.keys(filteredComponents).length})`,
                  children: (
                    <div className="p-4 bg-gray-50">
                      {Object.keys(filteredComponents).length === 0 ? (
                        <Empty />
                      ) : (
                        <Row gutter={[16, 16]}>
                          {Object.entries(filteredComponents).map(
                            ([name, detail]) => (
                              <Col span={8} key={name}>
                                <Card
                                  title={name}
                                  size="small"
                                  extra={
                                    <Tag>
                                      {detail.routes_called?.length || 0} Calls
                                    </Tag>
                                  }
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
                                      {detail.routes_called?.map(
                                        (route: string) => (
                                          <div
                                            key={route}
                                            className="text-xs border-b border-gray-100 py-1 last:border-0 truncate"
                                            title={route}
                                          >
                                            <ApiOutlined className="mr-1 text-blue-500" />
                                            {route}
                                          </div>
                                        ),
                                      )}
                                    </div>
                                  </div>
                                </Card>
                              </Col>
                            ),
                          )}
                        </Row>
                      )}
                    </div>
                  ),
                },
                {
                  key: "disconnected",
                  label: (
                    <span className="text-red-500">
                      Frontend Only ({categoryStats.frontend_only})
                    </span>
                  ),
                  children: (
                    <Table
                      dataSource={filteredLists.frontendOnly}
                      columns={columnsFrontendOnly}
                      rowKey="endpoint"
                      pagination={{ pageSize: 20 }}
                      size="middle"
                    />
                  ),
                },
                {
                  key: "backend",
                  label: `Backend Only (${categoryStats.backend_only})`,
                  children: (
                    <Table
                      dataSource={filteredLists.backendOnly}
                      columns={columnsBackendOnly}
                      rowKey="endpoint"
                      pagination={{ pageSize: 20 }}
                      size="middle"
                    />
                  ),
                },
                {
                  key: "unused",
                  label: `Unused (${categoryStats.unused})`,
                  children: (
                    <Table
                      dataSource={filteredLists.unused}
                      columns={columnsUnused}
                      rowKey="endpoint"
                      pagination={{ pageSize: 20 }}
                      size="middle"
                    />
                  ),
                },
              ]}
            />
          </Card>
        </>
      )}
    </div>
  );
}
