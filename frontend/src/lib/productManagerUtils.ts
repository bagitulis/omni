import { stripHtmlSafe } from "./sanitizer";

export function formatNumber(num: number | string | null | undefined): string {
  if (!num && num !== 0) return "-";
  return new Intl.NumberFormat("id-ID").format(Number(num));
}

export function formatDate(
  timestamp: number | string | null | undefined,
): string {
  if (!timestamp) return "-";
  const dateObj =
    typeof timestamp === "string"
      ? new Date(timestamp)
      : new Date(parseInt(String(timestamp)));
  return dateObj.toLocaleDateString("id-ID");
}

export function formatDateTime(dateString: string | null | undefined): string {
  if (!dateString) return "-";
  return new Date(dateString).toLocaleString("id-ID");
}

/** @deprecated Use stripHtmlSafe from sanitizer.ts directly */
export function stripHtml(html: string | null | undefined): string {
  return stripHtmlSafe(html);
}

export function createDefaultFilterState(
  fieldNames: string[],
): Record<string, string> {
  return fieldNames.reduce<Record<string, string>>((acc, field) => {
    acc[field] = "";
    return acc;
  }, {});
}

export function createDefaultVisibleColumns(
  fieldNames: string[],
): Record<string, boolean> {
  return fieldNames.reduce<Record<string, boolean>>((acc, field) => {
    acc[field] = true;
    return acc;
  }, {});
}

export function debounce<T extends (...args: Parameters<T>) => ReturnType<T>>(
  func: T,
  delay: number,
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => func(...args), delay);
  };
}
