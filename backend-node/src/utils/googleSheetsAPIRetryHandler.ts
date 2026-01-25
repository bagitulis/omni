/**
 * Google Sheets API Retry Handler
 * Handles retry logic with token refresh for API calls
 *
 * Single Responsibility: Retry mechanism only
 */

import { Logger } from "winston";
import { GoogleSheetsErrorHandler } from "../utils/googleSheetsErrorHandler";

export class GoogleSheetsAPIRetryHandler {
  private errorHandler: GoogleSheetsErrorHandler;
  private logger: Logger;
  private authService: any;
  private maxRetries = 1;

  constructor(logger: Logger) {
    this.logger = logger;
    this.errorHandler = new GoogleSheetsErrorHandler(logger);
  }

  /**
   * Set auth service for token refresh
   */
  setAuthService(authService: any): void {
    this.authService = authService;
  }

  /**
   * Execute operation with retry on token error
   */
  async execute<T>(
    operation: () => Promise<T>,
    operationName: string,
    retryCount = 0
  ): Promise<T | null> {
    try {
      return await operation();
    } catch (error) {
      const errorInfo = this.errorHandler.logError(
        operationName,
        error,
        "warn"
      );

      // If token error and can refresh, try again
      if (
        this.errorHandler.isTokenError(error) &&
        this.authService &&
        retryCount < this.maxRetries
      ) {
        this.logger.info(
          `🔄 Detected token error (${errorInfo.code}), attempting refresh...`
        );

        const refreshed = await this.authService.refreshToken();

        if (refreshed) {
          this.logger.info(`✅ Token refreshed, retrying: ${operationName}`);
          return this.execute(operation, operationName, retryCount + 1);
        } else {
          this.logger.error(
            `❌ Token refresh failed, user must re-authenticate`
          );
          throw error;
        }
      }

      throw error;
    }
  }
}
