import { ReloadOutlined, SearchOutlined } from "@ant-design/icons";
import {
  Alert,
  Button,
  Card,
  Input,
  Spin,
  Table,
  type TableProps,
  Tabs,
  Tooltip,
  Typography,
  theme,
  Flex,
} from "antd";
import { useRouteMapping } from "@/hooks/useRouteMapping";
import type { ViewMode, ComponentDetail } from "@/types/routeMapping";
import {
  columnsBackendOnly,
  columnsConnected,
  columnsFrontendOnly,
  columnsUnused,
} from "./columns";
import { ComponentsList } from "./components/ComponentsList";
import { RouteStats } from "./components/RouteStats";
import { GraphView } from "./components/GraphView";

const { Title, Text } = Typography;

export function RouteMappingPage() {
  const { token } = theme.useToken();
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

  if (error) {
    return (
      <div style={{ padding: 24 }}>
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

  const renderTable = <T extends object>(
    dataSource: T[],
    columns: TableProps<T>["columns"],
  ) => (
    <Table
      dataSource={dataSource}
      columns={columns}
      rowKey={(record) => {
        const row = record as { endpoint?: string; method?: string };
        return `${row.method || "UNKNOWN"}:${row.endpoint || ""}`;
      }}
      pagination={{ pageSize: 20 }}
      size="middle"
    />
  );

  return (
    <div style={{ padding: 24 }}>
      <Flex vertical gap={24}>
        <Flex justify="space-between" align="center">
          <div>
            <Title level={2} style={{ margin: 0 }}>
              Route Mapping
            </Title>
            <Text type="secondary">
              Visualize and analyze API route connections
            </Text>
          </div>
          <Flex gap={8}>
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
          </Flex>
        </Flex>

        {isLoading && !data ? (
          <Flex justify="center" style={{ padding: 48 }}>
            <Spin size="large" tip="Analyzing Routes..." />
          </Flex>
        ) : (
          <>
            <RouteStats data={data} categoryStats={categoryStats} />

            <Card
              bordered={false}
              bodyStyle={{ padding: "0" }}
              style={{ overflow: "hidden" }}
            >
              <Tabs
                activeKey={viewMode}
                onChange={handleTabChange}
                type="card"
                size="large"
                tabBarStyle={{
                  margin: 0,
                  padding: "8px 8px 0 8px",
                  background: token.colorFillQuaternary,
                }}
                items={[
                  {
                    key: "categories",
                    label: `Connected (${categoryStats.connected})`,
                    children: renderTable(
                      filteredLists.connected,
                      columnsConnected,
                    ),
                  },
                  {
                    key: "component",
                    label: `Components (${Object.keys(filteredComponents).length})`,
                    children: (
                      <ComponentsList
                        filteredComponents={
                          filteredComponents as Record<string, ComponentDetail>
                        }
                      />
                    ),
                  },
                  {
                    key: "disconnected",
                    label: (
                      <span style={{ color: token.colorError }}>
                        Frontend Only ({categoryStats.frontend_only})
                      </span>
                    ),
                    children: renderTable(
                      filteredLists.frontendOnly,
                      columnsFrontendOnly,
                    ),
                  },
                  {
                    key: "backend",
                    label: `Backend Only (${categoryStats.backend_only})`,
                    children: renderTable(
                      filteredLists.backendOnly,
                      columnsBackendOnly,
                    ),
                  },
                  {
                    key: "unused",
                    label: `Unused (${categoryStats.unused})`,
                    children: renderTable(filteredLists.unused, columnsUnused),
                  },
                  {
                    key: "graph",
                    label: "Graph View",
                    children: <GraphView data={data} />,
                  },
                ]}
              />
            </Card>
          </>
        )}
      </Flex>
    </div>
  );
}

export default RouteMappingPage;
