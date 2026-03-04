/**
 * Standardized API Response Types
 * Single Responsibility: Define consistent response structure across all endpoints
 */

/**
 * Standard API response interface
 * All endpoints should return this format for consistency
 */
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: ApiError;
  meta?: ResponseMeta;
}

/**
 * Standard API error structure
 */
export interface ApiError {
  message: string;
  code: string;
  statusCode?: number;
  details?: Record<string, unknown>;
}

/**
 * Response metadata for pagination, etc.
 */
export interface ResponseMeta {
  total?: number;
  offset?: number;
  limit?: number;
  page?: number;
  totalPages?: number;
}

/**
 * Paginated response data
 */
export interface PaginatedData<T> {
  items: T[];
  total: number;
  offset: number;
  limit: number;
}

/**
 * Error codes for consistent error identification
 */
export const ErrorCodes = {
  // Client errors (4xx)
  BAD_REQUEST: "BAD_REQUEST",
  UNAUTHORIZED: "UNAUTHORIZED",
  FORBIDDEN: "FORBIDDEN",
  NOT_FOUND: "NOT_FOUND",
  VALIDATION_ERROR: "VALIDATION_ERROR",
  CONFLICT: "CONFLICT",
  RATE_LIMITED: "RATE_LIMITED",

  // Server errors (5xx)
  INTERNAL_ERROR: "INTERNAL_ERROR",
  SERVICE_UNAVAILABLE: "SERVICE_UNAVAILABLE",
  DATABASE_ERROR: "DATABASE_ERROR",
  EXTERNAL_API_ERROR: "EXTERNAL_API_ERROR",

  // Business logic errors
  INVALID_OPERATION: "INVALID_OPERATION",
  RESOURCE_EXISTS: "RESOURCE_EXISTS",
  RESOURCE_NOT_FOUND: "RESOURCE_NOT_FOUND",
  TENANT_MISMATCH: "TENANT_MISMATCH",
} as const;

export type ErrorCode = (typeof ErrorCodes)[keyof typeof ErrorCodes];
