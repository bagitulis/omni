import {
  useCallback,
  useDeferredValue,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider,
  useReactFlow,
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
import { ComponentNode, RouteNode } from "./GraphNodes";

const { useToken } = theme;
const { Text } = Typography;
const { useBreakpoint } = Grid;

type FilterKey = keyof Omit<FilterOptions, "searchTerm">;
const FILTER_OPTS: { key: FilterKey; label: string }[] = [
  { key: "showConnected", label: "Connected" },
  { key: "showFrontendOnly", label: "Frontend Only" },
  { key: "showBackendOnly", label: "Backend Only" },
  { key: "showUnused", label: "Unused" },
];

interface GraphViewProps {
  data?: RouteData;
}

const GraphContent = ({ data }: GraphViewProps) => {
  const { token } = useToken();
  const screens = useBreakpoint();
  const { fitView } = useReactFlow();

  const [nodes, setNodes, onNodesChange] = useNodesState<GraphNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  const [filters, setFilters] = useState<FilterOptions>({
    showConnected: true,
    showFrontendOnly: true,
    showBackendOnly: false,
    showUnused: false,
    searchTerm: "",
  });

  const deferredSearchTerm = useDeferredValue(filters.searchTerm);
  const deferredFilters = useMemo(
    () => ({ ...filters, searchTerm: deferredSearchTerm }),
    [filters, deferredSearchTerm],
  );

  const nodeTypes = useMemo(
    () => ({ component: ComponentNode, route: RouteNode }),
    [],
  );

  useEffect(() => {
    if (!data) return;
    const { nodes: rawNodes, edges: rawEdges } = transformDataToGraph(
      data,
      deferredFilters,
      token,
    );
    setNodes(runForceLayout(rawNodes, rawEdges));
    setEdges(rawEdges);
    setSelectedNodeId(null);
    const t = setTimeout(() => fitView({ padding: 0.2, duration: 800 }), 50);
    return () => clearTimeout(t);
  }, [data, deferredFilters, token, setNodes, setEdges, fitView]);

  const handleReset = useCallback(() => {
    fitView({ padding: 0.2, duration: 800 });
    setSelectedNodeId(null);
  }, [fitView]);

  const { displayNodes, displayEdges } = useMemo(() => {
    if (!selectedNodeId) return { displayNodes: nodes, displayEdges: edges };

    const connectedNodeIds = new Set<string>([selectedNodeId]);
    const connectedEdgeIds = new Set<string>();
    edges.forEach((e) => {
      if (e.source === selectedNodeId || e.target === selectedNodeId) {
        connectedEdgeIds.add(e.id);
        connectedNodeIds.add(e.source);
        connectedNodeIds.add(e.target);
      }
    });

    return {
      displayNodes: nodes.map((n) => ({
        ...n,
        style: {
          ...n.style,
          opacity: connectedNodeIds.has(n.id) ? 1 : 0.12,
          transition: "opacity 0.2s, box-shadow 0.2s",
          ...(n.id === selectedNodeId
            ? { boxShadow: `0 0 0 2px ${token.colorPrimary}`, borderColor: token.colorPrimary }
            : {}),
        },
      })),
      displayEdges: edges.map((e) => {
        const lit = connectedEdgeIds.has(e.id);
        return {
          ...e,
          animated: lit,
          style: {
            ...e.style,
            opacity: lit ? 1 : 0.05,
            stroke: lit ? token.colorPrimary : e.style?.stroke,
            strokeWidth: lit ? 2 : 1,
          },
        };
      }),
    };
  }, [nodes, edges, selectedNodeId, token]);

  const hintText = selectedNodeId
    ? `Connections for "${selectedNodeId}" · Click node again or canvas to deselect`
    : "Click a node to highlight its connections";

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
      <Card size="small" styles={{ body: { padding: "12px 16px" } }}>
        <Flex gap={12} wrap="wrap" align="center" justify="space-between">
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
            {FILTER_OPTS.map(({ key, label }) => (
              <Checkbox
                key={key}
                checked={filters[key]}
                onChange={(e) =>
                  setFilters((prev) => ({ ...prev, [key]: e.target.checked }))
                }
              >
                {label}
              </Checkbox>
            ))}
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
        <Text type="secondary" style={{ fontSize: 11, display: "block", marginTop: 6 }}>
          {hintText}
        </Text>
      </Card>

      {nodes.length > 300 && (
        <Alert
          type="warning"
          showIcon
          message={`Graph contains ${nodes.length} nodes. Rendering may be slow.`}
          style={{ marginBottom: 0 }}
        />
      )}

      <div
        style={{
          flex: 1,
          border: `1px solid ${token.colorBorderSecondary}`,
          borderRadius: 6,
          overflow: "hidden",
        }}
      >
        <ReactFlow
          nodes={displayNodes}
          edges={displayEdges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={(_, node) =>
            setSelectedNodeId((prev) => (prev === node.id ? null : node.id))
          }
          onPaneClick={() => setSelectedNodeId(null)}
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
