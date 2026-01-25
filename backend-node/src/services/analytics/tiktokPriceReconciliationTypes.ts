/**
 * TikTok Price Reconciliation Types
 * Type definitions for price reconciliation analysis
 */

export interface SkuGroup {
  sku: string;
  sellerSku: string;
  productName: string;
  variantName: string;
  inventoryPrice: number | null;
  expectedIncome: number | null;
  totalTransactions: number;
  uniqueUnitPrices: number[];
  uniqueActualIncomes: number[];
  hasMultiplePrices: boolean;
  hasPriceDifference: boolean;
  status: "OK" | "PRICE_DIFF" | "NO_INVENTORY";
}

export interface ReconciliationSummary {
  totalSku: number;
  totalTransactions: number;
  skuOk: number;
  skuWithPriceDiff: number;
  skuNoInventory: number;
}

export interface ReconciliationResult {
  summary: ReconciliationSummary;
  skuGroups: SkuGroup[];
}

export interface AnalyticsSettingsData {
  priceColumn: string;
  formulaDeduction: number;
  formulaMultiplier: number;
}

export interface SkuData {
  unitPrices: Set<number>;
  actualIncomes: Set<number>;
  count: number;
}

export interface SkuInfo {
  productName: string;
  sellerSku: string;
  variantName: string;
}
