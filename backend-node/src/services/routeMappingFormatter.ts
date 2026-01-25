import {
  ComponentInfo,
  FrontendOnlyRoute,
  RouteCategories,
  RouteCategoryKey,
  RouteInfo,
} from "./routeTypes";

export const CATEGORY_LABELS: Record<
  RouteCategoryKey,
  { title: string; detail: string }
> = {
  connected: {
    title: "Connected (Frontend & Backend)",
    detail: "Routes that exist in backend and are called from the frontend.",
  },
  frontend_only: {
    title: "Frontend Calls Missing Backend",
    detail: "Frontend calls with no matching backend implementation.",
  },
  backend_only: {
    title: "Backend-Only (Internal)",
    detail: "Backend routes not called by frontend but likely used internally.",
  },
  unused: {
    title: "Unused Routes",
    detail: "Routes that are not connected to any known caller.",
  },
};

export function formatComponents(
  components: Map<string, ComponentInfo>
): Record<string, any> {
  const formatted: Record<string, any> = {};

  components.forEach((comp) => {
    if (!comp.routesCalled || comp.routesCalled.length === 0) {
      return;
    }

    formatted[comp.name] = {
      name: comp.name,
      category: comp.category || "general",
      path: comp.path || "",
      routes_called: comp.routesCalled,
      buttons: comp.buttons || [],
      total_routes: comp.routesCalled.length,
    };
  });

  return formatted;
}

export function formatFrontendOnly(
  routes: FrontendOnlyRoute[]
): Record<string, any> {
  const formatted: Record<string, any> = {};

  routes.forEach((route) => {
    formatted[route.endpoint] = {
      endpoint: route.endpoint,
      components: route.components,
      status: route.status || "NOT_IMPLEMENTED",
    };
  });

  return formatted;
}

export function formatRouteArray(
  routes: RouteInfo[],
  status: string
): Record<string, any> {
  const formatted: Record<string, any> = {};

  routes.forEach((route) => {
    const key = `${route.method} ${route.endpoint}`;
    formatted[key] = {
      endpoint: route.endpoint,
      method: route.method,
      category: route.category,
      is_dynamic: Boolean(route.isDynamic),
      status,
    };
  });

  return formatted;
}

export function toCategoryRecord(
  categories: RouteCategories
): Record<string, RouteInfo[] | FrontendOnlyRoute[]> {
  return {
    connected: categories.connected,
    frontend_only: categories.frontendOnly,
    backend_only: categories.backendOnly,
    unused: categories.unused,
  };
}

export function normalizeCategoryKey(
  category?: string
): RouteCategoryKey | null {
  if (!category) {
    return null;
  }

  const normalized = category.toLowerCase().replace(/-/g, "_");
  if (normalized === "connected") return "connected";
  if (normalized === "frontend_only") return "frontend_only";
  if (normalized === "backend_only") return "backend_only";
  if (normalized === "unused") return "unused";
  return null;
}
