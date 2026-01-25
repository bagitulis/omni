/**
 * Security Routes
 * API endpoints for security monitoring and notifications
 */

import { Router, Response } from "express";
import { authMiddleware, roleMiddleware } from "../middleware/authMiddleware";
import { getSecurityNotificationService } from "../services/securityNotificationService";
import { getLogger } from "../utils/logger";

const router = Router();
const logger = getLogger("SecurityRoutes");

// All security routes require authentication and admin role
router.use(authMiddleware);
router.use(roleMiddleware("owner", "admin", "developer"));

/**
 * GET /api/security/events
 * Get recent security events
 */
router.get("/events", async (req, res: Response) => {
  try {
    const limit = parseInt(req.query.limit as string) || 50;
    const securityService = getSecurityNotificationService();
    const events = securityService.getRecentEvents(limit);

    res.json({
      success: true,
      data: events,
      count: events.length,
    });
  } catch (error: any) {
    logger.error(`Failed to get security events: ${error.message}`);
    res.status(500).json({
      success: false,
      error: "Failed to retrieve security events",
    });
  }
});

/**
 * GET /api/security/stats
 * Get security event statistics
 */
router.get("/stats", async (_req, res: Response) => {
  try {
    const securityService = getSecurityNotificationService();
    const stats = securityService.getEventStats();
    const recentEvents = securityService.getRecentEvents(100);

    // Count by type
    const typeCounts = {
      critical: recentEvents.filter((e) => e.type === "critical").length,
      error: recentEvents.filter((e) => e.type === "error").length,
      warning: recentEvents.filter((e) => e.type === "warning").length,
    };

    res.json({
      success: true,
      data: {
        totalEvents: recentEvents.length,
        byType: typeCounts,
        byCode: stats,
      },
    });
  } catch (error: any) {
    logger.error(`Failed to get security stats: ${error.message}`);
    res.status(500).json({
      success: false,
      error: "Failed to retrieve security statistics",
    });
  }
});

/**
 * GET /api/security/alerts
 * Get critical and error level events (for notification badge)
 */
router.get("/alerts", async (_req, res: Response) => {
  try {
    const securityService = getSecurityNotificationService();
    const criticalEvents = securityService.getEventsByType("critical");
    const errorEvents = securityService.getEventsByType("error");

    res.json({
      success: true,
      data: {
        critical: criticalEvents.slice(0, 10),
        errors: errorEvents.slice(0, 20),
        counts: {
          critical: criticalEvents.length,
          error: errorEvents.length,
        },
      },
    });
  } catch (error: any) {
    logger.error(`Failed to get security alerts: ${error.message}`);
    res.status(500).json({
      success: false,
      error: "Failed to retrieve security alerts",
    });
  }
});

export default router;
