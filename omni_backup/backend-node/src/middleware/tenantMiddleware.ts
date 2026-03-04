import { Request, Response, NextFunction } from "express";
// import { getDbManager } from "../services/dbConnectionManager";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";

const logger = getLogger("TenantMiddleware");

/**
 * Paths that don't require tenant context
 * These are public endpoints, external webhooks, or health checks
 */
const TENANT_EXEMPT_PATHS = [
  "/api/health",
  "/api/csrf-token",
  "/api/captcha",
  "/api/auth/login",
  "/api/auth/register",
  "/api/webhooks", // External platform webhooks (Shopee, Lazada, Tiktok)
  "/api/platform-auth", // OAuth callbacks
  "/api/status",
  "/api/n8n", // n8n integration uses secret + tenantId query param
  "/api/order-manager", // n8n order manager webhook
];

/**
 * Check if path is exempt from tenant requirement
 */
function isTenantExemptPath(path: string): boolean {
  return TENANT_EXEMPT_PATHS.some(
    (exemptPath) => path === exemptPath || path.startsWith(exemptPath + "/")
  );
}

export interface AuthRequest extends Request {
  tenantId?: string;
  username?: string;
  userId?: string;
  userRole?: string;
}

export const tenantMiddleware = (
  req: AuthRequest,
  res: Response,
  next: NextFunction
): void => {
  // Skip tenant requirement for exempt paths
  if (isTenantExemptPath(req.path)) {
    // Still try to extract tenant if present (for logging/auditing)
    const headerTenantId = req.headers["x-tenant-id"] as string;
    if (headerTenantId) {
      req.tenantId = headerTenantId;
      tenantContext.run(headerTenantId, () => next());
    } else {
      next();
    }
    return;
  }

  // Extract tenant from x-tenant-id header
  const headerTenantId = req.headers["x-tenant-id"] as string;

  // Extract tenant from Authorization token payload
  const authHeader = req.headers.authorization;
  const username = req.headers["x-username"] as string;

  let tokenTenantId: string | undefined;
  let userRole = "user"; // default role

  // Try to extract from token if present
  if (authHeader && authHeader.startsWith("Bearer ")) {
    try {
      const token = authHeader.substring(7);
      const decoded = JSON.parse(
        Buffer.from(token.split(".")[1], "base64").toString()
      );
      tokenTenantId = decoded.tenantId;
      userRole = decoded.role || "user";
    } catch (error) {
      // Token decode failed - will be handled by authMiddleware
    }
  }

  // SECURITY: Validate that header tenant matches token tenant
  // If both are present, they MUST match (prevents tenant spoofing)
  if (headerTenantId && tokenTenantId && headerTenantId !== tokenTenantId) {
    logger.warn(
      `⚠️ Tenant ID mismatch! Header: ${headerTenantId}, Token: ${tokenTenantId}`
    );
    res.status(403).json({
      success: false,
      error: "Forbidden - Tenant ID mismatch between header and token",
    });
    return;
  }

  // Priority: Token tenant > Header tenant > Username
  // Token is the most trusted source (signed by server)
  const tenantId = tokenTenantId || headerTenantId || username;

  if (!tenantId) {
    logger.warn(
      "⚠️ No tenant ID found in request (token, header, or username)"
    );
    res.status(400).json({
      success: false,
      error:
        "Bad Request - Tenant ID required (via x-tenant-id header or authentication token)",
    });
    return;
  }

  req.tenantId = tenantId;
  req.userRole = userRole;

  // CRITICAL: Run next() within tenant context for AsyncLocalStorage isolation
  tenantContext.run(tenantId, () => {
    next();
  });
};

export const getTenantFromRequest = (req: AuthRequest): string => {
  return req.tenantId || tenantContext.getTenantId();
};
