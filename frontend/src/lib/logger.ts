/**
 * Centralized Logger Service
 *
 * Provides structured logging with timestamp and log levels.
 * - debug/info: DEV only (suppressed in prod)
 * - warn/error: Always output (even in prod — these are important)
 *
 * Note: Terser's drop_console in vite.config.ts strips all console.* in minified builds,
 * but logger checks env independently for defense-in-depth.
 */
class Logger {
  private isDev: boolean;

  constructor() {
    this.isDev = import.meta.env.DEV;
  }

  debug(message: string, context?: Record<string, unknown>): void {
    if (!this.isDev) return;
    console.debug(
      `[DEBUG] ${new Date().toISOString()} ${message}`,
      context ?? "",
    );
  }

  info(message: string, context?: Record<string, unknown>): void {
    if (!this.isDev) return;
    console.info(
      `[INFO] ${new Date().toISOString()} ${message}`,
      context ?? "",
    );
  }

  warn(message: string, context?: Record<string, unknown>): void {
    console.warn(
      `[WARN] ${new Date().toISOString()} ${message}`,
      context ?? "",
    );
  }

  error(message: string, context?: Record<string, unknown>): void {
    console.error(
      `[ERROR] ${new Date().toISOString()} ${message}`,
      context ?? "",
    );
  }
}

export const logger = new Logger();
