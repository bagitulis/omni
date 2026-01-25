/**
 * TikTok Price Reconciliation Service
 * Analyzes price differences between TikTok settlement and inventory
 * Single Responsibility: Price analysis for TikTok orders
 */

import { getLogger } from "../../utils/logger";
import {
  SkuGroup,
  ReconciliationSummary,
  ReconciliationResult,
  AnalyticsSettingsData,
  SkuData,
  SkuInfo,
} from "./tiktokPriceReconciliationTypes";

const logger = getLogger("TiktokPriceReconciliationService");

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PrismaClientAny = any;

export class TiktokPriceReconciliationService {
  private prisma: PrismaClientAny;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  /**
   * Get analytics settings for TikTok
   */
  async getSettings(): Promise<AnalyticsSettingsData> {
    const settings = await this.prisma.analyticsSettings.findFirst({
      where: {
        tenantId: this.tenantId,
        platform: "tiktok",
      },
    });

    return {
      priceColumn: settings?.priceColumn || "HARGA",
      formulaDeduction: settings?.formulaDeduction || 1500,
      formulaMultiplier: settings?.formulaMultiplier || 0.84,
    };
  }

  /**
   * Save analytics settings for TikTok
   */
  async saveSettings(settings: Partial<AnalyticsSettingsData>): Promise<void> {
    const existingSettings = await this.prisma.analyticsSettings.findFirst({
      where: {
        tenantId: this.tenantId,
        platform: "tiktok",
      },
    });

    if (existingSettings) {
      await this.prisma.analyticsSettings.update({
        where: { id: existingSettings.id },
        data: settings,
      });
    } else {
      await this.prisma.analyticsSettings.create({
        data: {
          tenantId: this.tenantId,
          platform: "tiktok",
          priceColumn: settings.priceColumn || "HARGA",
          formulaDeduction: settings.formulaDeduction || 1500,
          formulaMultiplier: settings.formulaMultiplier || 0.84,
        },
      });
    }
  }

  /**
   * Analyze price reconciliation for a month
   */
  async analyze(month: number, year: number): Promise<ReconciliationResult> {
    logger.info(
      `Analyzing TikTok ${year}-${month} for tenant ${this.tenantId}`
    );

    const settings = await this.getSettings();

    // Get all TikTok escrow items for the month
    // Exclude problematic orders (refunds/cancels with negative settlement)
    const items = await this.prisma.tiktokEscrowItem.findMany({
      where: {
        tenantId: this.tenantId,
        escrowOrder: {
          month,
          year,
          totalSettlementAmount: { gte: 0 }, // Exclude refund/cancel orders
        },
        // Exclude items with zero sale price (refund items)
        salePrice: { gt: 0 },
        // Exclude items with negative settlement (refund items)
        settlementAmount: { gte: 0 },
      },
      include: {
        escrowOrder: {
          include: {
            items: true,
          },
        },
      },
    });

    logger.info(
      `Found ${items.length} TikTok items (excluding refunds/cancels)`
    );

    // Build skuId to real sellerSku mapping from TiktokSku table
    const skuIdToRealSku = await this.buildSkuMapping(items);

    // Group by SKU (use real sellerSku from TiktokSku table, fallback to skuId)
    const skuMap = new Map<string, SkuData>();
    const skuInfo = new Map<string, SkuInfo>();

    for (const item of items) {
      // Get real SKU from TiktokSku table mapping, fallback to skuId
      const realSku = skuIdToRealSku.get(item.skuId) || item.skuId || "UNKNOWN";
      const variantName = item.sellerSku || ""; // Original sellerSku is actually variant name
      const quantity = item.quantity || 1;

      // Calculate per-unit sale price
      const unitPrice = Math.round(item.salePrice / quantity);

      // Only calculate actual income for SINGLE-ITEM orders
      const orderItemCount = item.escrowOrder.items?.length || 1;
      if (orderItemCount === 1) {
        const unitActualIncome = Math.round(
          item.escrowOrder.totalSettlementAmount / quantity
        );

        if (!skuMap.has(realSku)) {
          skuMap.set(realSku, {
            unitPrices: new Set(),
            actualIncomes: new Set(),
            count: 0,
          });
        }
        const skuData = skuMap.get(realSku)!;
        skuData.actualIncomes.add(unitActualIncome);
      }

      if (!skuInfo.has(realSku)) {
        skuInfo.set(realSku, {
          productName: item.productName || "Unknown",
          sellerSku: realSku,
          variantName: variantName,
        });
      }

      if (!skuMap.has(realSku)) {
        skuMap.set(realSku, {
          unitPrices: new Set(),
          actualIncomes: new Set(),
          count: 0,
        });
      }

      const skuData = skuMap.get(realSku)!;
      skuData.unitPrices.add(unitPrice);
      skuData.count++;
    }

    // Build SKU groups with inventory lookup
    const skuGroups: SkuGroup[] = [];

    for (const [sku, skuData] of skuMap) {
      const info = skuInfo.get(sku)!;

      const uniqueUnitPrices = Array.from(skuData.unitPrices).sort(
        (a, b) => a - b
      );
      const uniqueActualIncomes = Array.from(skuData.actualIncomes).sort(
        (a, b) => a - b
      );

      // Lookup inventory
      const inventory = await this.lookupInventory(sku, settings.priceColumn);
      const inventoryPrice = inventory?.price || null;
      const expectedIncome = inventoryPrice
        ? this.calculateExpectedIncome(inventoryPrice, settings)
        : null;

      // Determine status
      const hasMultiplePrices = uniqueUnitPrices.length > 1;
      const hasPriceDifference =
        inventoryPrice !== null &&
        uniqueUnitPrices.some((p) => p !== inventoryPrice);

      let status: SkuGroup["status"];
      if (inventoryPrice === null) {
        status = "NO_INVENTORY";
      } else if (hasPriceDifference || hasMultiplePrices) {
        status = "PRICE_DIFF";
      } else {
        status = "OK";
      }

      skuGroups.push({
        sku,
        sellerSku: info.sellerSku,
        productName: info.productName,
        variantName: info.variantName,
        inventoryPrice,
        expectedIncome,
        totalTransactions: skuData.count,
        uniqueUnitPrices,
        uniqueActualIncomes,
        hasMultiplePrices,
        hasPriceDifference,
        status,
      });
    }

    // Sort: problems first, then by transaction count
    skuGroups.sort((a, b) => {
      if (a.status !== "OK" && b.status === "OK") return -1;
      if (a.status === "OK" && b.status !== "OK") return 1;
      return b.totalTransactions - a.totalTransactions;
    });

    // Build summary
    const summary: ReconciliationSummary = {
      totalSku: skuGroups.length,
      totalTransactions: items.length,
      skuOk: skuGroups.filter((s) => s.status === "OK").length,
      skuWithPriceDiff: skuGroups.filter((s) => s.status === "PRICE_DIFF")
        .length,
      skuNoInventory: skuGroups.filter((s) => s.status === "NO_INVENTORY")
        .length,
    };

    return { summary, skuGroups };
  }

