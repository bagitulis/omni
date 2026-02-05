/**
 * Route Execution Config Types
 * Types for manual trigger mode configuration
 * API types use snake_case to match backend JSON response
 */

export type ExecutionMode = "queue" | "direct";
export type ExecutionPriority = "low" | "normal" | "high";

export interface RouteExecutionConfig {
  id: number;
  route_key: string;
  route_name: string;
  description?: string;
  execution_mode: ExecutionMode;
  priority: ExecutionPriority;
  enabled: boolean;
  icon?: string;
  category?: string;
  created_at: Date;
  updated_at: Date;
}

export interface RouteExecutionConfigInput {
  route_key: string;
  route_name: string;
  description?: string;
  execution_mode: ExecutionMode;
  priority?: ExecutionPriority;
  icon?: string;
  category?: string;
}
