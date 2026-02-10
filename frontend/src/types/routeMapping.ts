export interface RouteSummary {
  endpoint: string;
  method?: string;
  is_dynamic?: boolean;
  category?: string;
}

export interface FrontendOnlySummary {
  endpoint: string;
  components: string[];
  status?: string;
}

export interface RouteCategoriesUi {
  connected: RouteSummary[];
  frontend_only: FrontendOnlySummary[];
  backend_only: RouteSummary[];
  unused: RouteSummary[];
}

export interface ComponentDetail {
  path?: string;
  routes_called?: string[];
  buttons?: string[];
}

export interface RouteCategoryLabel {
  title: string;
  detail: string;
}

export interface RouteData {
  total_routes: number;
  total_components: number;
  total_categories: number;
  total_dynamic_routes: number;
  total_called_routes: number;
  total_disconnected_routes: number;
  total_unused_routes: number;
  connection_rate: string;
  by_category: Record<string, RouteSummary[]>;
  categories: RouteCategoriesUi;
  category_labels: Record<string, RouteCategoryLabel>;
  category_stats: Record<string, number>;
  components: Record<string, ComponentDetail>;
  disconnected_routes: Record<string, any>;
  backend_only_routes: Record<string, any>;
  unused_routes: Record<string, any>;
  button_to_endpoints: Record<string, Record<string, string[]>>;
  statistics?: Record<string, any>;
  timestamp: string;
}

export interface RouteStatistics {
  statistics: Record<string, any>;
}

export type ViewMode =
  | "categories"
  | "component"
  | "disconnected"
  | "backend"
  | "unused";
