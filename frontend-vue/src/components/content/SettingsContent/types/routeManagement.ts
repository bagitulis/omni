/**
 * Route Management Types
 * Centralized type definitions
 * API types use snake_case to match backend JSON response
 */

export interface RouteConfig {
  id?: string;
  route_path: string;
  route_method: string;
  route_name?: string;
  description?: string;
  category?: string;
  enabled: boolean;
  caching_enabled: boolean;
  cache_ttl: number;
  cache_strategy: string;
  queue_enabled: boolean;
  queue_max_size: number;
  queue_priority: string;
  max_concurrent: number;
  rate_limit_enabled: boolean;
  rate_limit_window: number;
  rate_limit_max: number;
  min_interval_ms: number;
  timeout: number;
  retry_enabled: boolean;
  max_retries: number;
  retry_delay_ms: number;
  custom_config?: string;
}

export type PresetType = "aggressive" | "balanced" | "minimal" | "realtime";
