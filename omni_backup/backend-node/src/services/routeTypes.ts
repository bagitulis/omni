/**
 * Route Types - Shared type definitions for route mapping
 */

export interface RouteInfo {
  endpoint: string;
  method: string;
  isDynamic?: boolean;
  category: string;
  source?: string;
}

export interface ComponentInfo {
  name: string;
  category: string;
  path: string;
  routesCalled: string[];
  buttons: string[];
}

export interface FrontendOnlyRoute {
  endpoint: string;
  components: string[];
  status: string;
}

export interface RouteCategories {
  connected: RouteInfo[];
  frontendOnly: FrontendOnlyRoute[];
  backendOnly: RouteInfo[];
  unused: RouteInfo[];
}

export type RouteCategoryKey =
  | "connected"
  | "frontend_only"
  | "backend_only"
  | "unused";

export interface AnalysisResult {
  routes: Map<string, RouteInfo>;
  components: Map<string, ComponentInfo>;
  connectedRoutes: Set<string>;
  disconnectedCalls: Map<string, string[]>;
  unusedRoutes: Map<string, RouteInfo>;
  categories: RouteCategories;
}
