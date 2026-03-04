import { Router, Request, Response } from "express";
import { getTokenManager, TokenManager } from "../services/tokenManager";
import { tenantContext } from "../utils/tenantContext";

const router = Router();

/**
 * Per-tenant initialization tracking
 * Uses Map to track init promises per tenant
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
 * Returns null if no tenant context
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
 * Token Status Endpoint - Get all platforms
 * GET /api/token-status
 * Returns status of all platform tokens
 * REQUIRES tenant context via x-tenant-id header
 */
router.get("/token-status", async (_req: Request, res: Response) => {
  try {
    const ctx = await getTenantTokenManager(res);
    if (!ctx) return;

    await ensureInitialized(ctx.tokenManager, ctx.tenantId);

    const allStatus = await ctx.tokenManager.getAllTokenStatus();

    res.status(200).json({
      status: "success",
      data: allStatus,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    res.status(500).json({
      status: "error",
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * Platform specific token status
 * GET /api/:platform/token-status
 */
router.get("/:platform/token-status", async (req: Request, res: Response) => {
  const { platform } = req.params;

  try {
    const ctx = await getTenantTokenManager(res);
    if (!ctx) return;

    await ensureInitialized(ctx.tokenManager, ctx.tenantId);

    const validPlatforms = ["shopee", "lazada", "tiktok"];

    if (!validPlatforms.includes(platform.toLowerCase())) {
      return res.status(400).json({
        status: "error",
        error: `Invalid platform: ${platform}. Must be one of: ${validPlatforms.join(
          ", "
        )}`,
      });
    }

    const status = await ctx.tokenManager.getTokenStatus(
      platform.toLowerCase() as "shopee" | "lazada" | "tiktok"
    );

    return res.status(200).json({
      status: "success",
      platform: platform.toLowerCase(),
      data: status,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    return res.status(500).json({
      status: "error",
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * Refresh All Tokens
 * POST /api/tokens/refresh-all
 * POST /api/tokens/refresh-all?force=true (force refresh even if not expired)
 * Refreshes all platform tokens
 */
router.post("/tokens/refresh-all", async (req: Request, res: Response) => {
  try {
    const ctx = await getTenantTokenManager(res);
    if (!ctx) return;

    await ensureInitialized(ctx.tokenManager, ctx.tenantId);

    const force = req.query.force === "true";
    const results = await ctx.tokenManager.refreshAllTokens(force);

    res.status(200).json({
      status: "success",
      data: results,
      message: force
        ? "Force token refresh completed"
        : "Token refresh completed",
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    res.status(500).json({
      status: "error",
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

export default router;
