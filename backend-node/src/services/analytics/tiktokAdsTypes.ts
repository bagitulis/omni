/**
 * TikTok Ads Analytics Types
 * Type definitions for TikTok Ads creative data
 * Single Responsibility: Type definitions only
 */

// ============================================
// Upload & Batch Types
// ============================================

export interface UploadBatchResult {
  success: boolean;
  batchId: string;
  fileName: string;
  periodStart: Date;
  periodEnd: Date;
  totalRows: number;
  insertedRows: number;
  skippedRows: number;
  updatedRows: number;
  errors: string[];
}

export interface UploadOptions {
  mode: "skip" | "update"; // skip = skip duplicates, update = replace duplicates
  tenantId: string;
  uploadedBy?: string;
}

// ============================================
// Creative Data Types (from Excel)
// ============================================

export interface RawCreativeData {
  campaignName: string;
  campaignId: string;
  productId: string;
  creativeType: string;
  videoTitle: string | null;
  videoId: string | null;
  tiktokAccount: string | null;
  postingTime: Date | null;
  status: string | null;
  authorizationType: string | null;
  cost: number;
  ordersSku: number;
  costPerOrder: number;
  grossRevenue: number;
  roi: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversionRate: number;
  watchRate2s: number | null;
  watchRate6s: number | null;
  watchRate25pct: number | null;
  watchRate50pct: number | null;
  watchRate75pct: number | null;
  watchRate100pct: number | null;
  currency: string;
}

export interface CreativeDataWithPeriod extends RawCreativeData {
  periodStart: Date;
  periodEnd: Date;
}

// ============================================
// Summary & Aggregation Types
// ============================================

export interface ProductSummary {
  productId: string;
  periodType: "daily" | "weekly" | "monthly";
  periodDate: Date;
  totalCost: number;
  totalOrders: number;
  totalRevenue: number;
  avgRoi: number;
  avgCtr: number;
  avgConversionRate: number;
  totalImpressions: number;
  totalClicks: number;
  bestCreativeType: string | null;
  topVideoId: string | null;
  topVideoTitle: string | null;
  roiTrend: "up" | "down" | "stable" | null;
  costTrend: "up" | "down" | "stable" | null;
  ordersTrend: "up" | "down" | "stable" | null;
}

// ============================================
// ML Prediction Types
// ============================================

export interface MLPrediction {
  productId: string;
  targetMonth: number;
  targetYear: number;
  predictedRoi: number | null;
  roiConfidence: number | null;
  recommendedBudget: number | null;
  budgetReason: string | null;
  performanceScore: "A" | "B" | "C" | "D" | "F" | null;
  performanceReason: string | null;
  creativeScore: number | null;
  bestCreativeType: string | null;
  insights: string[];
  hasAnomaly: boolean;
  anomalyType: string | null;
  anomalyDescription: string | null;
}

export interface MLTrainingData {
  productId: string;
  month: number;
  year: number;
  cost: number;
  orders: number;
  revenue: number;
  roi: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversionRate: number;
  creativeType: string;
}

// ============================================
// Query & Filter Types
// ============================================

export interface CreativeDataFilter {
  tenantId: string;
  periodStart?: Date;
  periodEnd?: Date;
  campaignId?: string;
  productId?: string;
  creativeType?: string;
  minRoi?: number;
  maxRoi?: number;
  limit?: number;
  offset?: number;
  orderBy?: "cost" | "revenue" | "roi" | "orders";
  orderDir?: "asc" | "desc";
}

export interface DashboardSummary {
  totalCost: number;
  totalRevenue: number;
  totalOrders: number;
  avgRoi: number;
  totalImpressions: number;
  totalClicks: number;
  avgCtr: number;
  avgConversionRate: number;
  topProducts: ProductPerformance[];
  creativeTypeComparison: CreativeTypeStats[];
  periodTrend: PeriodTrend[];
}

export interface ProductPerformance {
  productId: string;
  cost: number;
  revenue: number;
  orders: number;
  roi: number;
}

export interface CreativeTypeStats {
  creativeType: string;
  cost: number;
  revenue: number;
  orders: number;
  roi: number;
  costPerOrder: number;
}

export interface PeriodTrend {
  period: string;
  cost: number;
  revenue: number;
  orders: number;
  roi: number;
}

// ============================================
// Excel Column Mapping
// ============================================

export const EXCEL_COLUMN_MAP = {
  "Nama kampanye": "campaignName",
  "ID Campaign": "campaignId",
  "ID produk": "productId",
  "Jenis materi iklan": "creativeType",
  "Judul video": "videoTitle",
  "ID video": "videoId",
  "Akun TikTok": "tiktokAccount",
  "Waktu posting": "postingTime",
  Status: "status",
  "Jenis otorisasi": "authorizationType",
  Biaya: "cost",
  "Pesanan SKU": "ordersSku",
  "Biaya per pesanan": "costPerOrder",
  "Pendapatan kotor": "grossRevenue",
  ROI: "roi",
  "Impresi iklan produk": "impressions",
  "Jumlah klik iklan produk": "clicks",
  "Tingkat klik iklan produk": "ctr",
  "Rasio konversi iklan": "conversionRate",
  "Rasio tayang video iklan 2 detik": "watchRate2s",
  "Rasio tayang video iklan 6 detik": "watchRate6s",
  "Rasio tayang video iklan 25%": "watchRate25pct",
  "Rasio tayang video iklan 50%": "watchRate50pct",
  "Rasio tayang video iklan 75%": "watchRate75pct",
  "Rasio tayang video iklan 100%": "watchRate100pct",
  "Mata uang": "currency",
} as const;

export type ExcelColumnKey = keyof typeof EXCEL_COLUMN_MAP;
