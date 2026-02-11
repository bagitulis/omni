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
} from "antd";
import { useRouteMapping } from "@/hooks/useRouteMapping";
import type { ViewMode } from "@/types/routeMapping";
import {
  columnsBackendOnly,
  columnsConnected,
  columnsFrontendOnly,
  columnsUnused,
} from "./columns";
import { ComponentsList } from "./components/ComponentsList";
import { RouteStats } from "./components/RouteStats";

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

  const renderTable = <T extends object>(
    dataSource: T[],
    columns: TableProps<T>["columns"],
  ) => (
    <Table
      dataSource={dataSource}
      columns={columns}
      rowKey="endpoint"
      pagination={{ pageSize: 20 }}
      size="middle"
    />
  );

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
          <RouteStats data={data} categoryStats={categoryStats} />

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
                  children: renderTable(
                    filteredLists.connected,
                    columnsConnected,
                  ),
                },
                {
                  key: "component",
                  label: `Components (${Object.keys(filteredComponents).length})`,
                  children: (
                    <ComponentsList filteredComponents={filteredComponents} />
                  ),
                },
                {
                  key: "disconnected",
                  label: (
                    <span className="text-red-500">
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
              ]}
            />
          </Card>
        </>
      )}
    </div>
  );
}
