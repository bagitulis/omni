/**
 * Route Control Types
 * Single Responsibility: Type definitions for route management
 */

export interface RouteConfig {
  id?: string;
  routePath: string;
  routeName?: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  category: string;
  description?: string;
  enabled: boolean;
  cacheConfig?: CacheConfig;
  queueConfig?: QueueConfig;
  rateLimitConfig?: RateLimitConfig;
  // Monitoring metrics (optional, populated from API)
  avgResponseTime?: number;
  errorRate?: number;
  totalRequests?: number;
  createdAt?: Date;
  updatedAt?: Date;
}

export interface CacheConfig {
  enabled: boolean;
  ttl: number; // seconds
  key?: string;
  invalidateOn?: string[];
}

export interface QueueConfig {
  enabled: boolean;
  priority: "low" | "medium" | "high";
  maxRetries: number;
  retryDelay: number; // milliseconds
}

export interface RateLimitConfig {
  enabled: boolean;
  maxRequests: number;
  windowMs: number;
}

export interface RouteCategory {
  name: string;
  count: number;
  color?: string;
}

export interface RouteMetrics {
  routeId: string;
  totalCalls: number;
  successCount: number;
  errorCount: number;
  avgResponseTime: number;
  lastCalled?: Date;
}

export type PresetType = "aggressive" | "balanced" | "minimal" | "realtime";
