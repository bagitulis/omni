export interface RouteConfig {
  id: number;
  route_path: string;
  route_method: string;
  route_name: string;
  description: string;
  category: string;
  enabled: boolean;
  caching_enabled: boolean;
  cache_ttl: number;
  cache_strategy: string;
  queue_enabled: boolean;
  queue_max_size: number;
  queue_priority: number;
  max_concurrent: number;
  rate_limit_enabled: boolean;
  rate_limit_window: number;
  rate_limit_max: number;
  min_interval_ms: number;
  timeout: number;
  retry_enabled: boolean;
  max_retries: number;
  retry_delay_ms: number;
  custom_config: Record<string, unknown> | null;
  created_at?: string;
  updated_at?: string;
}
