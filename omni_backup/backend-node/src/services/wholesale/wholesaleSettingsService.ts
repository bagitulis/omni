/**
 * Wholesale Settings Service
 * RESPONSIBILITY: Manage wholesale settings per tenant
 *
 * Settings stored in WholesaleSettings table:
 * - adminFee: Admin Shopee fee (default 1500)
 * - minOrder1, maxOrder1: Tier 1 qty range (editable)
 * - maxOrderTier3: Max qty for tier 3 (default 1000)
 *
 * Tier calculation:
 * - Tier 1: minOrder1 to maxOrder1 (editable)
 * - Tier 2: maxOrder1+1 to maxOrder1+2 (auto)
 * - Tier 3: maxOrder1+3 to maxOrderTier3 (auto)
 *
 * Price formula: (Harga - Admin) + (Admin / MinQty)
 */

import { PrismaClient } from "@prisma/client";
import { getLogger } from "../../utils/logger";
import { Logger } from "winston";

export interface WholesaleSettingsData {
  adminFee: number;
  maxOrderTier3: number;
  minOrder1: number;
  maxOrder1: number;
}

export interface WholesaleTierCalculated {
  tier: number;
  minCount: number;
  maxCount: number;
  unitPrice: number;
}

export interface WholesalePreview {
  basePrice: number;
  adminFee: number;
  tiers: WholesaleTierCalculated[];
}

const DEFAULT_SETTINGS: WholesaleSettingsData = {
  adminFee: 1500,
  maxOrderTier3: 1000,
  minOrder1: 2,
  maxOrder1: 3,
};

export class WholesaleSettingsService {
  private logger: Logger;

  constructor(private prisma?: PrismaClient) {
    this.logger = getLogger("WholesaleSettingsService");
  }

  /**
   * Get settings for tenant (create default if not exists)
   */
  async getSettings(tenantId: string): Promise<WholesaleSettingsData> {
    // Return defaults if no prisma
    if (!this.prisma) {
      this.logger.warn("⚠️ No Prisma client, returning default settings");
      return DEFAULT_SETTINGS;
    }

    try {
      let settings = await this.prisma.wholesaleSettings.findUnique({
        where: { tenantId },
      });

      if (!settings) {
        settings = await this.prisma.wholesaleSettings.create({
          data: {
            tenantId,
            platform: "shopee",
            ...DEFAULT_SETTINGS,
          },
        });
        this.logger.info(`✅ Created default settings for tenant: ${tenantId}`);
      }

      return {
        adminFee: settings.adminFee,
        maxOrderTier3: settings.maxOrderTier3,
        minOrder1: settings.minOrder1,
        maxOrder1: settings.maxOrder1,
      };
    } catch (error: any) {
      this.logger.error(`❌ Get settings error: ${error.message}`);
      return DEFAULT_SETTINGS;
    }
  }

  /**
   * Update settings for tenant
   */
  async updateSettings(
    tenantId: string,
    data: Partial<WholesaleSettingsData>
  ): Promise<WholesaleSettingsData> {
    // Return defaults if no prisma
    if (!this.prisma) {
      this.logger.warn("⚠️ No Prisma client, returning merged settings");
      return { ...DEFAULT_SETTINGS, ...data };
    }

    try {
      // Validate
      if (data.minOrder1 !== undefined && data.minOrder1 < 2) {
        throw new Error("minOrder1 must be at least 2");
      }
      if (
        data.maxOrder1 !== undefined &&
        data.maxOrder1 <= (data.minOrder1 || 2)
      ) {
        throw new Error("maxOrder1 must be greater than minOrder1");
      }

      const settings = await this.prisma.wholesaleSettings.upsert({
        where: { tenantId },
        update: data,
        create: {
          tenantId,
          platform: "shopee",
          ...DEFAULT_SETTINGS,
          ...data,
        },
      });

      this.logger.info(`✅ Updated settings for tenant: ${tenantId}`);

      return {
        adminFee: settings.adminFee,
        maxOrderTier3: settings.maxOrderTier3,
        minOrder1: settings.minOrder1,
        maxOrder1: settings.maxOrder1,
      };
    } catch (error: any) {
      this.logger.error(`❌ Update settings error: ${error.message}`);
      throw error;
    }
  }

  /**
   * Calculate wholesale tiers for a given base price
   *
   * Formula: Price = (Harga - Admin) + (Admin / MinQty)
   */
  calculateTiers(
    basePrice: number,
    settings: WholesaleSettingsData
  ): WholesaleTierCalculated[] {
    const { adminFee, minOrder1, maxOrder1, maxOrderTier3 } = settings;

    // Calculate tier boundaries
    const min1 = minOrder1;
    const max1 = maxOrder1;
    const min2 = max1 + 1;
    const max2 = min2 + 1;
    const min3 = max2 + 1;
    const max3 = maxOrderTier3;

    // Calculate prices using formula: (Harga - Admin) + (Admin / MinQty)
    const price1 = Math.round(basePrice - adminFee + adminFee / min1);
    const price2 = Math.round(basePrice - adminFee + adminFee / min2);
    const price3 = Math.round(basePrice - adminFee + adminFee / min3);

    return [
      { tier: 1, minCount: min1, maxCount: max1, unitPrice: price1 },
      { tier: 2, minCount: min2, maxCount: max2, unitPrice: price2 },
      { tier: 3, minCount: min3, maxCount: max3, unitPrice: price3 },
    ];
  }

  /**
   * Preview wholesale calculation for a price
   */
  preview(
    basePrice: number,
    settings: WholesaleSettingsData
  ): WholesalePreview {
    return {
      basePrice,
      adminFee: settings.adminFee,
      tiers: this.calculateTiers(basePrice, settings),
    };
  }

  /**
   * Convert calculated tiers to Shopee API format
   */
  toShopeeFormat(tiers: WholesaleTierCalculated[]): Array<{
    min_count: number;
    max_count: number;
    unit_price: number;
  }> {
    return tiers.map((tier) => ({
      min_count: tier.minCount,
      max_count: tier.maxCount,
      unit_price: tier.unitPrice,
    }));
  }
}
