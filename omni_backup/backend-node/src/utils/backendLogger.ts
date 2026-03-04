/**
 * Backend Logger Utility
 * Provides consistent logging across backend services
 * Only logs important messages: errors, success, warnings, and startup info
 */

class BackendLogger {
  private isDev = process.env.NODE_ENV === "development";

  error(context: string, message: string, error?: any): void {
    const timestamp = new Date().toISOString();
    console.error(
      `❌ [${timestamp}] [${context}] ${message}`,
      error ? "\n" + error : ""
    );
  }

  success(context: string, message: string): void {
    if (this.isDev) {
      const timestamp = new Date().toISOString();
      console.log(`✅ [${timestamp}] [${context}] ${message}`);
    }
  }

  warning(context: string, message: string): void {
    const timestamp = new Date().toISOString();
    console.warn(`⚠️  [${timestamp}] [${context}] ${message}`);
  }

  // For initialization and startup info
  info(context: string, message: string): void {
    const timestamp = new Date().toISOString();
    console.log(`ℹ️  [${timestamp}] [${context}] ${message}`);
  }

  // For HTTP responses (only log failures by default)
  httpResponse(
    method: string,
    path: string,
    status: number,
    duration: number
  ): void {
    const timestamp = new Date().toISOString();
    const statusEmoji =
      status >= 200 && status < 300 ? "✅" : status >= 400 ? "❌" : "⚠️ ";
    console.log(
      `${statusEmoji} [${timestamp}] ${method} ${path} - ${status} (${duration}ms)`
    );
  }

  // Only for development with DEBUG_MODE=true
  debug(context: string, message: string, data?: any): void {
    if (this.isDev && process.env.DEBUG_MODE === "true") {
      const timestamp = new Date().toISOString();
      console.log(`🔍 [${timestamp}] [${context}] ${message}`, data || "");
    }
  }
}

export const backendLogger = new BackendLogger();
