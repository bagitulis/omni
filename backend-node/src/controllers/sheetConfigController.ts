import { Request, Response } from "express";
import { getPrisma } from "../services/prismaClient";
import { getLogger } from "../utils/logger";

const logger = getLogger("SheetConfigController");

/**
 * Save sheet configuration (spreadsheet URL, name, purpose, sheets)
 */
export const saveSheetConfig = async (
  req: Request,
  res: Response
): Promise<void> => {
  try {
    const { spreadsheetUrl, spreadsheetId, spreadsheetName, purpose, sheets } =
      req.body;
    const tenantId = (req as any).tenantId;
    const userId = (req as any).userId;

    if (!tenantId) {
      res.status(400).json({ error: "Tenant context required" });
      return;
    }

    const prisma = getPrisma(tenantId);

    if (!spreadsheetUrl || !spreadsheetId || !spreadsheetName) {
      res.status(400).json({ error: "Missing required fields" });
      return;
    }

    const config = await prisma.spreadsheet.upsert({
      where: {
        spreadsheetId_tenantId: { spreadsheetId, tenantId },
      },
      update: {
        spreadsheetUrl,
        spreadsheetName,
        purpose: purpose || "other",
        sheets: JSON.stringify(sheets || []),
        lastUsedAt: new Date(),
      },
      create: {
        spreadsheetId,
        spreadsheetUrl,
        spreadsheetName,
        purpose: purpose || "other",
        sheets: JSON.stringify(sheets || []),
        tenantId,
        registeredBy: userId,
      },
    });

    res.status(200).json({
      success: true,
      message: "Sheet configuration saved",
      data: config,
    });
  } catch (error) {
    logger.error("Failed to save sheet config:", error);
    res.status(500).json({ error: "Failed to save configuration" });
  }
};

/**
 * Load sheet configurations for current user/tenant
 */
export const loadSheetConfigs = async (
  req: Request,
  res: Response
): Promise<void> => {
  try {
    const tenantId = (req as any).tenantId;

    if (!tenantId) {
      res.status(400).json({ error: "Tenant context required" });
      return;
    }

    const prisma = getPrisma(tenantId);

    const sheets = await prisma.spreadsheet.findMany({
      where: { tenantId },
    });

    res.status(200).json({
      success: true,
      sheets: sheets.map((s) => ({
        id: s.id,
        spreadsheetId: s.spreadsheetId,
        spreadsheetName: s.spreadsheetName,
        spreadsheetUrl: s.spreadsheetUrl,
        purpose: s.purpose,
        registeredAt: s.createdAt,
        registeredBy: s.registeredBy,
      })),
    });
  } catch (error) {
    logger.error("Failed to load sheet configs:", error);
    res.status(500).json({ error: "Failed to load configurations" });
  }
};

/**
 * Delete sheet configuration
 */
export const deleteSheetConfig = async (
  req: Request,
  res: Response
): Promise<void> => {
  try {
    const { sheetId } = req.params;
    const tenantId = (req as any).tenantId;

    if (!tenantId) {
      res.status(400).json({ error: "Tenant context required" });
      return;
    }

    const prisma = getPrisma(tenantId);

    await prisma.spreadsheet.deleteMany({
      where: { id: sheetId, tenantId },
    });

    res.status(200).json({ success: true, message: "Sheet deleted" });
  } catch (error) {
    logger.error("Failed to delete sheet config:", error);
    res.status(500).json({ error: "Failed to delete sheet" });
  }
};
