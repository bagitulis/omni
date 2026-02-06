import { computed, ref, type Ref } from "vue";

type RouteFlow = "queued" | "direct";
type FlowStatus = "pending" | "success" | "error";

interface FlowEntry {
  id: string;
  routeName: string;
  path: string;
  flow: RouteFlow;
  status: FlowStatus;
  startedAt: number;
  durationMs?: number;
  error?: string | null;
}

const STORAGE_KEY = "route_flow_direct_routes";
const DEFAULT_DIRECT_ROUTES = [
  "order-manager",
  "OrderManager",
  "OrderManagerShopee",
  "OrderManagerLazada",
  "OrderManagerTiktok",
  "/order-manager",
  "inventory",
  "Inventory",
  "/inventory",
  // Script Monitor routes (instant navigation)
  "script-monitor",
  "ScriptMonitor",
  "ScriptMonitorCurrent",
  "ScriptMonitorQueue",
  "ScriptMonitorHistory",
  "ScriptMonitorAutoFunctions",
  "/script-monitor",
  // Report routes (instant navigation)
  "report",
  "Report",
  "ShopeeReport",
  "TiktokReport",
  "/report",
  // Analytics routes (instant navigation)
  "analytics",
  "Analytics",
  "TiktokAdsAnalytics",
  "ShopeeAdsAnalytics",
  "/analytics",
];

const directRoutes: Ref<string[]> = ref(loadPersistedDirectRoutes());
const flowLog: Ref<FlowEntry[]> = ref([]);

function loadPersistedDirectRoutes(): string[] {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      if (Array.isArray(parsed) && parsed.length > 0) {
        // Merge stored routes with defaults to ensure new defaults are included
        const merged = new Set([...DEFAULT_DIRECT_ROUTES, ...parsed]);
        return Array.from(merged);
      }
    }
  } catch (error) {
    console.warn("Failed to load direct route overrides", error);
  }
  return [...DEFAULT_DIRECT_ROUTES];
}

function persistDirectRoutes(): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(directRoutes.value));
  } catch (error) {
    console.warn("Failed to persist direct routes", error);
  }
}

function normalize(value?: string): string {
  return (value || "").trim().toLowerCase();
}

function isDirectRoute(routeName?: string, path?: string): boolean {
  const normalizedName = normalize(routeName);
  const normalizedPath = normalize(path);
  return directRoutes.value.some((entry) => {
    const normalizedEntry = normalize(entry);
    if (!normalizedEntry) return false;
    if (normalizedName && normalizedName === normalizedEntry) return true;
    return normalizedPath.startsWith(normalizedEntry);
  });
}

function getFlowType(routeName?: string, path?: string): RouteFlow {
  return isDirectRoute(routeName, path) ? "direct" : "queued";
}

function addDirectRoute(value: string): void {
  const cleaned = value.trim();
  if (!cleaned) return;
  if (!directRoutes.value.includes(cleaned)) {
    directRoutes.value.push(cleaned);
    persistDirectRoutes();
  }
}

function removeDirectRoute(value: string): void {
  directRoutes.value = directRoutes.value.filter((route) => route !== value);
  persistDirectRoutes();
}

function resetDirectRoutes(): void {
  directRoutes.value = [...DEFAULT_DIRECT_ROUTES];
  persistDirectRoutes();
}

function recordFlowStart(routeName: string, path: string, flow: RouteFlow): string {
  const entry: FlowEntry = {
    id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    routeName,
    path,
    flow,
    status: "pending",
    startedAt: Date.now(),
  };

  flowLog.value = [entry, ...flowLog.value].slice(0, 50);
  return entry.id;
}

function completeFlow(
  flowId: string,
  success: boolean,
  durationMs: number,
  error: string | null = null
): void {
  flowLog.value = flowLog.value.map((entry) =>
    entry.id === flowId
      ? {
          ...entry,
          status: success ? "success" : "error",
          durationMs,
          error,
        }
      : entry
  );
}

export function useRouteFlowController() {
  return {
    directRoutes,
    recentFlows: computed(() => flowLog.value.slice(0, 10)),
    flows: computed(() => flowLog.value),
    defaultDirectRoutes: DEFAULT_DIRECT_ROUTES,
    isDirectRoute,
    getFlowType,
    addDirectRoute,
    removeDirectRoute,
    resetDirectRoutes,
    recordFlowStart,
    completeFlow,
  };
}
