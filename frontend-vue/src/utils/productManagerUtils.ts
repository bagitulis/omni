/**
 * Product Manager Utilities
 * Common formatting and helper functions used across all product manager components
 */

import { stripHtmlSafe } from "./sanitizer";

export const formatNumber = (
  num: number | string | null | undefined
): string => {
  if (!num && num !== 0) return "-";
  return new Intl.NumberFormat("id-ID").format(Number(num));
};

export const formatDate = (
  timestamp: number | string | null | undefined
): string => {
  if (!timestamp) return "-";
  const dateObj =
    typeof timestamp === "string"
      ? new Date(timestamp)
      : new Date(parseInt(String(timestamp)));
  return dateObj.toLocaleDateString("id-ID");
};

export const formatDateTime = (
  dateString: string | null | undefined
): string => {
  if (!dateString) return "-";
  return new Date(dateString).toLocaleString("id-ID");
};

/**
 * Strip HTML tags safely (SECURITY: prevents XSS)
 * @deprecated Use stripHtmlSafe from sanitizer.ts directly
 */
export const stripHtml = (html: string | null | undefined): string => {
  return stripHtmlSafe(html);
};

export const createDefaultFilterState = (
  fieldNames: string[]
): Record<string, string> => {
  return fieldNames.reduce(
    (acc, field) => {
      acc[field] = "";
      return acc;
    },
    {} as Record<string, string>
  );
};

export const createDefaultVisibleColumns = (
  fieldNames: string[]
): Record<string, boolean> => {
  return fieldNames.reduce(
    (acc, field) => {
      acc[field] = true;
      return acc;
    },
    {} as Record<string, boolean>
  );
};

export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  delay: number
): ((...args: Parameters<T>) => void) => {
  let timeoutId: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => func(...args), delay);
  };
};
