import winston, { Logger } from "winston";

const loggers: Map<string, Logger> = new Map();

/**
 * Get or create a logger instance
 * Singleton pattern to avoid duplicate logger instances
 */
export function getLogger(label: string): Logger {
  if (loggers.has(label)) {
    return loggers.get(label)!;
  }

  const logger = winston.createLogger({
    level: process.env.LOG_LEVEL || "info",
    format: winston.format.combine(
      winston.format.timestamp({ format: "YYYY-MM-DD HH:mm:ss" }),
      winston.format.printf(({ timestamp, level, message, label }) => {
        const emoji = getEmojiForLevel(level);
        return `${emoji} [${timestamp}] [${label}] ${message}`;
      })
    ),
    defaultMeta: { label },
    transports: [
      new winston.transports.Console(),
      new winston.transports.File({
        filename: "logs/error.log",
        level: "error",
        maxsize: 5242880, // 5MB
        maxFiles: 5,
      }),
      new winston.transports.File({
        filename: "logs/combined.log",
        maxsize: 5242880, // 5MB
        maxFiles: 10,
      }),
    ],
  });

  loggers.set(label, logger);
  return logger;
}

/**
 * Get emoji for log level
 */
function getEmojiForLevel(level: string): string {
  switch (level.toLowerCase()) {
    case "error":
      return "❌";
    case "warn":
      return "⚠️";
    case "info":
      return "ℹ️";
    case "debug":
      return "🔧";
    default:
      return "📝";
  }
}

/**
 * Close all loggers gracefully
 */
export async function closeLoggers(): Promise<void> {
  for (const [_label, logger] of loggers) {
    if (logger && typeof (logger as any).close === "function") {
      await new Promise<void>((resolve) => {
        (logger as any).close(() => {
          resolve();
        });
      });
    }
  }
  loggers.clear();
}
