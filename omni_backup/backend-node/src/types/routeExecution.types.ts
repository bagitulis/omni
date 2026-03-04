/**
 * Route Execution Config Types
 * Defines types for manual trigger mode configuration
 */

export type ExecutionMode = "queue" | "direct";
export type ExecutionPriority = "low" | "normal" | "high";

export interface IRouteExecutionConfig {
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

export interface IRouteExecutionConfigInput {
  routeKey: string;
  routeName: string;
  description?: string;
  executionMode: ExecutionMode;
  priority?: ExecutionPriority;
  enabled?: boolean;
  icon?: string;
  category?: string;
}

export interface IRouteExecutionConfigUpdate {
  routeName?: string;
  description?: string;
  executionMode?: ExecutionMode;
  priority?: ExecutionPriority;
  enabled?: boolean;
  icon?: string;
  category?: string;
}
