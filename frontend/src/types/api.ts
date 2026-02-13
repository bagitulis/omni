// API Response types
/**
 * Generic API response wrapper.
 * @template T - The type of the data property.
 */
export interface ApiResponse<T = unknown> {
  success?: boolean;
  data?: T;
  message?: string;
  error?: string;
  status?: string;
}

/**
 * Health check API response.
 */
export interface HealthCheckResponse {
  status: string;
  message?: string;
  timestamp?: string;
}

/**
 * Status response for platform token information.
 * Mirrors Python backend /api/status response format
 */
export interface StatusResponse {
  connection_status: "connected" | "disconnected";
  data: {
    [platform: string]: string; // Formatted token status string
  };
  success: boolean;
  timestamp: string;
  [key: string]: unknown;
}

/**
 * Response for execution actions.
 */
export interface ExecutionResponse {
  success: boolean;
  data?: unknown;
  error?: string;
  message?: string;
}

/**
 * Response for shipping file list.
 */
export interface ShippingFileResponse {
  files: string[];
  message?: string;
}

/**
 * Response for shipping process actions.
 */
export interface ShippingProcessResponse {
  success: boolean;
  message: string;
  data?: unknown;
}

/**
 * Debug check response for backend functions.
 */
export interface DebugCheckResponse {
  /** Map of function names to their enabled/disabled status. */
  functions: Record<string, boolean>;
  /** Optional message. */
  message?: string;
}

/**
 * Standardized API error response.
 */
export interface ApiErrorResponse {
  /** Error message. */
  error: string;
  /** Optional message. */
  message?: string;
  /** Optional HTTP status code. */
  status?: number;
}
