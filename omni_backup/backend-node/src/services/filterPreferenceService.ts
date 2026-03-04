import { getLogger } from "../utils/logger";
import { getPrisma } from "./prismaClient";
import { tenantContext } from "../utils/tenantContext";

const log = getLogger("FilterPreferenceService");

interface FilterPreferences {
  visibleColumns: Record<string, boolean>;
  columnFilters: Record<string, any>;
  searchQuery: string;
  lockedColumns?: string[];
}

/**
 * Filter Preference Service with Multi-Tenant Support
 *
 * SINGLE RESPONSIBILITY: Manage filter preferences for each tenant
 * All operations are automatically scoped by tenantId
 */
class FilterPreferenceService {
  /**
   * Get prisma client for current tenant context
   * Called fresh each time to ensure correct tenant
   */
  private getPrismaClient() {
    return getPrisma();
  }

  async saveFilterPreference(
    platform: "tiktok" | "lazada" | "shopee" | "inventory",
    tab: string = "master",
    preferences: FilterPreferences
  ) {
    try {
      const tenantId = tenantContext.getTenantId();
      const prisma = this.getPrismaClient();

      // Use upsert to handle create/update atomically (prevents race condition)
      const result = await prisma.filterPreference.upsert({
        where: {
          tenantId_platform_tab: {
            tenantId,
            platform,
            tab,
          },
        },
        update: {
          visibleColumns: JSON.stringify(preferences.visibleColumns),
          columnFilters: JSON.stringify(preferences.columnFilters),
          searchQuery: preferences.searchQuery,
          lockedColumns: JSON.stringify(preferences.lockedColumns || []),
        },
        create: {
          tenantId,
          platform,
          tab,
          visibleColumns: JSON.stringify(preferences.visibleColumns),
          columnFilters: JSON.stringify(preferences.columnFilters),
          searchQuery: preferences.searchQuery,
          lockedColumns: JSON.stringify(preferences.lockedColumns || []),
        },
      });

      log.info(`Filter preferences saved for ${tenantId}/${platform} - ${tab}`);
      return { success: true, data: result };
    } catch (error) {
      log.error(
        `❌ Error saving filter preference: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
      return {
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  async getFilterPreference(
    platform: "tiktok" | "lazada" | "shopee" | "inventory",
    tab: string = "master",
    limit: number = 10,
    offset: number = 0
  ) {
    try {
      const tenantId = tenantContext.getTenantId();
      const prisma = this.getPrismaClient();
      const result = await prisma.filterPreference.findFirst({
        where: { tenantId, platform, tab },
      });

      if (!result) {
        const response = {
          success: true,
          data: {
            platform,
            tab,
            visibleColumns: [],
            columnFilters: {},
            searchQuery: "",
            lockedColumns: [],
            items: [],
          },
          pagination: { total: 0, limit, offset },
        };
        return response;
      }

      // Parse stored filters (these are typically column filters, not paginated data)
      const parsedFilters = JSON.parse(result.columnFilters || "{}");
      const parsedVisibleColumns = JSON.parse(result.visibleColumns || "[]");
      const parsedLockedColumns = JSON.parse(result.lockedColumns || "[]");

      // Ensure visibleColumns is always an array
      const visibleColumnsArray = Array.isArray(parsedVisibleColumns)
        ? parsedVisibleColumns
        : Object.keys(parsedVisibleColumns).filter(
            (key) => parsedVisibleColumns[key] === true
          );

      // If filters contain items array (for paginated results), apply pagination
      let items = [];
      let total = 0;

      if (Array.isArray(parsedFilters.items)) {
        items = parsedFilters.items;
        total = items.length;
        items = items.slice(offset, offset + limit);
      }

      // CRITICAL: Ensure all fields are always returned with correct types
      const response = {
        success: true,
        data: {
          platform: result.platform,
          tab: result.tab,
          visibleColumns: visibleColumnsArray,
          columnFilters: parsedFilters,
          searchQuery: result.searchQuery || "",
          lockedColumns: Array.isArray(parsedLockedColumns)
            ? parsedLockedColumns
            : [],
          items, // Paginated items if available
        },
        pagination: {
          total,
          limit,
          offset,
          hasMore: offset + limit < total,
        },
      };

      return response;
    } catch (error) {
      log.error(
        `❌ Error getting filter preference: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
      return {
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  async deleteFilterPreference(
    platform: "tiktok" | "lazada" | "shopee" | "inventory",
    tab: string = "master"
  ) {
    try {
      const tenantId = tenantContext.getTenantId();
      const prisma = this.getPrismaClient();
      await prisma.filterPreference.deleteMany({
        where: { tenantId, platform, tab },
      });

      log.info(
        `✅ Filter preference deleted for ${tenantId}/${platform} - ${tab}`
      );
      return { success: true };
    } catch (error) {
      log.error(
        `❌ Error deleting filter preference: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
      return {
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }
}

let instance: FilterPreferenceService;

export function getFilterPreferenceService(): FilterPreferenceService {
  if (!instance) {
    instance = new FilterPreferenceService();
  }
  return instance;
}

// Note: Do NOT export singleton instance directly
// Always use getFilterPreferenceService() to avoid module-load errors
export default getFilterPreferenceService;
