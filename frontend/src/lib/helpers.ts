/**
 * Returns the standard number of decimal digits for a given currency.
 * Most SE Asian currencies use 0 (IDR, VND) or 2 (THB, PHP, MYR, SGD).
 */
function currencyDecimalDigits(currency: string): number {
  switch (currency.toUpperCase()) {
    case "IDR":
    case "VND":
    case "JPY":
    case "KRW":
      return 0;
    default:
      return 2;
  }
}

// Utility helper functions for the application
import { logger } from "@/lib/logger";

/**
 * Format currency with appropriate code, symbol, and precision.
 * Defaults to Indonesian Rupiah (IDR) with 0 decimal places.
 */
export function formatCurrency(amount: number, currency = "IDR", decimals?: number): string {
  const safeAmount = Number.isFinite(amount) ? amount : 0;
  const fractionDigits = decimals ?? currencyDecimalDigits(currency);
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(safeAmount);
}

/**
 * Format price as number with thousand separators.
 */
export function formatPrice(price: number | null | undefined): string {
  const safePrice = typeof price === "number" && Number.isFinite(price) ? price : 0;
  return new Intl.NumberFormat("id-ID").format(safePrice);
}

/**
 * Format number with thousand separators.
 */
export function formatNumber(num: number): string {
  const safeNumber = Number.isFinite(num) ? num : 0;
  return new Intl.NumberFormat("id-ID").format(safeNumber);
}

/**
 * Format date to readable string.
 */
export function formatDate(date: string | Date): string {
  try {
    const parsedDate = typeof date === "string" ? new Date(date) : date;
    return new Intl.DateTimeFormat("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      timeZone: "UTC",
    }).format(parsedDate);
  } catch (error) {
    logger.error("Error formatting date", { error });
    return String(date);
  }
}

/**
 * Format time to readable string.
 */
export function formatTime(timestamp: string): string {
  const date = new Date(timestamp);
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    timeZone: "UTC",
  }).format(date);
}

/**
 * Truncate text to specified length.
 */
export function truncateText(text: string, maxLength: number = 100): string {
  if (!text) {
    return "";
  }
  if (text.length <= maxLength) {
    return text;
  }
  return `${text.substring(0, maxLength)}...`;
}

/**
 * Check if value is empty.
 */
export function isEmpty(value: unknown): boolean {
  if (value === null || value === undefined) {
    return true;
  }
  if (typeof value === "string") {
    return value.trim() === "";
  }
  if (Array.isArray(value)) {
    return value.length === 0;
  }
  if (typeof value === "object") {
    return Object.keys(value).length === 0;
  }
  return false;
}

/**
 * Deep clone object.
 */
export function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== "object") {
    return obj;
  }

  if (obj instanceof Date) {
    return new Date(obj.getTime()) as T;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => deepClone(item)) as T;
  }

  const clonedObject: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
    clonedObject[key] = deepClone(value);
  }

  return clonedObject as T;
}

/**
 * Merge objects shallowly.
 */
export function mergeObjects<T extends Record<string, unknown>>(
  target: T,
  ...sources: Partial<T>[]
): T {
  return Object.assign({}, target, ...sources);
}

/**
 * Check if object has property.
 */
export function hasProperty(obj: unknown, prop: string): boolean {
  return obj !== null && typeof obj === "object" && prop in obj;
}

/**
 * Get nested property value safely.
 */
export function getNestedValue(
  obj: unknown,
  path: string,
  defaultValue?: unknown,
): unknown {
  const keys = path.split(".");
  let value: unknown = obj;

  for (const key of keys) {
    if (value !== null && typeof value === "object" && key in value) {
      value = (value as Record<string, unknown>)[key];
      continue;
    }

    return defaultValue;
  }

  return value;
}

/**
 * Sleep for specified milliseconds.
 */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

/**
 * Get resource status class based on percentage.
 */
export function getResourceClass(percent: number): string {
  if (percent > 80) {
    return "critical";
  }
  if (percent > 60) {
    return "warning";
  }
  return "healthy";
}

/**
 * Debounce function.
 */
export function debounce<T extends (...args: unknown[]) => void>(
  func: T,
  wait: number,
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;

  return (...args: Parameters<T>): void => {
    if (timeout !== null) {
      clearTimeout(timeout);
    }

    timeout = setTimeout(() => {
      func(...args);
      timeout = null;
    }, wait);
  };
}

/**
 * Throttle function.
 */
export function throttle<T extends (...args: unknown[]) => void>(
  func: T,
  limit: number,
): (...args: Parameters<T>) => void {
  let lastExecution = 0;

  return (...args: Parameters<T>): void => {
    const now = Date.now();
    if (now - lastExecution >= limit) {
      lastExecution = now;
      func(...args);
    }
  };
}
