export interface TokenStatusData {
  isExpired: boolean;
  expiresAt: string | null;
  refreshTokenExpiresAt: string | null;
  status?: string;
  valid?: boolean;
}

export interface TokenStatusMap {
  shopee?: TokenStatusData;
  tiktok?: TokenStatusData;
  lazada?: TokenStatusData;
  [key: string]: TokenStatusData | undefined;
}

// Backend response format from /platform-auth/status
export interface BackendPlatformStatus {
  connected: boolean;
  expired: boolean;
  expires_soon: boolean;
  expires_at: number;
  refresh_token_expires_at?: number;
  shop_id?: string;
  shop_name?: string;
}

export interface BackendTokenStatusResponse {
  [platform: string]: BackendPlatformStatus;
}

export const PLATFORM_CONFIG: Record<string, { color: string; label: string }> =
  {
    shopee: { color: "#ee4d2d", label: "Shopee" },
    tiktok: { color: "#000000", label: "TikTok" },
    lazada: { color: "#0f146d", label: "Lazada" },
  };
