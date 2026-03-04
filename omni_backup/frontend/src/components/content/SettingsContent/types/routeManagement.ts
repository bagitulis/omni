/**
 * Route Management Types
 * Centralized type definitions
 */

export interface RouteConfig {
  id?: string;
  routePath: string;
  routeMethod: string;
  routeName?: string;
  description?: string;
  category?: string;
  enabled: boolean;
  cachingEnabled: boolean;
  cacheTTL: number;
  cacheStrategy: string;
  queueEnabled: boolean;
  queueMaxSize: number;
  queuePriority: string;
  maxConcurrent: number;
  rateLimitEnabled: boolean;
  rateLimitWindow: number;
  rateLimitMax: number;
  minIntervalMs: number;
  timeout: number;
  retryEnabled: boolean;
  maxRetries: number;
  retryDelayMs: number;
  customConfig?: string;
}

export type PresetType = 'aggressive' | 'balanced' | 'minimal' | 'realtime';