  /**
   * Build mapping from skuId to real sellerSku from TiktokSku table
   * TikTok Settlement API returns variant name in seller_sku field,
   * but TiktokSku table has the real SKU from Product API
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private async buildSkuMapping(items: any[]): Promise<Map<string, string>> {
    const skuIdToRealSku = new Map<string, string>();

    // Get unique skuIds from items
    const skuIds = [
      ...new Set(items.map((item) => item.skuId).filter(Boolean)),
    ];

    if (skuIds.length === 0) return skuIdToRealSku;

    // Lookup real sellerSku from TiktokSku table
    const tiktokSkus = await this.prisma.tiktokSku.findMany({
      where: {
        skuId: { in: skuIds },
      },
      select: {
        skuId: true,
        sellerSku: true,
      },
    });

    for (const sku of tiktokSkus) {
      if (sku.sellerSku) {
        skuIdToRealSku.set(sku.skuId, sku.sellerSku);
      }
    }

    logger.info(
      `Built SKU mapping: ${skuIdToRealSku.size}/${skuIds.length} skuIds mapped to real SKUs`
    );
    return skuIdToRealSku;
  }

  /**
   * Lookup inventory price for SKU
   */
  private async lookupInventory(
    sku: string,
    priceColumn: string
  ): Promise<{ price: number; name: string } | null> {
    try {
      const record = await this.prisma.inventoryRecord.findFirst({
        where: {
          tenantId: this.tenantId,
          keyValue: sku,
          keyColumnName: "SKU",
        },
      });

      if (!record) return null;

      const data = JSON.parse(record.data);
      const price = parseFloat(data[priceColumn] || "0");
      const name = data["Nama Barang"] || "Unknown";

      return price > 0 ? { price, name } : null;
    } catch {
      return null;
    }
  }

  /**
   * Calculate expected income from marketplace price
   */
  private calculateExpectedIncome(
    marketplacePrice: number,
    settings: AnalyticsSettingsData
  ): number {
    return (
      (marketplacePrice - settings.formulaDeduction) *
      settings.formulaMultiplier
    );
  }
}
