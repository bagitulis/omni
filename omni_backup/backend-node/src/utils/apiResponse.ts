/**
 * API Response Helper Functions
 * Single Responsibility: Create standardized API responses
 */

import { Response } from "express";
import {
  ApiResponse,
  ApiError,
  ResponseMeta,
  ErrorCodes,
  ErrorCode,
} from "../types/apiResponse";

/**
 * Send a successful response
 */
export function sendSuccess<T>(
  res: Response,
  data: T,
  statusCode = 200,
  meta?: ResponseMeta
): Response {
  const response: ApiResponse<T> = {
    success: true,
    data,
  };

  if (meta) {
    response.meta = meta;
  }

  return res.status(statusCode).json(response);
}

/**
 * Send a paginated successful response
 */
export function sendPaginated<T>(
  res: Response,
  items: T[],
  total: number,
  offset: number,
  limit: number
): Response {
  return sendSuccess(
    res,
    items,
    200,
    {
      total,
      offset,
      limit,
      page: Math.floor(offset / limit) + 1,
      totalPages: Math.ceil(total / limit),
    }
  );
}

/**
 * Send an error response
 */
export function sendError(
  res: Response,
  message: string,
  code: ErrorCode = ErrorCodes.INTERNAL_ERROR,
  statusCode = 500,
  details?: Record<string, unknown>
): Response {
  const error: ApiError = {
    message,
    code,
    statusCode,
  };

  if (details) {
    error.details = details;
  }

  const response: ApiResponse = {
    success: false,
    error,
  };

  return res.status(statusCode).json(response);
}

/**
 * Send a 400 Bad Request error
 */
export function sendBadRequest(
  res: Response,
  message: string,
  details?: Record<string, unknown>
): Response {
  return sendError(res, message, ErrorCodes.BAD_REQUEST, 400, details);
}

/**
 * Send a 401 Unauthorized error
 */
export function sendUnauthorized(
  res: Response,
  message = "Authentication required"
): Response {
  return sendError(res, message, ErrorCodes.UNAUTHORIZED, 401);
}

/**
 * Send a 403 Forbidden error
 */
export function sendForbidden(
  res: Response,
  message = "Access denied"
): Response {
  return sendError(res, message, ErrorCodes.FORBIDDEN, 403);
}

/**
 * Send a 404 Not Found error
 */
export function sendNotFound(
  res: Response,
  resource = "Resource"
): Response {
  return sendError(
    res,
    `${resource} not found`,
    ErrorCodes.NOT_FOUND,
    404
  );
}

/**
 * Send a 409 Conflict error
 */
export function sendConflict(
  res: Response,
  message: string
): Response {
  return sendError(res, message, ErrorCodes.CONFLICT, 409);
}

/**
 * Send a 500 Internal Server Error
 */
export function sendInternalError(
  res: Response,
  error: Error | string
): Response {
  const message = error instanceof Error ? error.message : error;
  return sendError(res, message, ErrorCodes.INTERNAL_ERROR, 500);
}

/**
 * Handle caught errors and send appropriate response
 * Extracts message safely from unknown error types
 */
export function handleError(
  res: Response,
  error: unknown,
  defaultMessage = "An unexpected error occurred"
): Response {
  let message = defaultMessage;

  if (error instanceof Error) {
    message = error.message;
  } else if (typeof error === "string") {
    message = error;
  }

  return sendInternalError(res, message);
}

/**
 * Create a successful response object (without sending)
 * Useful for services that need to return response objects
 */
export function createSuccessResponse<T>(
  data: T,
  meta?: ResponseMeta
): ApiResponse<T> {
  const response: ApiResponse<T> = {
    success: true,
    data,
  };

  if (meta) {
    response.meta = meta;
  }

  return response;
}

/**
 * Create an error response object (without sending)
 * Useful for services that need to return response objects
 */
export function createErrorResponse(
  message: string,
  code: ErrorCode = ErrorCodes.INTERNAL_ERROR,
  details?: Record<string, unknown>
): ApiResponse {
  const error: ApiError = {
    message,
    code,
  };

  if (details) {
    error.details = details;
  }

  return {
    success: false,
    error,
  };
}
