import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider,
  useReactFlow,
  Handle,
  Position,
  type NodeProps,
  type Edge,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import {
  Alert,
  Button,
  Checkbox,
  Flex,
  Input,
  Space,
  theme,
  Typography,
  Tag,
  Card,
  Grid,
} from "antd";
import { ReloadOutlined, SearchOutlined } from "@ant-design/icons";
import type { RouteData } from "@/types/routeMapping";
import {
  transformDataToGraph,
  runForceLayout,
  type FilterOptions,
  type GraphNode,
} from "./GraphViewHelpers";

const { useToken } = theme;
const { Text } = Typography;
const { useBreakpoint } = Grid;

// --- Custom Nodes ---

const ComponentNode = ({ data }: NodeProps<GraphNode>) => {
  const { token } = useToken();
  return (
    <div
      style={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
      }}
    >
      <Text strong style={{ fontSize: 12 }}>
        {data.label}
      </Text>
      {data.details && (
        <Text type="secondary" style={{ fontSize: 10 }} ellipsis>
          {data.details}
        </Text>
      )}
      <Handle
        type="source"
        position={Position.Right}
        style={{ background: token.colorBorder }}
      />
    </div>
  );
};

const RouteNode = ({ data }: NodeProps<GraphNode>) => {
  return (
    <div
      style={{ height: "100%", display: "flex", alignItems: "center", gap: 4 }}
    >
      <Handle
        type="target"
        position={Position.Left}
        style={{ background: data.color }}
      />
      <Tag
        color={data.details}
        style={{
          margin: 0,
          fontSize: 10,
          padding: "0 4px",
          lineHeight: "16px",
        }}
      >
        {data.method}
      </Tag>
      <Text style={{ fontSize: 11, color: data.color }} ellipsis>
        {data.label}
      </Text>
    </div>
  );
};

// --- Main Component ---

interface GraphViewProps {
  data?: RouteData;
}

const GraphContent = ({ data }: GraphViewProps) => {
  const { token } = useToken();
  const screens = useBreakpoint();
  const { fitView } = useReactFlow();

  const [nodes, setNodes, onNodesChange] = useNodesState<GraphNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  const [filters, setFilters] = useState<FilterOptions>({
    showConnected: true,
    showFrontendOnly: true,
    showBackendOnly: false,
    showUnused: false,
    searchTerm: "",
  });

  const nodeTypes = useMemo(
    () => ({
      component: ComponentNode,
      route: RouteNode,
    }),
    [],
  );

  // Update Graph when data or filters change
  useEffect(() => {
    if (!data) return;

    // Transform
    const { nodes: rawNodes, edges: rawEdges } = transformDataToGraph(
      data,
      filters,
      token,
    );

    // Layout
    const layoutNodes = runForceLayout(rawNodes, rawEdges);

    setNodes(layoutNodes);
    setEdges(rawEdges);

    // Fit view after small delay to allow render; clean up on re-run
    const timerId = setTimeout(
      () => fitView({ padding: 0.2, duration: 800 }),
      50,
    );
    return () => clearTimeout(timerId);
  }, [data, filters, token, setNodes, setEdges, fitView]);

  const handleReset = useCallback(() => {
    fitView({ padding: 0.2, duration: 800 });
  }, [fitView]);

  const showWarning = nodes.length > 300;

  return (
    <div
      style={{
        height: "calc(100vh - 300px)",
        minHeight: 500,
        display: "flex",
        flexDirection: "column",
        gap: 16,
      }}
    >
      {/* Toolbar */}
      <Card size="small" styles={{ body: { padding: "12px 16px" } }}>
        <Flex gap={16} wrap="wrap" align="center" justify="space-between">
          <Space wrap>
            <Input
              placeholder="Filter components..."
              prefix={<SearchOutlined />}
              value={filters.searchTerm}
              onChange={(e) =>
                setFilters((prev) => ({ ...prev, searchTerm: e.target.value }))
              }
              style={{ width: 200 }}
              allowClear
            />
            <Checkbox
              checked={filters.showConnected}
              onChange={(e) =>
                setFilters((prev) => ({
                  ...prev,
                  showConnected: e.target.checked,
                }))
              }
            >
              Connected
            </Checkbox>
            <Checkbox
              checked={filters.showFrontendOnly}
              onChange={(e) =>
                setFilters((prev) => ({
                  ...prev,
                  showFrontendOnly: e.target.checked,
                }))
              }
            >
              Frontend Only
            </Checkbox>
            <Checkbox
              checked={filters.showBackendOnly}
              onChange={(e) =>
                setFilters((prev) => ({
                  ...prev,
                  showBackendOnly: e.target.checked,
                }))
              }
            >
              Backend Only
            </Checkbox>
            <Checkbox
              checked={filters.showUnused}
              onChange={(e) =>
                setFilters((prev) => ({
                  ...prev,
                  showUnused: e.target.checked,
                }))
              }
            >
              Unused
            </Checkbox>
          </Space>

          <Space>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {nodes.filter((n) => n.type === "component").length} comp,{" "}
              {nodes.filter((n) => n.type === "route").length} routes
            </Text>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              Reset View
            </Button>
          </Space>
        </Flex>
      </Card>

      {/* Warning */}
      {showWarning && (
        <Alert
          type="warning"
          showIcon
          message={`Graph contains ${nodes.length} nodes. Rendering may be slow.`}
          style={{ marginBottom: 0 }}
        />
      )}

      {/* Graph Area */}
      <div
        style={{
          flex: 1,
          border: `1px solid ${token.colorBorderSecondary}`,
          borderRadius: 6,
          overflow: "hidden",
        }}
      >
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          fitView
          minZoom={0.1}
          maxZoom={1.5}
        >
          <Background color={token.colorFill} gap={16} />
          <Controls />
          {screens.md && <MiniMap style={{ height: 120 }} zoomable pannable />}
        </ReactFlow>
      </div>
    </div>
  );
};

export const GraphView = (props: GraphViewProps) => (
  <ReactFlowProvider>
    <GraphContent {...props} />
  </ReactFlowProvider>
);
