import { type Node, type Edge } from "@xyflow/react";
import type { RouteData } from "@/types/routeMapping";
import type { GlobalToken } from "antd";

// Re-export layout function so existing consumers don't break
export { runForceLayout } from "./graphLayout";

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

// --- Constants ---
const COMPONENT_WIDTH = 180;
const ROUTE_WIDTH = 160;

// --- Color Helpers ---

/** Converts any CSS color string to rgba with the given alpha (0–1). */
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
  return color;
};

const getRouteKey = (method: string | undefined, endpoint: string) =>
  `${method || "GET"} ${endpoint}`;

const METHOD_COLOR_MAP: Record<string, keyof GlobalToken> = {
  GET: "colorInfo",
  POST: "colorSuccess",
  DELETE: "colorError",
  PUT: "colorWarning",
};

const getMethodColor = (method: string = "GET", token: GlobalToken) => {
  const upper = method.toUpperCase();
  if (upper === "PATCH") return "#722ed1";
  const key = METHOD_COLOR_MAP[upper];
  return key ? (token[key] as string) : (token.colorTextSecondary as string);
};

const STATUS_COLOR_MAP: Record<string, keyof GlobalToken> = {
  connected: "colorSuccess",
  frontend_only: "colorWarning",
  backend_only: "colorTextQuaternary",
  unused: "colorTextQuaternary",
};

const getStatusColor = (status: string, token: GlobalToken) => {
  const key = STATUS_COLOR_MAP[status];
  return key ? (token[key] as string) : (token.colorBorder as string);
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

  const statusEntries: [
    keyof RouteData["categories"],
    "connected" | "frontend_only" | "backend_only" | "unused",
  ][] = [
    ["connected", "connected"],
    ["frontend_only", "frontend_only"],
    ["backend_only", "backend_only"],
    ["unused", "unused"],
  ];

  for (const [cat, status] of statusEntries) {
    data.categories[cat].forEach((r) => {
      routeStatusMap.set(getRouteKey(r.method, r.endpoint), status);
    });
  }

  // 2. Filter & Collect Component Nodes
  const relevantRoutes = new Set<string>();

  Object.entries(data.components).forEach(([name, detail]) => {
    if (
      filters.searchTerm &&
      !name.toLowerCase().includes(filters.searchTerm.toLowerCase())
    )
      return;

    const routesCalled = detail.routes_called || [];
    if (routesCalled.length === 0) return;

    const hasRelevantRoute = routesCalled.some((routeStr) => {
      const status = routeStatusMap.get(routeStr) || "unknown";
      if (status === "connected" && filters.showConnected) return true;
      if (status === "frontend_only" && filters.showFrontendOnly) return true;
      if (status === "backend_only" && filters.showBackendOnly) return true;
      if (status === "unused" && filters.showUnused) return true;
      return false;
    });

    if (hasRelevantRoute || filters.searchTerm) {
      nodes.push({
        id: name,
        type: "component",
        position: { x: 0, y: 0 },
        data: { label: name, details: detail.path || "" },
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
    data.categories.backend_only.forEach((r) =>
      relevantRoutes.add(getRouteKey(r.method, r.endpoint)),
    );
  }
  if (filters.showUnused) {
    data.categories.unused.forEach((r) =>
      relevantRoutes.add(getRouteKey(r.method, r.endpoint)),
    );
  }

  // 4. Create Route Nodes
  relevantRoutes.forEach((routeStr) => {
    const parts = routeStr.split(" ");
    const method = parts[0];
    const endpoint = parts.slice(1).join(" ");
    const status = routeStatusMap.get(routeStr) || "unknown";
    const color = getStatusColor(status, token);
    const methodColor = getMethodColor(method, token);

    nodes.push({
      id: routeStr,
      type: "route",
      position: { x: 0, y: 0 },
      data: { label: endpoint, method, status, color, details: methodColor },
      style: {
        width: ROUTE_WIDTH,
        background: withAlpha(color, 0.1),
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
