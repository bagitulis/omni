/**
 * Spreadsheet Registry Controller
 * Manages spreadsheet registration and configuration
 *
 * Single Responsibility: Handle spreadsheet registry operations
 */

import { Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { getSpreadsheetRegistry } from "../services/spreadsheetRegistryService";
import { getGoogleSheetsAutoDetector } from "../services/googleSheetsAutoDetector";
import { getGoogleSheetsAuthService } from "../services/googleSheetsAuthService";

const logger = getLogger("SpreadsheetRegistry");

/**
 * POST /api/google/registry/register
 * Register new spreadsheet by URL
 */
export async function registerSpreadsheetController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { url, purpose } = req.body;
    // "automated" = fallback for logging when user not authenticated
    // NOT related to system.db tenant - per AGENTS.MD architecture
    const userId = (req as any).user?.id || "automated";

    if (!url || !purpose) {
      res.status(400).json({
        success: false,
        error: "URL and purpose are required",
      });
      return;
    }

    // Get auth service and detector
    const authService = getGoogleSheetsAuthService();
    const detector = getGoogleSheetsAutoDetector();
    const registry = getSpreadsheetRegistry();

    // Ensure auth is initialized
    await authService.ensureInitialized();

    const authClient = authService.getActiveAuthClient();
    if (!authClient) {
      res.status(500).json({
        success: false,
        error: "Service account not initialized",
      });
      return;
    }

    // Validate access
    const isAccessible = await detector.validateAccess(url, authClient);
    if (!isAccessible) {
      res.status(400).json({
        success: false,
        error:
          "Cannot access spreadsheet. Check URL and service account permissions.",
      });
      return;
    }

    // Detect metadata
    const metadata = await detector.detectMetadata(url, authClient);
    if (!metadata) {
      res.status(400).json({
        success: false,
        error: "Failed to detect spreadsheet metadata",
      });
      return;
    }

    // Register
    const registered = registry.registerSpreadsheet({
      url,
      spreadsheetId: metadata.spreadsheetId,
      spreadsheetName: metadata.name,
      sheets: metadata.sheets,
      purpose: purpose as any,
      createdBy: userId,
      createdAt: new Date(),
      isLocked: true,
      lastUsedAt: null,
      lastModifiedBy: userId,
      isBeingEdited: false,
      editingLockedUntil: null,
      syncSettings: {
        autoSync: false,
      },
    });

    logger.info(`✅ Registered spreadsheet: ${registered.spreadsheetName}`);

    res.json({
      success: true,
      data: registered,
    });
  } catch (error: any) {
    logger.error(`Error registering spreadsheet: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * GET /api/google/registry/:id
 * Get spreadsheet registration by ID
 */
export async function getSpreadsheetController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { id } = req.params;
    const registry = getSpreadsheetRegistry();

    const spreadsheet = registry.getSpreadsheet(id);
    if (!spreadsheet) {
      res.status(404).json({
        success: false,
        error: "Spreadsheet not found",
      });
      return;
    }

    res.json({
      success: true,
      data: spreadsheet,
    });
  } catch (error: any) {
    logger.error(`Error getting spreadsheet: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * GET /api/google/registry
 * Get all spreadsheets for user
 */
export async function getAllSpreadsheetsController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    // "automated" = fallback for logging when user not authenticated
    const userId = (req as any).user?.id || "automated";
    const registry = getSpreadsheetRegistry();

    const spreadsheets = registry.getSpreadsheetsByUser(userId);

    res.json({
      success: true,
      data: spreadsheets,
    });
  } catch (error: any) {
    logger.error(`Error getting spreadsheets: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/google/registry/:id/unlock
 * Unlock spreadsheet for editing
 */
export async function unlockSpreadsheetController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { id } = req.params;
    // "automated" = fallback for logging when user not authenticated
    const userId = (req as any).user?.id || "automated";
    const registry = getSpreadsheetRegistry();

    const success = registry.unlockSpreadsheet(id, userId);
    if (!success) {
      res.status(400).json({
        success: false,
        error: "Failed to unlock spreadsheet",
      });
      return;
    }

    res.json({
      success: true,
      message: "Spreadsheet unlocked",
    });
  } catch (error: any) {
    logger.error(`Error unlocking spreadsheet: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/google/registry/:id/lock
 * Lock spreadsheet after editing
 */
export async function lockSpreadsheetController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { id } = req.params;
    // "automated" = fallback for logging when user not authenticated
    const userId = (req as any).user?.id || "automated";
    const registry = getSpreadsheetRegistry();

    const success = registry.lockSpreadsheet(id, userId);
    if (!success) {
      res.status(400).json({
        success: false,
        error: "Failed to lock spreadsheet",
      });
      return;
    }

    res.json({
      success: true,
      message: "Spreadsheet locked",
    });
  } catch (error: any) {
    logger.error(`Error locking spreadsheet: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * DELETE /api/google/registry/:id
 * Delete spreadsheet registration
 */
export async function deleteSpreadsheetController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { id } = req.params;
    const registry = getSpreadsheetRegistry();

    const success = registry.deleteSpreadsheet(id);
    if (!success) {
      res.status(404).json({
        success: false,
        error: "Spreadsheet not found",
      });
      return;
    }

    logger.info(`✅ Deleted spreadsheet: ${id}`);

    res.json({
      success: true,
      message: "Spreadsheet deleted",
    });
  } catch (error: any) {
    logger.error(`Error deleting spreadsheet: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}
