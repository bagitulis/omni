import { type Node, type Edge } from "@xyflow/react";
import {
  forceSimulation,
  forceLink,
  forceManyBody,
  forceCollide,
  forceX,
  forceY,
  type SimulationNodeDatum,
} from "d3-force";
import type { RouteData } from "@/types/routeMapping";
import type { GlobalToken } from "antd";

// --- Types ---

export type GraphNode = Node & {
  data: {
    label: string;
    details?: string;
    status?:
      | "connected"
      | "frontend_only"
      | "backend_only"
      | "unused"
      | "unknown";
    method?: string;
    color?: string;
  };
};

export interface GraphData {
  nodes: GraphNode[];
  edges: Edge[];
}

export interface FilterOptions {
  showConnected: boolean;
  showFrontendOnly: boolean;
  showBackendOnly: boolean;
  showUnused: boolean;
  searchTerm: string;
}

// --- D3 Types ---
interface SimNode extends SimulationNodeDatum {
  id: string;
  type: "component" | "route";
  x?: number;
  y?: number;
}

// --- Constants ---
const COMPONENT_WIDTH = 180;
const ROUTE_WIDTH = 160;

/** Converts any CSS color string to rgba with the given alpha (0–1).
 *  Handles hex (#rrggbb, #rgb) and rgb/rgba(...) formats safely.
 *  Falls back to the original color if parsing fails.
 */
