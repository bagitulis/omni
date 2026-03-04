/**
 * Google Sheets Error Handler Utility
 * Handles OAuth and API errors with recovery strategies
 *
 * Single Responsibility: Detect and handle Google Sheets specific errors
 */

import { Logger } from "winston";

export interface GoogleError {
  code?: string;
  message: string;
  status?: number;
}

export class GoogleSheetsErrorHandler {
  private logger: Logger;

  constructor(logger: Logger) {
    this.logger = logger;
  }

  /**
   * Check if error is an OAuth token error
   */
  isTokenError(error: any): boolean {
    const message = this.getErrorMessage(error);
    const code = this.getErrorCode(error);

    // Check for common token-related errors
    return (
      code === "invalid_grant" || // Token expired or revoked
      code === "invalid_token" || // Token is invalid
      message.includes("invalid_grant") ||
      message.includes("invalid_token") ||
      message.includes("Unauthorized") ||
      message.includes("401")
    );
  }

  /**
   * Check if error is an authentication error
   */
  isAuthError(error: any): boolean {
    const code = this.getErrorCode(error);
    const status = this.getErrorStatus(error);

    return (
      code === "unauthenticated" ||
      code === "UNAUTHENTICATED" ||
      status === 401 ||
      status === 403
    );
  }

  /**
   * Check if error is a rate limit error
   */
  isRateLimitError(error: any): boolean {
    const code = this.getErrorCode(error);
    const status = this.getErrorStatus(error);

    return (
      code === "429" ||
      code === "RESOURCE_EXHAUSTED" ||
      status === 429 ||
      this.getErrorMessage(error).includes("Rate Limit")
    );
  }

  /**
   * Check if error is a quota error
   */
  isQuotaError(error: any): boolean {
    const message = this.getErrorMessage(error);
    return (
      message.includes("quota") ||
      message.includes("Quota") ||
      message.includes("exceeded")
    );
  }

  /**
   * Check if error is a network/timeout error
   */
  isNetworkError(error: any): boolean {
    const message = this.getErrorMessage(error);
    return (
      error instanceof Error &&
      (message.includes("ETIMEDOUT") ||
        message.includes("ECONNREFUSED") ||
        message.includes("ENOTFOUND") ||
        message.includes("timeout"))
    );
  }

  /**
   * Extract error code
   */
  getErrorCode(error: any): string | undefined {
    return (
      error?.code ||
      error?.errors?.[0]?.reason ||
      error?.error?.error ||
      error?.error?.code
    );
  }

  /**
   * Extract error message
   */
  getErrorMessage(error: any): string {
    if (typeof error === "string") return error;
    if (error instanceof Error) return error.message;
    return (
      error?.message ||
      error?.errors?.[0]?.message ||
      error?.error?.message ||
      JSON.stringify(error)
    );
  }

  /**
   * Extract HTTP status code
   */
  getErrorStatus(error: any): number | undefined {
    return (
      error?.status ||
      error?.statusCode ||
      error?.response?.status ||
      error?.code
    );
  }

  /**
   * Log error with context
   */
  logError(context: string, error: any, severity: "error" | "warn" = "error") {
    const message = this.getErrorMessage(error);
    const code = this.getErrorCode(error);
    const status = this.getErrorStatus(error);

    const fullContext = `[${context}] ${message}${
      code ? ` (Code: ${code})` : ""
    }${status ? ` (Status: ${status})` : ""}`;

    if (severity === "error") {
      this.logger.error(fullContext);
    } else {
      this.logger.warn(fullContext);
    }

    return { message, code, status };
  }

  /**
   * Determine if error is recoverable
   */
  isRecoverable(error: any): boolean {
    // Token errors are recoverable via refresh
    if (this.isTokenError(error)) return true;

    // Rate limits are recoverable via retry
    if (this.isRateLimitError(error)) return true;

    // Network errors are recoverable via retry
    if (this.isNetworkError(error)) return true;

    return false;
  }

  /**
   * Get recovery recommendation
   */
  getRecoveryStrategy(error: any): string {
    if (this.isTokenError(error)) {
      return "token_refresh";
    }
    if (this.isRateLimitError(error)) {
      return "exponential_backoff";
    }
    if (this.isNetworkError(error)) {
      return "retry";
    }
    if (this.isQuotaError(error)) {
      return "quota_exhausted";
    }
    return "none";
  }
}
