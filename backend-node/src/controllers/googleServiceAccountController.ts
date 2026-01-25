import { Request, Response } from "express";
import { getServiceAccountRotationManager } from "../services/googleSheetsServiceAccountRotationManager";
import { getGoogleSheetsAutoDetector } from "../services/googleSheetsAutoDetector";
import { getGoogleSheetsAutoDiscovery } from "../services/googleSheetsAutoDiscoveryService";
import { getLogger } from "../utils/logger";

const logger = getLogger("GoogleServiceAccountController");

export class GoogleServiceAccountController {
  /**
   * GET /api/google/service-accounts
   * Get list of service accounts and current active one
   */
  async getServiceAccounts(_req: Request, res: Response): Promise<void> {
    try {
      const manager = getServiceAccountRotationManager();
      const serviceAccounts = manager.getAllAccounts();
      const activeAccount = manager.getActiveAccount();

      res.status(200).json({
        serviceAccounts: serviceAccounts.map((acc) => ({
          id: acc.clientEmail,
          email: acc.clientEmail,
          isActive: acc.isActive,
          quotaUsed: acc.quotaUsedPercent,
          usageCount: acc.usageCount,
          lastUsedAt: acc.lastUsedAt,
        })),
        activeAccount: activeAccount
          ? {
              id: activeAccount.clientEmail,
              email: activeAccount.clientEmail,
              isActive: activeAccount.isActive,
              quotaUsed: activeAccount.quotaUsedPercent,
            }
          : null,
      });
    } catch (error: any) {
      logger.error("Error getting service accounts", error);
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * GET /api/google/service-accounts/stats
   * Get detailed service account statistics
   */
  async getServiceAccountStats(_req: Request, res: Response): Promise<void> {
    try {
      const manager = getServiceAccountRotationManager();
      const discovery = getGoogleSheetsAutoDiscovery();

      const stats = manager.getAccountStats();
      const discoveredFiles = await discovery.getServiceAccountsInfo();
      const allAccounts = manager.getAllAccounts();
      const activeAccount = manager.getActiveAccount();

      res.status(200).json({
        success: true,
        data: {
          totalAccounts: allAccounts.length,
          activeAccount: activeAccount
            ? {
                email: activeAccount.clientEmail,
                usageCount: activeAccount.usageCount,
                quotaUsedPercent: activeAccount.quotaUsedPercent,
                lastUsedAt: activeAccount.lastUsedAt,
              }
            : null,
          accounts: allAccounts.map((acc) => ({
            email: acc.clientEmail,
            projectId: acc.projectId,
            isActive: acc.isActive,
            usageCount: acc.usageCount,
            quotaUsedPercent: acc.quotaUsedPercent,
            lastUsedAt: acc.lastUsedAt,
            keyPath: acc.keyPath,
          })),
          discoveredFiles: discoveredFiles,
          stats: stats,
        },
      });
    } catch (error: any) {
      logger.error("Error getting service account stats", error);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/google/service-accounts/switch
   * Switch to different service account
   */
  async switchServiceAccount(req: Request, res: Response): Promise<void> {
    try {
      const { email } = req.body;
      if (!email) {
        res.status(400).json({ error: "Service account email is required" });
        return;
      }

      const manager = getServiceAccountRotationManager();
      const success = await manager.rotateServiceAccount("manual");

      if (!success) {
        res
          .status(400)
          .json({ error: "Failed to switch or only one account available" });
        return;
      }

      const newActive = manager.getActiveAccount();
      res.status(200).json({
        message: `Switched to service account`,
        activeAccount: newActive
          ? {
              id: newActive.clientEmail,
              email: newActive.clientEmail,
              isActive: newActive.isActive,
            }
          : null,
      });
    } catch (error: any) {
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * POST /api/google/service-accounts/detect-sheet
   * Detect sheet metadata from URL (name + sheets list)
   */
  async detectSheet(req: Request, res: Response): Promise<void> {
    try {
      const { spreadsheetUrl } = req.body;
      if (!spreadsheetUrl) {
        res.status(400).json({ error: "Spreadsheet URL is required" });
        return;
      }

      const detector = getGoogleSheetsAutoDetector();
      const manager = getServiceAccountRotationManager();
      const activeAccount = manager.getActiveAccount();

      if (!activeAccount || !activeAccount.authClient) {
        res.status(400).json({ error: "No service account available" });
        return;
      }

      const metadata = await detector.detectMetadata(
        spreadsheetUrl,
        activeAccount.authClient
      );

      if (!metadata) {
        res.status(400).json({ error: "Failed to detect spreadsheet" });
        return;
      }

      res.status(200).json(metadata);
    } catch (error: any) {
      res.status(500).json({ error: error.message });
    }
  }
}
