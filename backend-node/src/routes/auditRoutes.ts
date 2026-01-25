import { Router } from "express";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { authMiddleware } from "../middleware/authMiddleware";
import { permissionMiddleware } from "../middleware/permissionMiddleware";
import { getAuditController } from "../controllers/auditController";
import { getLogger } from "../utils/logger";
import { sendForbidden } from "../utils/apiResponse";

getLogger("AuditRoutes"); // Initialize logger for module
const controller = getAuditController();

/**
 * Audit Routes
 * Single Responsibility: Audit log endpoints with permission checks
 *
 * Endpoints:
 *   GET /api/audit/logs/user - Get user's own audit logs
 *   GET /api/audit/logs/tenant - Get tenant audit logs (admin only)
 *   POST /api/audit/cleanup - Delete old audit logs (owner only)
 */
export const auditRouter = Router();

// Authentication middleware - verify JWT token first
auditRouter.use(authMiddleware);

// Extract role from token
auditRouter.use((req: any, _res, next) => {
  if (req.role) {
    (req as AuthRequest).userRole = req.role;
  }
  next();
});

/**
 * GET /api/audit/logs/user
 * Get audit logs for current user (authenticated users)
 *
 * Query parameters:
 *   - limit: max records to return (default 50)
 */
auditRouter.get(
  "/logs/user",
  permissionMiddleware("audit.view"),
  async (req: AuthRequest, res) => {
    await controller.getUserAuditLogs(req, res);
  }
);

/**
 * GET /api/audit/logs/tenant
 * Get audit logs for entire tenant (admin only)
 *
 * Query parameters:
 *   - limit: max records to return (default 100)
 */
auditRouter.get(
  "/logs/tenant",
  permissionMiddleware("audit.view"),
  async (req: AuthRequest, res) => {
    await controller.getTenantAuditLogs(req, res);
  }
);

/**
 * POST /api/audit/cleanup
 * Delete audit logs older than specified days (owner only)
 *
 * Body:
 * {
 *   "daysOld": 90
 * }
 */
auditRouter.post(
  "/cleanup",
  permissionMiddleware("audit.view"),
  async (req: AuthRequest, res) => {
    // Only owner can cleanup
    if ((req as AuthRequest).userRole !== "owner") {
      return sendForbidden(res, "Only owner can cleanup audit logs");
    }
    return await controller.cleanupAuditLogs(req, res);
  }
);

export default auditRouter;
