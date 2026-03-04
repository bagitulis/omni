/**
 * Google Sheets Auth Controller
 * Handles service account authentication status
 * Single Responsibility: Authentication status only
 */

import { Request, Response } from "express";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getLogger } from "../utils/logger";

const logger = getLogger("GoogleSheetsAuth");

/**
 * GET /api/google/auth/status
 * Get Google Sheets service account status
 */
export async function getAuthStatusController(
  _req: Request,
  res: Response
): Promise<void> {
  try {
    logger.info("🔄 Checking Google Sheets service account status");

    const service = getGoogleSheetsService();
    const status = await service.getAuthStatus();

    logger.info(
      `✅ Auth status retrieved | Authenticated: ${status.authenticated}`
    );

    res.json({
      success: true,
      data: status,
    });
  } catch (error: any) {
    logger.error(`❌ Error getting auth status: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}
