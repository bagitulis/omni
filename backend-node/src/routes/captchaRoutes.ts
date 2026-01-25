/**
 * CAPTCHA Routes
 * Endpoints for Google reCAPTCHA configuration
 *
 * These endpoints are public (no auth required)
 */

import { Router, Request, Response } from "express";
import captchaService from "../services/captchaService";
import { getLogger } from "../utils/logger";

const router = Router();
const logger = getLogger("CaptchaRoutes");

/**
 * Get reCAPTCHA site key for frontend
 * GET /api/captcha/site-key
 *
 * Returns the public site key for initializing reCAPTCHA on frontend
 */
router.get("/site-key", (_req: Request, res: Response) => {
  try {
    const siteKey = captchaService.getSiteKey();
    const configured = captchaService.isConfigured();

    res.status(200).json({
      success: true,
      siteKey: siteKey || null,
      configured,
    });
  } catch (error: any) {
    logger.error("Error getting reCAPTCHA site key:", error);
    res.status(500).json({
      success: false,
      error: "Failed to get reCAPTCHA configuration",
    });
  }
});

/**
 * Get CAPTCHA stats (admin only, for monitoring)
 * GET /api/captcha/stats
 */
router.get("/stats", (_req: Request, res: Response) => {
  const stats = captchaService.getStats();
  res.status(200).json({
    success: true,
    ...stats,
  });
});

export default router;
