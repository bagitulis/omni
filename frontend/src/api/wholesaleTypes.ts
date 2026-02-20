/**
 * Wholesale API Types
 * Separated from API functions per SRP principle.
 */

import type { WholesaleSettingsApi } from "./wholesaleHelpers";

export interface WholesaleTier {
  min_count: number;
  max_count: number;
  unit_price: number;
}

export interface WholesaleInfo {
  item_id: number;
  tiers: WholesaleTier[];
}

export interface WholesaleResult {
  success: boolean;
  error?: string;
  data?: unknown;
}

export type WholesaleSettings = WholesaleSettingsApi;

export interface WholesaleTierCalculated {
  tier: number;
  min_count: number;
  max_count: number;
  unit_price: number;
}

export interface SkuLookupResult {
  item_id: number;
  sku: string;
}

/** Shared shape for batch operation failure data */
interface BatchFailureData {
  total_skus: number;
  unique_items: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: unknown[];
}

export interface BatchDeleteBySkusResult {
  success: boolean;
  message: string;
  data: BatchFailureData;
}

export interface BatchUpdateBySkusResult {
  success: boolean;
  message: string;
  data: BatchFailureData & {
    settings_used: WholesaleSettings;
  };
}

export interface BatchMpqResult {
  success: boolean;
  data?: unknown;
  error?: string;
}

export interface TiktokBatchMpqResult {
  success: boolean;
  data?: unknown;
  error?: string;
}

export interface TierPreviewResult {
  tiers: WholesaleTierCalculated[];
}

export interface BatchWholesaleResetResult {
  success: boolean;
  data?: unknown;
  error?: string;
}
