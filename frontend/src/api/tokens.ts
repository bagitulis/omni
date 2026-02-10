import apiClient from "@/api/client";

// Backend uses camelCase for token endpoints (exception to snake_case rule)
export interface PlatformTokenStatus {
  platform: string;
  shopID?: string;
  isValid: boolean;
  needsRefresh: boolean;
  isExpired?: boolean;
  expiresAt?: string;
  refreshTokenExpiresAt?: string;
}

export interface AllTokenStatusResponse {
  [key: string]: PlatformTokenStatus;
}

export interface RefreshTokenResponse {
  platform: string;
  isValid: boolean;
  expiresAt?: string;
}

export interface RefreshAllResponse {
  [key: string]: {
    success: boolean;
    isValid?: boolean;
    expiresAt?: string;
    error?: string;
    skipped?: boolean;
    reason?: string;
  };
}

/**
 * Get status of all platform tokens
 */
export async function getAllTokenStatus(): Promise<AllTokenStatusResponse> {
  const response =
    await apiClient.get<AllTokenStatusResponse>("/tokens/status");
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to fetch token status");
  }
  return response.data;
}

/**
 * Get status of a specific platform token
 */
export async function getTokenStatus(
  platform: string,
): Promise<PlatformTokenStatus> {
  const response = await apiClient.get<PlatformTokenStatus>(
    `/tokens/status/${platform}`,
  );
  // Note: Backend might return success: false but with data if token is invalid/expired
  // We should return the data if available, even if success is false, or handle it in the component
  if (!response.data) {
    throw new Error(response.error || "Failed to fetch platform token status");
  }
  return response.data;
}

/**
 * Refresh token for a specific platform
 */
export async function refreshToken(
  platform: string,
): Promise<RefreshTokenResponse> {
  const response = await apiClient.post<RefreshTokenResponse>(
    `/tokens/refresh/${platform}`,
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to refresh token");
  }
  return response.data;
}

/**
 * Refresh tokens for all platforms
 */
export async function refreshAllTokens(
  force: boolean = false,
): Promise<RefreshAllResponse> {
  const response = await apiClient.post<RefreshAllResponse>(
    "/tokens/refresh-all",
    null,
    {
      params: { force },
    },
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to refresh all tokens");
  }
  return response.data;
}
