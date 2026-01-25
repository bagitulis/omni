import { Router, Request, Response } from "express";
import { getTokenManager } from "../services/tokenManager";
import { tenantContext } from "../utils/tenantContext";

const router = Router();

/**
 * Health check endpoint
 * Used by frontend to verify backend is running
 * Does NOT require tenant context - public endpoint
 */
router.get("/health", async (_req: Request, res: Response) => {
  try {
    res.status(200).json({
      status: "healthy",
      timestamp: new Date().toISOString(),
      uptime: process.uptime(),
      environment: process.env.NODE_ENV || "development",
      message: "✅ Backend migration server is running",
    });
  } catch (error) {
    res.status(500).json({
      status: "unhealthy",
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * Status endpoint - mirrors Python backend /api/status
 * Returns connection status and detailed token info for all platforms
 * DEPRECATED: Use /api/token-status instead
 * Kept for backward compatibility with Python backend
 *
 * If no tenant header is provided, returns basic status without token info
 */
router.get("/status", async (_req: Request, res: Response) => {
  try {
    // Check if tenant context is available
    const tenantId = tenantContext.getTenantIdOrNull();

    // If no tenant, return basic status (for public health checks)
    if (!tenantId) {
      const now = new Date();
      const timestamp =
        now.toISOString().replace("T", " ").slice(0, 23) +
        now.getMilliseconds().toString().padStart(3, "0");

      return res.status(200).json({
        connection_status: "connected",
        data: {
          message:
            "Backend is running. Provide x-tenant-id header for token status.",
        },
        success: true,
        timestamp,
      });
    }

    // Get per-tenant TokenManager
    const tokenManager = await getTokenManager(tenantId);

    const platforms = ["shopee", "lazada", "tiktok"] as const;
    const data: Record<string, string> = {};

    // Fetch token status for each platform
    for (const platform of platforms) {
      try {
        const status = await tokenManager.getTokenStatus(platform);
        data[platform] = formatTokenStatus(platform, status);
      } catch (err) {
        data[platform] = `${capitalize(
          platform
        )} Token Status:\n  ❌ Error fetching token data`;
      }
    }

    // Format timestamp like Python: "2025-12-20 17:06:03.585025"
    const now = new Date();
    const timestamp =
      now.toISOString().replace("T", " ").slice(0, 23) +
      now.getMilliseconds().toString().padStart(3, "0");

    return res.status(200).json({
      connection_status: "connected",
      data,
      success: true,
      timestamp,
    });
  } catch (error) {
    const now = new Date();
    const timestamp =
      now.toISOString().replace("T", " ").slice(0, 23) +
      now.getMilliseconds().toString().padStart(3, "0");

    return res.status(500).json({
      connection_status: "disconnected",
      data: {},
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
      timestamp,
    });
  }
});

/**
 * Format token status similar to Python backend
 * Returns formatted string with emoji indicators and expiry times
 * Example:
 * Lazada Token Status:
 *   Access Token: ✅ VALID
 *     Expires in: 28d 21h 30m
 *   Refresh Token: ✅ VALID
 *     Expires in: 28d 21h 30m
 */
function formatTokenStatus(
  platform: "shopee" | "lazada" | "tiktok",
  status: any
): string {
  if (!status) {
    return `${capitalize(
      platform
    )} Token Status:\n  ❌ No token data available`;
  }

  const accessTokenStatus = status.isExpired ? "❌ EXPIRED" : "✅ VALID";
  const refreshTokenStatus = status.refreshTokenExpiresAt
    ? new Date(status.refreshTokenExpiresAt).getTime() < Date.now()
      ? "❌ EXPIRED"
      : "✅ VALID"
    : "⚠️ UNKNOWN";

  // Calculate time remaining for access token
  const expiresAt = new Date(status.expiresAt);
  const now = new Date();
  const diffMs = expiresAt.getTime() - now.getTime();

  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

  const accessTimeRemaining = `${days}d ${hours}h ${minutes}m`;

  // Calculate time remaining for refresh token
  let refreshTimeRemaining = "Unknown";
  if (status.refreshTokenExpiresAt) {
    const refreshExpiresAt = new Date(status.refreshTokenExpiresAt);
    const refreshDiffMs = refreshExpiresAt.getTime() - now.getTime();

    const refreshDays = Math.floor(refreshDiffMs / (1000 * 60 * 60 * 24));
    const refreshHours = Math.floor(
      (refreshDiffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)
    );
    const refreshMinutes = Math.floor(
      (refreshDiffMs % (1000 * 60 * 60)) / (1000 * 60)
    );

    refreshTimeRemaining = `${refreshDays}d ${refreshHours}h ${refreshMinutes}m`;
  }

  return (
    `${capitalize(platform)} Token Status:\n` +
    `  Access Token: ${accessTokenStatus}\n` +
    `    Expires in: ${accessTimeRemaining}\n` +
    `  Refresh Token: ${refreshTokenStatus}\n` +
    `    Expires in: ${refreshTimeRemaining}`
  );
}

/**
 * Capitalize first letter
 */
function capitalize(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export default router;
