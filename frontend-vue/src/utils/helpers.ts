// Utility helper functions for the application

/**
 * Format currency to Indonesian Rupiah
 */
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount || 0);
}

/**
 * Format price as number with thousand separators
 */
export function formatPrice(price: number | null | undefined): string {
  return new Intl.NumberFormat("id-ID").format(price || 0);
}

/**
 * Format number with thousand separators
 */
export function formatNumber(num: number): string {
  return new Intl.NumberFormat("id-ID").format(num || 0);
}

/**
 * Format date to readable string
 */
export function formatDate(date: string | Date): string {
  try {
    const d = typeof date === "string" ? new Date(date) : date;
    return new Intl.DateTimeFormat("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(d);
  } catch (error) {
    console.error("Error formatting date:", error);
    return String(date);
  }
}

/**
 * Format time to readable string
 */
export function formatTime(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString();
}

/**
 * Truncate text to specified length
 */
export function truncateText(text: string, maxLength: number = 100): string {
  if (!text) return "";
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + "...";
}

/**
 * Check if value is empty
 */
export function isEmpty(value: any): boolean {
  if (value === null || value === undefined) return true;
  if (typeof value === "string") return value.trim() === "";
  if (Array.isArray(value)) return value.length === 0;
  if (typeof value === "object") return Object.keys(value).length === 0;
  return false;
}

/**
 * Deep clone object
 */
export function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== "object") return obj;
  if (obj instanceof Date) return new Date(obj.getTime()) as any;
  if (obj instanceof Array) {
    const clonedArr: any[] = [];
    obj.forEach((value, index) => {
      clonedArr[index] = deepClone(value);
    });
    return clonedArr as any;
  }
  if (obj instanceof Object) {
    const clonedObj: any = {};
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        clonedObj[key] = deepClone(obj[key]);
      }
    }
    return clonedObj;
  }
  return obj;
}

/**
 * Merge objects shallowly
 */
export function mergeObjects<T extends Record<string, any>>(
  target: T,
  ...sources: Partial<T>[]
): T {
  return Object.assign({}, target, ...sources);
}

/**
 * Check if object has property
 */
export function hasProperty(obj: any, prop: string): boolean {
  return obj && typeof obj === "object" && prop in obj;
}

/**
 * Get nested property value safely
 */
export function getNestedValue(
  obj: any,
  path: string,
  defaultValue?: any
): any {
  const keys = path.split(".");
  let value = obj;

  for (const key of keys) {
    if (value && typeof value === "object" && key in value) {
      value = value[key];
    } else {
      return defaultValue;
    }
  }

  return value;
}

/**
 * Sleep for specified milliseconds
 */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Get resource status class based on percentage
 */
export function getResourceClass(percent: number): string {
  if (percent > 80) return "critical";
  if (percent > 60) return "warning";
  return "healthy";
}

/**
 * Debounce function
 */
export function debounce<T extends (..._args: any[]) => any>(
  func: T,
  wait: number
): (..._args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;

  return function executedFunction(..._args: Parameters<T>): void {
    const later = (): void => {
      timeout = null;

      (func as any)(..._args);
    };

    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

/**
 * Throttle function
 */
export function throttle<T extends (..._args: any[]) => any>(
  func: T,
  limit: number
): (..._args: Parameters<T>) => void {
  let inThrottle: boolean = false;
  const argsRef: any[] = [];

  return function executedFunction(..._args: Parameters<T>): void {
    argsRef.length = 0;
    argsRef.push(..._args);
    if (!inThrottle) {
      func.apply(null, argsRef);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
}
