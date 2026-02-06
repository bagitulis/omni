/**
 * Route Control Types
 * Single Responsibility: Type definitions for route management
 * API types use snake_case to match backend JSON response
 */

export interface RouteConfig {
  id?: string;
  route_path: string;
  route_name?: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  category: string;
  description?: string;
  enabled: boolean;
  cache_config?: CacheConfig;
  queue_config?: QueueConfig;
  rate_limit_config?: RateLimitConfig;
  // Monitoring metrics (optional, populated from API)
  avg_response_time?: number;
  error_rate?: number;
  total_requests?: number;
  created_at?: Date;
  updated_at?: Date;
}

export interface CacheConfig {
  enabled: boolean;
  ttl: number; // seconds
  key?: string;
  invalidate_on?: string[];
}

export interface QueueConfig {
  enabled: boolean;
  priority: "low" | "medium" | "high";
  max_retries: number;
  retry_delay: number; // milliseconds
}

export interface RateLimitConfig {
  enabled: boolean;
  max_requests: number;
  window_ms: number;
}

export interface RouteCategory {
  name: string;
  count: number;
  color?: string;
}

export interface RouteMetrics {
  route_id: string;
  total_calls: number;
  success_count: number;
  error_count: number;
  avg_response_time: number;
  last_called?: Date;
}

export type PresetType = "aggressive" | "balanced" | "minimal" | "realtime";
