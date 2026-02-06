import { getAuthHeaders, getApiBaseUrl } from "../../../../utils/apiHeaders";

const API_BASE_URL = getApiBaseUrl("/inventory");

/**
 * Options for fetchWithRetry
 */
interface FetchOptions {
  timeout?: number;
  retryTimeout?: number;
  retries?: number;
  method?: string;
  body?: string;
}

/**
 * Fetch with automatic timeout and retry logic
 * Reduces code duplication across inventory data fetching functions
 */
export async function fetchWithRetry(
  endpoint: string,
  options: FetchOptions = {},
): Promise<any> {
  const {
    timeout = 30000,
    retryTimeout = 60000,
    retries = 1,
    method = "GET",
    body,
  } = options;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  const cacheBuster = `_t=${Date.now()}`;
  const separator = endpoint.includes("?") ? "&" : "?";
  const url = `${API_BASE_URL}${endpoint}${separator}${cacheBuster}`;

  const fetchOptions = {
    method,
    signal: controller.signal,
    headers: {
      ...getAuthHeaders(),
      ...(body ? { "Content-Type": "application/json" } : {}),
    },
    ...(body ? { body } : {}),
  } as globalThis.RequestInit;

  try {
    const response = await fetch(url, fetchOptions);
    clearTimeout(timeoutId);

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    clearTimeout(timeoutId);

    // Handle timeout with retry
    if (error instanceof Error && error.name === "AbortError" && retries > 0) {
      console.warn(`⚠️ Request timeout (${timeout}ms), retrying...`);
      return fetchWithRetry(endpoint, {
        timeout: retryTimeout,
        retryTimeout,
        retries: retries - 1,
        method,
        body,
      });
    }

    throw error;
  }
}

/**
 * Check if API response is successful
 * Supports both legacy (status: "SUCCESS") and new (success: true) format
 */
export function isSuccessResponse(result: any): boolean {
  return result?.success === true || result?.status === "SUCCESS";
}

/**
 * Extract data from API response
 * Handles both formats: { data: {...} } and direct data
 */
export function extractResponseData<T>(result: any, defaultValue: T): T {
  if (isSuccessResponse(result)) {
    return result.data ?? result ?? defaultValue;
  }
  return defaultValue;
}

/**
 * Default stats object
 */
export const DEFAULT_STATS = {
  total_records: 0,
  total_columns: 0,
  columns: [],
  last_sync: null,
  db_size_kb: 0,
};