const withAlpha = (color: string, alpha: number): string => {
  const hex6 = /^#([0-9a-f]{6})$/i.exec(color);
  if (hex6) {
    const [, h] = hex6;
    const r = parseInt(h.slice(0, 2), 16);
    const g = parseInt(h.slice(2, 4), 16);
    const b = parseInt(h.slice(4, 6), 16);
    return `rgba(${r},${g},${b},${alpha})`;
  }
  const hex3 = /^#([0-9a-f]{3})$/i.exec(color);
  if (hex3) {
    const [, h] = hex3;
    const r = parseInt(h[0] + h[0], 16);
    const g = parseInt(h[1] + h[1], 16);
    const b = parseInt(h[2] + h[2], 16);
    return `rgba(${r},${g},${b},${alpha})`;
  }
  const rgbMatch = /^rgba?\((\d+),\s*(\d+),\s*(\d+)/.exec(color);
  if (rgbMatch) {
    return `rgba(${rgbMatch[1]},${rgbMatch[2]},${rgbMatch[3]},${alpha})`;
  }
  // Fallback: return the original color unchanged
  return color;
};

// --- Helpers ---

const getRouteKey = (method: string | undefined, endpoint: string) => {
  return `${method || "GET"} ${endpoint}`;
};

const getMethodColor = (method: string = "GET", token: GlobalToken) => {
  switch (method.toUpperCase()) {
    case "GET":
      return token.colorInfo; // Blue
    case "POST":
      return token.colorSuccess; // Green
    case "DELETE":
      return token.colorError; // Red
    case "PUT":
      return token.colorWarning; // Orange
    case "PATCH":
      return "#722ed1"; // Purple (preset)
    default:
      return token.colorTextSecondary;
  }
};

const getStatusColor = (status: string, token: GlobalToken) => {
  switch (status) {
    case "connected":
      return token.colorSuccess;
    case "frontend_only":
      return token.colorWarning;
    case "backend_only":
      return token.colorTextQuaternary;
    case "unused":
      return token.colorTextQuaternary;
    default:
      return token.colorBorder;
  }
};

// --- Main Transformation Logic ---

export const transformDataToGraph = (
  data: RouteData,
  filters: FilterOptions,
  token: GlobalToken,
): GraphData => {
  const nodes: GraphNode[] = [];
  const edges: Edge[] = [];

  // 1. Build Route Status Map
  const routeStatusMap = new Map<
    string,
    "connected" | "frontend_only" | "backend_only" | "unused"
  >();

  data.categories.connected.forEach((r) => {
    routeStatusMap.set(getRouteKey(r.method, r.endpoint), "connected");
  });
  data.categories.frontend_only.forEach((r) => {
    routeStatusMap.set(getRouteKey(r.method, r.endpoint), "frontend_only");
  });
  data.categories.backend_only.forEach((r) => {
    routeStatusMap.set(getRouteKey(r.method, r.endpoint), "backend_only");
  });
  data.categories.unused.forEach((r) => {
    routeStatusMap.set(getRouteKey(r.method, r.endpoint), "unused");
  });

  // 2. Filter & Collect Component Nodes
  const relevantRoutes = new Set<string>();

  Object.entries(data.components).forEach(([name, detail]) => {
    // Search filter
    if (
      filters.searchTerm &&
      !name.toLowerCase().includes(filters.searchTerm.toLowerCase())
    ) {
      return;
    }

    const routesCalled = detail.routes_called || [];

    // Check if component has ANY relevant routes based on filters
    const hasRelevantRoute = routesCalled.some((routeStr) => {
      const status = routeStatusMap.get(routeStr) || "unknown";
      if (status === "connected" && filters.showConnected) return true;
      if (status === "frontend_only" && filters.showFrontendOnly) return true;
      if (status === "backend_only" && filters.showBackendOnly) return true;
      if (status === "unused" && filters.showUnused) return true;
      if (status === "unknown") return false;
      return false;
    });

    if (routesCalled.length === 0) return;

    if (hasRelevantRoute || filters.searchTerm) {
      nodes.push({
        id: name,
        type: "component",
        position: { x: 0, y: 0 },
        data: {
          label: name,
          details: detail.path || "",
        },
        style: {
          width: COMPONENT_WIDTH,
          background: token.colorFillSecondary,
          border: `1px solid ${token.colorBorder}`,
          borderRadius: 3,
          padding: 8,
        },
      });

      routesCalled.forEach((r) => {
        const status = routeStatusMap.get(r);
        if (
          (status === "connected" && filters.showConnected) ||
          (status === "frontend_only" && filters.showFrontendOnly)
        ) {
          relevantRoutes.add(r);
          edges.push({
            id: `${name}-${r}`,
            source: name,
            target: r,
            type: "default",
            style: { stroke: token.colorBorder },
          });
        }
      });
    }
  });

  // 3. Add Backend/Unused Routes
  if (filters.showBackendOnly) {
    data.categories.backend_only.forEach((r) => {
      relevantRoutes.add(getRouteKey(r.method, r.endpoint));
    });
  }
  if (filters.showUnused) {
    data.categories.unused.forEach((r) => {
      relevantRoutes.add(getRouteKey(r.method, r.endpoint));
    });
  }

  // 4. Create Route Nodes
  relevantRoutes.forEach((routeStr) => {
    const parts = routeStr.split(" ");
    const method = parts[0];
    const endpoint = parts.slice(1).join(" ");
    const status = routeStatusMap.get(routeStr) || "unknown";
    const color = getStatusColor(status, token);
    const methodColor = getMethodColor(method, token);

    const bgAlpha = withAlpha(color, 0.1);

    nodes.push({
      id: routeStr,
      type: "route",
      position: { x: 0, y: 0 },
      data: {
        label: endpoint,
        method: method,
        status: status,
        color: color,
        details: methodColor,
      },
      style: {
        width: ROUTE_WIDTH,
        background: bgAlpha,
        borderColor: color,
        borderStyle: "solid",
        borderWidth: 1,
        borderRadius: 3,
        padding: "4px 8px",
      },
    });
  });

  return { nodes, edges };
};

export const runForceLayout = (nodes: GraphNode[], edges: Edge[]) => {
  const simNodes: SimNode[] = nodes.map((n) => ({
    id: n.id,
    type: n.type as "component" | "route",
    x: n.position.x,
    y: n.position.y,
  }));

  const simLinks = edges.map((e) => ({
    source: e.source,
    target: e.target,
  }));

  const simulation = forceSimulation(simNodes)
    .force(
      "link",
      forceLink(simLinks)
        .id((d) => (d as SimNode).id)
        .distance(200),
    )
    .force("charge", forceManyBody().strength(-300))
    .force(
      "collide",
      forceCollide().radius((d) =>
        (d as SimNode).type === "component" ? 100 : 50,
      ),
    )
    .force(
      "x",
      forceX()
        .x((d) => ((d as SimNode).type === "component" ? -300 : 300))
        .strength(0.5),
    )
    .force("y", forceY().strength(0.1));

  simulation.tick(300);

  return nodes.map((node, i) => {
    const simNode = simNodes[i];
    return {
      ...node,
      position: { x: simNode.x || 0, y: simNode.y || 0 },
    };
  });
};
