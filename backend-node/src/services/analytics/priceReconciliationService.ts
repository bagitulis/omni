/**
 * Price Reconciliation Service
 * Analyzes price differences between escrow and inventory
 * Single Responsibility: Price analysis
 */

import { getLogger } from "../../utils/logger";

const logger = getLogger("PriceReconciliationService");

// Use 'any' for Prisma client to avoid type issues with dynamically added models
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PrismaClientAny = any;

interface SkuGroup {
  sku: string;
  modelSku: string;
  itemName: string;
  modelName: string;             // Variant name (e.g., "Hitam")
  inventoryPrice: number | null;
  expectedIncome: number | null;
  totalTransactions: number;
  uniqueUnitPrices: number[];    // List of unique unit marketplace prices
  uniqueActualIncomes: number[]; // List of unique actual incomes per unit
  hasMultiplePrices: boolean;
  hasPriceDifference: boolean;
  status: "OK" | "PRICE_DIFF" | "NO_INVENTORY";
}

interface ReconciliationSummary {
  totalSku: number;
  totalTransactions: number;
  skuOk: number;
  skuWithPriceDiff: number;
  skuNoInventory: number;
}

interface ReconciliationResult {
  summary: ReconciliationSummary;
  skuGroups: SkuGroup[];
}

interface AnalyticsSettingsData {
  priceColumn: string;
  formulaDeduction: number;
  formulaMultiplier: number;
}

export class PriceReconciliationService {
  private prisma: PrismaClientAny;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  /**
   * Get analytics settings
   */
  async getSettings(): Promise<AnalyticsSettingsData> {
    const settings = await this.prisma.analyticsSettings.findUnique({
      where: {
        tenantId_platform: {
          tenantId: this.tenantId,
          platform: "shopee",
        },
      },
    });

    return {
      priceColumn: settings?.priceColumn || "HARGA",
      formulaDeduction: settings?.formulaDeduction || 1500,
      formulaMultiplier: settings?.formulaMultiplier || 0.84,
    };
  }

  /**
   * Save analytics settings
   */
  async saveSettings(settings: Partial<AnalyticsSettingsData>): Promise<void> {
    await this.prisma.analyticsSettings.upsert({
      where: {
        tenantId_platform: {
          tenantId: this.tenantId,
          platform: "shopee",
        },
      },
      create: {
        tenantId: this.tenantId,
        platform: "shopee",
        priceColumn: settings.priceColumn || "HARGA",
        formulaDeduction: settings.formulaDeduction || 1500,
        formulaMultiplier: settings.formulaMultiplier || 0.84,
      },
      update: settings,
    });
  }

  /**
   * Analyze price reconciliation for a month
   */
  async analyze(month: number, year: number): Promise<ReconciliationResult> {
    logger.info(`Analyzing ${year}-${month} for tenant ${this.tenantId}`);

    const settings = await this.getSettings();

    // Get all escrow items for the month with order details
    const items = await this.prisma.shopeeEscrowItem.findMany({
      where: {
        tenantId: this.tenantId,
        escrowOrder: {
          month,
          year,
        },
      },
      include: {
        escrowOrder: {
          include: {
            items: true, // Include all items in the order to count them
          },
        },
      },
    });

    logger.info(`Found ${items.length} items`);

    // Group by SKU, collect unique unit prices and actual incomes
    const skuMap = new Map<
      string,
      { unitPrices: Set<number>; actualIncomes: Set<number>; count: number }
    >();
    const skuInfo = new Map<
      string,
      { itemName: string; modelSku: string; modelName: string }
    >();

    for (const item of items) {
      const sku = item.modelSku || item.sku || "UNKNOWN";
      const quantity = item.quantity || 1;

      // Calculate per-unit marketplace price
      const unitPrice = Math.round(item.originalPrice / quantity);

      // Only calculate actual income for SINGLE-ITEM orders
      // This ensures accurate per-unit escrow calculation
      const orderItemCount = item.escrowOrder.items?.length || 1;
      if (orderItemCount === 1) {
        const unitActualIncome = Math.round(
          item.escrowOrder.escrowAmount / quantity
        );

        if (!skuMap.has(sku)) {
          skuMap.set(sku, {
            unitPrices: new Set(),
            actualIncomes: new Set(),
            count: 0,
          });
        }
        const skuData = skuMap.get(sku)!;
        skuData.actualIncomes.add(unitActualIncome);
      }

      if (!skuInfo.has(sku)) {
        skuInfo.set(sku, {
          itemName: item.itemName || "Unknown",
          modelSku: item.modelSku || "",
          modelName: item.modelName || "",
        });
      }

      if (!skuMap.has(sku)) {
        skuMap.set(sku, {
          unitPrices: new Set(),
          actualIncomes: new Set(),
          count: 0,
        });
      }

      const skuData = skuMap.get(sku)!;
      skuData.unitPrices.add(unitPrice);
      skuData.count++;
    }

    // Build SKU groups with inventory lookup
    const skuGroups: SkuGroup[] = [];

    for (const [sku, skuData] of skuMap) {
      const info = skuInfo.get(sku)!;

      // Convert Sets to sorted arrays
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
        modelSku: info.modelSku,
        itemName: info.itemName,
        modelName: info.modelName,
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
