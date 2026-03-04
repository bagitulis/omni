/**
 * Wholesale Types
 * Interfaces for wholesale operations across platforms
 */

export interface WholesaleTier {
  minCount: number;
  maxCount: number;
  unitPrice: number;
}

export interface WholesaleInfo {
  itemId: number;
  productName: string;
  hasWholesale: boolean;
  tiers: WholesaleTier[];
}

export interface WholesaleResult {
  success: boolean;
  itemId?: number;
  message?: string;
  error?: string;
  data?: any;
}

export interface BatchDeleteBySkusResult {
  success: boolean;
  message: string;
  data: BatchDeleteData;
}

export interface BatchDeleteData {
  totalSkus: number;
  uniqueItems: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: WholesaleResult[];
}

export interface SkuLookupResult {
  itemId: string;
  modelId: string | null;
  sellerSku: string | null;
  productName?: string;
}

export interface WholesaleSettings {
  tenantId: string;
  platform: string;
  adminFee: number;
  minOrder1: number;
  maxOrder1: number;
  maxOrderTier3: number;
}

export interface WholesaleTierCalculated {
  tier: number;
  minCount: number;
  maxCount: number;
  unitPrice: number;
}

export interface BatchUpdateItem {
  sku: string;
  price: number;
}

export interface BatchUpdateBySkusResult {
  success: boolean;
  message: string;
  data: BatchUpdateData;
}

export interface BatchUpdateData {
  totalSkus: number;
  uniqueItems: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: WholesaleResult[];
  settingsUsed: WholesaleSettings;
}

export interface BatchMpqResult {
  success: boolean;
  message: string;
  data: BatchMpqData;
}

export interface BatchMpqData {
  totalSkus: number;
  uniqueItems: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: MpqItemResult[];
}

export interface MpqItemResult {
  success: boolean;
  itemId: number;
  error?: string;
  message?: string;
}

export interface TiktokBatchMpqResult {
  success: boolean;
  message: string;
  data: TiktokBatchMpqData;
}

export interface TiktokBatchMpqData {
  totalSkus: number;
  uniqueProducts: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: TiktokMpqItemResult[];
}

export interface TiktokMpqItemResult {
  success: boolean;
  productId: string;
  error?: string;
  message?: string;
}

export interface TierPreviewResult {
  basePrice: number;
  settings: WholesaleSettings;
  tiers: WholesaleTierCalculated[];
}

export interface BatchWholesaleResetResult {
  success: boolean;
  data: any;
  message?: string;
}
