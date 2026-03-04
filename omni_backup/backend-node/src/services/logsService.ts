import * as fs from "fs";
import * as path from "path";

interface LogEntry {
  timestamp: string;
  level: string;
  logger: string;
  message: string;
}

const MAX_LOG_ENTRIES = 1000;
let debugLogsBuffer: LogEntry[] = [];

// Store original console methods
const originalLog = console.log;
const originalError = console.error;
const originalWarn = console.warn;
const originalInfo = console.info;

/**
 * Add log entry to buffer
 */
function addLogEntry(entry: LogEntry): void {
  debugLogsBuffer.push(entry);

  // Keep only last MAX_LOG_ENTRIES
  if (debugLogsBuffer.length > MAX_LOG_ENTRIES) {
    debugLogsBuffer.shift();
  }
}

/**
 * Intercept console methods to capture logs
 */
function setupConsoleInterception(): void {
  console.log = function (...args: any[]) {
    const message = args
      .map((arg) =>
        typeof arg === "object" ? JSON.stringify(arg) : String(arg)
      )
      .join(" ");
    addLogEntry({
      timestamp: new Date().toISOString(),
      level: "INFO",
      logger: "console",
      message,
    });
    originalLog.apply(console, args);
  };

  console.error = function (...args: any[]) {
    const message = args
      .map((arg) =>
        typeof arg === "object" ? JSON.stringify(arg) : String(arg)
      )
      .join(" ");
    addLogEntry({
      timestamp: new Date().toISOString(),
      level: "ERROR",
      logger: "console",
      message,
    });
    originalError.apply(console, args);
  };

  console.warn = function (...args: any[]) {
    const message = args
      .map((arg) =>
        typeof arg === "object" ? JSON.stringify(arg) : String(arg)
      )
      .join(" ");
    addLogEntry({
      timestamp: new Date().toISOString(),
      level: "WARNING",
      logger: "console",
      message,
    });
    originalWarn.apply(console, args);
  };

  console.info = function (...args: any[]) {
    const message = args
      .map((arg) =>
        typeof arg === "object" ? JSON.stringify(arg) : String(arg)
      )
      .join(" ");
    addLogEntry({
      timestamp: new Date().toISOString(),
      level: "INFO",
      logger: "console",
      message,
    });
    originalInfo.apply(console, args);
  };
}

/**
 * Initialize log capturing from Winston/file system and console
 */
export function initializeLogsService(): void {
  // Setup console interception first
  setupConsoleInterception();

  // Read from log files if they exist
  const logDir = path.join(process.cwd(), "logs");

  if (fs.existsSync(logDir)) {
    try {
      const files = fs.readdirSync(logDir);
      const logFiles = files.filter((f) => f.endsWith(".log"));

      for (const file of logFiles) {
        const filePath = path.join(logDir, file);
        const content = fs.readFileSync(filePath, "utf-8");
        const lines = content.split("\n");

        for (const line of lines) {
          if (line.trim()) {
            try {
              // Try to parse as JSON if it's in JSON format
              if (line.startsWith("{")) {
                const entry = JSON.parse(line);
                addLogEntry(entry);
              } else {
                // Parse simple log format
                const match = line.match(
                  /\[(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z)\]\s+(\w+)\s+(\[.*?\])\s+(.*)/
                );
                if (match) {
                  addLogEntry({
                    timestamp: match[1],
                    level: match[2],
                    logger: match[3],
                    message: match[4],
                  });
                }
              }
            } catch (e) {
              // Skip unparseable lines
            }
          }
        }
      }
    } catch (error) {
      originalError("Error initializing logs service:", error);
    }
  }

  originalLog("[✅] Logs service initialized - Console logging enabled");
}

/**
 * Get all debug logs with optional filtering
 */
export function getDebugLogs(
  level?: string,
  limit: number = 500
): {
  logs: LogEntry[];
  total_count: number;
  filtered_count: number;
} {
  let filtered = debugLogsBuffer;

  if (level) {
    filtered = debugLogsBuffer.filter(
      (log) => log.level === level.toUpperCase()
    );
  }

  // Get last `limit` entries
  const logs = filtered.slice(-limit);

  return {
    logs,
    total_count: debugLogsBuffer.length,
    filtered_count: filtered.length,
  };
}

/**
 * Clear all debug logs
 */
export function clearDebugLogs(): number {
  const oldCount = debugLogsBuffer.length;
  debugLogsBuffer = [];
  return oldCount;
}

/**
 * Log message to buffer (called from application)
 */
export function logMessage(
  level: string,
  logger: string,
  message: string
): void {
  addLogEntry({
    timestamp: new Date().toISOString(),
    level: level.toUpperCase(),
    logger,
    message,
  });

  // Also log to console for real-time debugging
  const timestamp = new Date().toISOString();
  originalLog(`[${timestamp}] ${level.toUpperCase()} [${logger}] ${message}`);
}
