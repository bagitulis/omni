import { Router, Request, Response } from "express";
import { getTokenManager, TokenManager } from "../services/tokenManager";
import { tenantContext } from "../utils/tenantContext";

const router = Router();

/**
 * Per-tenant initialization tracking
 */
const initPromises: Map<string, Promise<void>> = new Map();

/**
 * Ensure TokenManager is initialized for the current tenant
 */
function ensureInitialized(
  tokenManager: TokenManager,
  tenantId: string
): Promise<void> {
  let initPromise = initPromises.get(tenantId);
  if (!initPromise) {
    initPromise = tokenManager.initialize();
    initPromises.set(tenantId, initPromise);
  }
  return initPromise;
}

/**
 * Helper to get tenant-aware TokenManager
 */
async function getTenantTokenManager(
  res: Response
): Promise<{ tokenManager: TokenManager; tenantId: string } | null> {
  const tenantId = tenantContext.getTenantIdOrNull();
  if (!tenantId) {
    res.status(400).json({
      status: "error",
      error: "Missing x-tenant-id header",
    });
    return null;
  }
  return { tokenManager: await getTokenManager(tenantId), tenantId };
}

/**
 * Refresh token for specific platform
 * POST /api/:platform/operations/refresh-token
 */
router.post(
  "/:platform/operations/refresh-token",
  async (req: Request, res: Response) => {
    const { platform } = req.params;

    try {
      const ctx = await getTenantTokenManager(res);
      if (!ctx) return;

      await ensureInitialized(ctx.tokenManager, ctx.tenantId);

      const validPlatforms = ["shopee", "lazada", "tiktok"];

      if (!validPlatforms.includes(platform.toLowerCase())) {
        return res.status(400).json({
          status: "error",
          error: `Invalid platform: ${platform}`,
        });
      }

      const success = await ctx.tokenManager.refreshToken(
        platform.toLowerCase() as "shopee" | "lazada" | "tiktok"
      );

      if (success) {
        const tokenStatus = await ctx.tokenManager.getTokenStatus(
          platform.toLowerCase() as "shopee" | "lazada" | "tiktok"
        );

        return res.status(200).json({
          status: "success",
          platform: platform.toLowerCase(),
          message: "Token refreshed successfully",
          data: tokenStatus,
        });
      } else {
        return res.status(500).json({
          status: "error",
          platform: platform.toLowerCase(),
          error: "Failed to refresh token",
        });
      }
    } catch (error) {
      return res.status(500).json({
        status: "error",
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

/**
 * Get token status for specific platform
 * GET /api/:platform/operations/token-status
 */
router.get(
  "/:platform/operations/token-status",
  async (req: Request, res: Response) => {
    const { platform } = req.params;

    try {
      const ctx = await getTenantTokenManager(res);
      if (!ctx) return;

      await ensureInitialized(ctx.tokenManager, ctx.tenantId);

      const validPlatforms = ["shopee", "lazada", "tiktok"];

      if (!validPlatforms.includes(platform.toLowerCase())) {
        return res.status(400).json({
          status: "error",
          error: `Invalid platform: ${platform}`,
        });
      }

      const status = await ctx.tokenManager.getTokenStatus(
        platform.toLowerCase() as "shopee" | "lazada" | "tiktok"
      );

      return res.status(200).json({
        status: "success",
        platform: platform.toLowerCase(),
        data: status,
      });
    } catch (error) {
      return res.status(500).json({
        status: "error",
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

export default router;
