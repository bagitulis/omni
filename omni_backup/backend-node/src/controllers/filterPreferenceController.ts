import { Request, Response } from "express";
import { getFilterPreferenceService } from "../services/filterPreferenceService";

class FilterPreferenceController {
  async saveFilterPreference(req: Request, res: Response): Promise<void> {
    try {
      const {
        platform,
        page,
        tab,
        visibleColumns,
        columnFilters,
        searchQuery,
        lockedColumns,
        filters,
      } = req.body;

      // Support both 'page' (new) and 'tab' (legacy) parameters
      const tabName = page || tab || "product";

      // Support inventory, tiktok, lazada, shopee
      const validPlatforms = ["tiktok", "lazada", "shopee", "inventory"];
      if (!platform || !validPlatforms.includes(platform)) {
        res.status(400).json({
          success: false,
          error: `Invalid platform. Must be one of: ${validPlatforms.join(
            ", "
          )}`,
        });
        return;
      }

      // Support both 'columnFilters' (new) and 'filters' (legacy) parameters
      const filtersToUse = columnFilters || filters || {};

      const result = await getFilterPreferenceService().saveFilterPreference(
        platform,
        tabName,
        {
          visibleColumns,
          columnFilters: filtersToUse,
          searchQuery: searchQuery || "",
          lockedColumns: lockedColumns || [],
        }
      );

      res.json(result);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      res.status(500).json({
        success: false,
        error: errorMessage,
      });
    }
  }

  async getFilterPreference(req: Request, res: Response): Promise<void> {
    try {
      const { platform, page, tab, limit, offset } = req.query;

      // Support both 'page' (new) and 'tab' (legacy) parameters
      const tabName = (page as string) || (tab as string) || "product";

      // Support inventory, tiktok, lazada, shopee
      const validPlatforms = ["tiktok", "lazada", "shopee", "inventory"];
      if (!platform || !validPlatforms.includes(platform as string)) {
        res.status(400).json({
          success: false,
          error: `Invalid platform. Must be one of: ${validPlatforms.join(
            ", "
          )}`,
        });
        return;
      }

      // Parse pagination params
      const limitNum = limit ? Math.min(parseInt(limit as string), 100) : 10;
      const offsetNum = offset ? parseInt(offset as string) : 0;

      // OPTIMIZED: Disable HTTP cache for filter preferences
      // Since filter prefs change frequently, don't cache in browser
      // Always validate with server to ensure fresh data
      res.setHeader("Cache-Control", "no-cache, must-revalidate");
      res.setHeader("ETag", `"${platform}_${tabName}"`);
      res.setHeader("Pragma", "no-cache");

      const result = await getFilterPreferenceService().getFilterPreference(
        platform as "tiktok" | "lazada" | "shopee" | "inventory",
        tabName,
        limitNum,
        offsetNum
      );

      res.json(result);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      res.status(500).json({
        success: false,
        error: errorMessage,
      });
    }
  }

  async deleteFilterPreference(req: Request, res: Response): Promise<void> {
    try {
      const { platform, page, tab } = req.query;

      // Support both 'page' (new) and 'tab' (legacy) parameters
      const tabName = (page as string) || (tab as string) || "product";

      // Support inventory, tiktok, lazada, shopee
      const validPlatforms = ["tiktok", "lazada", "shopee", "inventory"];
      if (!platform || !validPlatforms.includes(platform as string)) {
        res.status(400).json({
          success: false,
          error: `Invalid platform. Must be one of: ${validPlatforms.join(
            ", "
          )}`,
        });
        return;
      }

      const result = await getFilterPreferenceService().deleteFilterPreference(
        platform as "tiktok" | "lazada" | "shopee" | "inventory",
        tabName
      );

      res.json(result);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      res.status(500).json({
        success: false,
        error: errorMessage,
      });
    }
  }
}

export default new FilterPreferenceController();
