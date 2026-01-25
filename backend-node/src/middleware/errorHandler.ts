import { Request, Response, NextFunction } from "express";
import { backendLogger } from "../utils/backendLogger";

interface ApiError extends Error {
  statusCode?: number;
  code?: string;
  isSecurityError?: boolean;
}

// Generic error messages to prevent information disclosure
const GENERIC_MESSAGES: Record<number, string> = {
  400: "Invalid request",
  401: "Authentication required",
  403: "Access denied",
  404: "Resource not found",
  409: "Conflict detected",
  429: "Too many requests",
  500: "Internal server error",
};

// Patterns that should never be exposed to clients
const SENSITIVE_PATTERNS = [
  /password/i,
  /token/i,
  /secret/i,
  /credential/i,
  /api[_-]?key/i,
  /database/i,
  /sql/i,
  /prisma/i,
  /connection/i,
];

/**
 * Check if error message contains sensitive information
 */
function containsSensitiveInfo(message: string): boolean {
  return SENSITIVE_PATTERNS.some((pattern) => pattern.test(message));
}

/**
 * Get safe error message for client response
 */
function getSafeMessage(err: ApiError, statusCode: number): string {
  // Security errors should show their message (already sanitized)
  if (err.isSecurityError && err.message.startsWith("SECURITY_ERROR:")) {
    return err.message.replace("SECURITY_ERROR: ", "");
  }

  // Never expose messages that might contain sensitive info
  if (containsSensitiveInfo(err.message)) {
    return GENERIC_MESSAGES[statusCode] || "An error occurred";
  }

  // For 500 errors, always use generic message
  if (statusCode >= 500) {
    return GENERIC_MESSAGES[500];
  }

  // For client errors, return the message if it's safe
  return err.message || GENERIC_MESSAGES[statusCode] || "An error occurred";
}

export const errorHandler = (
  err: ApiError,
  req: Request,
  res: Response,
  _next: NextFunction
) => {
  const statusCode = err.statusCode || 500;
  const code = err.code || "INTERNAL_ERROR";
  const isProduction = process.env.NODE_ENV === "production";

  // Log error securely (full details for debugging)
  backendLogger.error("Request Error", err.message, {
    code,
    statusCode,
    path: req.path,
    method: req.method,
    userId: (req as any).userId,
    tenantId: (req as any).tenantId,
    isSecurityError: err.isSecurityError,
    ...(!isProduction && { stack: err.stack }),
  });

  // Get safe message for client
  const safeMessage = getSafeMessage(err, statusCode);

  // Response to client (minimal info in production)
  res.status(statusCode).json({
    success: false,
    error: {
      message: safeMessage,
      code,
      statusCode,
      ...(!isProduction && { details: err.stack }),
    },
  });
};

export default errorHandler;
