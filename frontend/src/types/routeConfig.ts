/**
 * RouteConfig type — aligned with BE model (RouteExecutionConfig in auto_function.go).
 *
 * The BE returns `is_enabled` (snake_case), but the FE UI fields reference `enabled`.
 * We keep both so that:
 *  - `is_enabled` maps directly to the BE JSON response
 *  - `enabled` is a computed getter used by FE table/form (set from is_enabled during transform)
 *
 * Fields like caching_enabled, queue_enabled etc. are FE-only aspirational fields
 * that the BE does not yet support. They are kept for UI completeness but won't
 * persist until the BE model is extended.
 */
export interface RouteConfig {
  id: number;
  route_key: string;
  route_path: string;
  is_enabled: boolean;
  mode: string;
  priority: string;
  rate_limit: number;
  timeout: number;
  category: string;
  created_at?: string;
  updated_at?: string;

  // FE-only convenience alias (populated from is_enabled during query transform)
  enabled: boolean;

  // FE-only UI fields (not yet supported by BE)
  route_method?: string;
  route_name?: string;
  description?: string;
  caching_enabled?: boolean;
  cache_ttl?: number;
  cache_strategy?: string;
  queue_enabled?: boolean;
  queue_max_size?: number;
  queue_priority?: number;
  max_concurrent?: number;
  rate_limit_enabled?: boolean;
  rate_limit_window?: number;
  rate_limit_max?: number;
  min_interval_ms?: number;
  retry_enabled?: boolean;
  max_retries?: number;
  retry_delay_ms?: number;
  custom_config?: Record<string, unknown> | null;
}
