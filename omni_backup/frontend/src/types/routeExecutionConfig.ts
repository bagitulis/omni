/**
 * Route Execution Config Types
 * Types for manual trigger mode configuration
 */

export type ExecutionMode = "queue" | "direct";
export type ExecutionPriority = "low" | "normal" | "high";

export interface RouteExecutionConfig {
  id: number;
  routeKey: string;
  routeName: string;
  description?: string;
  executionMode: ExecutionMode;
  priority: ExecutionPriority;
  enabled: boolean;
  icon?: string;
  category?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface RouteExecutionConfigInput {
  routeKey: string;
  routeName: string;
  description?: string;
  executionMode: ExecutionMode;
  priority?: ExecutionPriority;
  icon?: string;
  category?: string;
}
