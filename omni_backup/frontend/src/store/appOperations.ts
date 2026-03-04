/**
 * App Store Operations
 * Handles status loading and token refresh logic
 */

import type { Ref } from "vue";
import type { StatusResponse } from "@/types/api";
import apiService from "@/services/api";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

const TOKEN_REFRESH_COOLDOWN_MS = 10 * 60 * 1000; // 10 minutes
const REFRESH_THRESHOLD_MS = 12 * 60 * 60 * 1000; // 12 hours before expiry

interface TokenEntry {
  isExpired?: boolean;
  connected?: boolean;
  refreshTokenExpiresAt?: string;
}

/**
 * Check if tokens need refresh based on status snapshot
 */
export function tokensNeedRefresh(snapshot: any): boolean {
  const entries = Object.values(snapshot || {});

  return entries.some((entry) => {
    if (typeof entry === "string") {
      return /EXPIRED|MISSING/i.test(entry);
    }

    if (entry && typeof entry === "object") {
      const tokenEntry = entry as TokenEntry;
      if (tokenEntry.isExpired || tokenEntry.connected === false) return true;
      if (tokenEntry.refreshTokenExpiresAt) {
        const expiresInMs =
          new Date(tokenEntry.refreshTokenExpiresAt).getTime() - Date.now();
        return expiresInMs < REFRESH_THRESHOLD_MS;
      }
    }

    return false;
  });
}

/**
 * Check if token refresh is in cooldown period
 */
export function isInCooldown(lastRefreshAt: number | null): boolean {
  return (
    lastRefreshAt !== null &&
    Date.now() - lastRefreshAt < TOKEN_REFRESH_COOLDOWN_MS
  );
}

/**
 * Refresh tokens for all platforms
 */
export async function refreshAllTokens(
  handleTokenOperation: (platform: string, operation: string) => Promise<any>,
  addLog: (message: string) => void
): Promise<void> {
  for (const platform of ["shopee", "lazada", "tiktok"]) {
    try {
      await handleTokenOperation(platform, "refresh_token");
    } catch (err: any) {
      addLog(`❌ Failed to refresh ${platform} token: ${err.message}`);
    }
  }
}

/**
 * Sync master product
 */
export async function syncMasterProduct(
  addLog: (message: string) => void
): Promise<void> {
  addLog("🔄 Syncing master product...");

  try {
    const response = await fetch(getApiBaseUrl("/master-product/sync"), {
      method: "POST",
      headers: {
        ...getAuthHeaders(),
        "Content-Type": "application/json",
      },
    });

    const data = await response.json();

    if (data.success) {
      addLog(
        `✅ Master product synced: ${data.message || "Sync completed successfully"}`
      );
    } else {
      addLog(
        `⚠️ Master product sync warning: ${data.message || "Sync completed with issues"}`
      );
    }
  } catch (error: any) {
    addLog(`⚠️ Error syncing master product: ${error.message}`);
  }
}

/**
 * Load status with optional token refresh
 */
export async function loadStatusWithTokens(
  status: Ref<any>,
  lastTokenRefreshAt: Ref<number | null>,
  connectionStatus: Ref<string>,
  isReconnecting: Ref<boolean>,
  loading: Ref<boolean>,
  handleTokenOperation: (platform: string, operation: string) => Promise<any>,
  addLog: (message: string) => void,
  shouldLog: boolean = true,
  refreshTokens: boolean = false
): Promise<StatusResponse> {
  loading.value = true;

  try {
    const statusResponse = await apiService.getStatus();

    if (statusResponse.success) {
      status.value = statusResponse.data;

      if (refreshTokens) {
        await handleTokenRefresh(
          status,
          lastTokenRefreshAt,
          handleTokenOperation,
          addLog,
          shouldLog
        );
      }

      handleReconnection(
        statusResponse,
        connectionStatus,
        isReconnecting,
        addLog,
        shouldLog
      );
    } else if (shouldLog) {
      addLog("❌ Failed to load status");
    }

    return statusResponse;
  } finally {
    loading.value = false;
  }
}

/**
 * Handle token refresh logic
 */
async function handleTokenRefresh(
  status: Ref<any>,
  lastTokenRefreshAt: Ref<number | null>,
  handleTokenOperation: (platform: string, operation: string) => Promise<any>,
  addLog: (message: string) => void,
  shouldLog: boolean
): Promise<void> {
  const inCooldown = isInCooldown(lastTokenRefreshAt.value);
  const shouldRefresh = tokensNeedRefresh(status.value);

  if (inCooldown) {
    if (shouldLog) addLog("⏳ Token refresh skipped (cooldown)");
    return;
  }

  if (!shouldRefresh) {
    if (shouldLog) addLog("✅ Tokens still valid, skipping refresh");
    return;
  }

  if (shouldLog) addLog("🔄 Refreshing tokens for all platforms...");
  await refreshAllTokens(handleTokenOperation, addLog);
  lastTokenRefreshAt.value = Date.now();

  // Reload status after token refresh
  if (shouldLog) addLog("🔄 Reloading token status...");
  try {
    const reloadedResponse = await apiService.getStatus();
    if (reloadedResponse.success) {
      status.value = reloadedResponse.data;
      if (shouldLog) addLog("✅ Token status updated");
    }
  } catch (err: any) {
    addLog(`⚠️ Failed to reload status: ${err.message}`);
  }
}

/**
 * Handle reconnection status
 */
function handleReconnection(
  statusResponse: StatusResponse,
  connectionStatus: Ref<string>,
  isReconnecting: Ref<boolean>,
  addLog: (message: string) => void,
  shouldLog: boolean
): void {
  if (
    (statusResponse as any).reconnected ||
    connectionStatus.value !== "connected"
  ) {
    connectionStatus.value = "connected";
    isReconnecting.value = true;
    addLog("✅ Backend reconnected successfully");
    setTimeout(() => {
      isReconnecting.value = false;
    }, 3000);
  } else if (shouldLog && !isReconnecting.value) {
    addLog("✅ Status loaded successfully");
  }
}
