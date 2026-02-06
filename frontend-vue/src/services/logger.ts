/**
 * Frontend Logger Utility
 * Filters out verbose debug logs (with emojis) for cleaner console
 * Only shows critical errors and important info
 */

class Logger {
  private isDev = import.meta.env.DEV;
  // Enable verbose mode only if explicitly set in localStorage
  private isVerbose =
    typeof window !== "undefined"
      ? localStorage.getItem("VERBOSE_LOGS") === "true"
      : false;

  error(message: string, error?: any): void {
    console.error(`❌ ${message}`, error || "");
  }

  success(message: string, data?: any): void {
    if (this.isVerbose && this.isDev) {
      console.log(`✅ ${message}`, data ? "|" : "", data || "");
    }
  }

  warning(message: string, data?: any): void {
    if (this.isVerbose) {
      console.warn(`⚠️  ${message}`, data ? "|" : "", data || "");
    }
  }

  // Only for development - won't show in production
  debug(message: string, data?: any): void {
    if (
      this.isVerbose &&
      this.isDev &&
      localStorage.getItem("DEBUG_MODE") === "true"
    ) {
      console.log(`🐛 ${message}`, data ? "|" : "", data || "");
    }
  }

  // Only for important initialization info
  info(message: string): void {
    console.log(`ℹ️  ${message}`);
  }

  // For API Backend connection only
  apiUrl(url: string): void {
    this.info(`API Backend: ${url}`);
  }
}

export const logger = new Logger();
