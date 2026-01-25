/**
 * Shopee Ads Types
 * Type definitions for Shopee Ads analytics
 * Single Responsibility: Types only
 */

// ============================================
// Upload Types
// ============================================

export interface UploadOptions {
  mode: "skip" | "update";
  uploadedBy?: string;
  periodLabel?: string;
}

export interface UploadBatchResult {
  success: boolean;
  batchId: string;
  fileName: string;
  periodStart: string;
  periodEnd: string;
  periodLabel: string;
  totalRows: number;
  insertedRows: number;
  skippedRows: number;
  updatedRows: number;
  errors: string[];
}

// ============================================
// Product Data Types
// ============================================

export interface ShopeeAdsProductRow {
  productId: string;
  productName: string;
  status: string | null;
  biddingMode: string | null;
  placement: string | null;
  startDate: Date | null;
  endDate: string | null;
  impressions: number;
  clicks: number;
  ctr: number;
  conversions: number;
  directConversions: number;
  conversionRate: number;
  directConversionRate: number;
  costPerConversion: number;
  costPerDirectConversion: number;
  unitsSold: number;
  directUnitsSold: number;
  revenue: number;
  directRevenue: number;
  cost: number;
  roas: number;
  directRoas: number;
  acos: number;
  directAcos: number;
  periodStart: Date;
  periodEnd: Date;
  periodLabel: string;
}

// ============================================
// Dashboard Types
// ============================================

export interface DashboardSummary {
  totalCost: number;
  totalRevenue: number;
  totalDirectRevenue: number;
  totalOrders: number;
  avgRoas: number;
  avgDirectRoas: number;
  totalImpressions: number;
  totalClicks: number;
  avgCtr: number;
  avgConversionRate: number;
  topProducts: ProductPerformance[];
  biddingModeComparison: BiddingModeStats[];
}

export interface ProductPerformance {
  productId: string;
  productName: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number;
}

export interface BiddingModeStats {
  biddingMode: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number;
  costPerOrder: number;
  productCount: number;
}

// ============================================
// Query Filter Types
// ============================================

export interface ProductDataFilter {
  periodStart?: string;
  periodEnd?: string;
  productId?: string;
  biddingMode?: string;
  minRoas?: number;
  maxRoas?: number;
  limit?: number;
  offset?: number;
  orderBy?: string;
  orderDir?: "asc" | "desc";
}

// ============================================
// Upload Batch Types
// ============================================

export interface UploadBatch {
  id: string;
  fileName: string;
  periodStart: Date;
  periodEnd: Date;
  totalRows: number;
  insertedRows: number;
  skippedRows: number;
  status: string;
  createdAt: Date;
}
