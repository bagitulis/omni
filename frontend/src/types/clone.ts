/**
 * Product Clone Types
 * API types use snake_case to match backend JSON response
 */

// =============================================================================
// Clone Request/Response Types
// =============================================================================

/**
 * Single product clone request
 */
export interface CloneRequest {
  source_platform: string;
  target_platform: string;
  source_item_id: string;
  sku?: string;
  category_id?: string;
  update_price?: boolean;
  new_price?: number;
  price_adjustment_type?: "fixed" | "percentage" | "amount";
  price_adjustment_value?: number;
  use_inventory?: boolean;
  save_as_draft?: boolean;
}

/**
 * Batch clone request
 */
export interface BatchCloneRequest {
  source_platform: string;
  target_platform: string;
  source_item_ids: string[];
  category_id?: string;
  price_adjustment_type?: "fixed" | "percentage" | "amount";
  price_adjustment_value?: number;
  save_as_draft?: boolean;
}

/**
 * Clone operation result
 */
export interface CloneResult {
  id: string;
  source_platform: string;
  target_platform: string;
  source_item_id: string;
  target_item_id?: string;
  status: string;
  message?: string;
  progress: number;
  started_at: string;
  completed_at?: string;
  sync_triggered?: boolean;
  sync_result?: string;
}

/**
 * Batch clone result
 */
export interface BatchCloneResult {
  batch_id: string;
  total_requested: number;
  total_success: number;
  total_failed: number;
  status: string;
  results: CloneResult[];
}

/**
 * Platform status for a SKU
 */
export interface PlatformStatus {
  shopee: boolean;
  lazada: boolean;
  tiktok: boolean;
}

/**
 * Available clone targets for a SKU
 */
export interface CloneTargetsResult {
  sku: string;
  status: PlatformStatus;
  sources: string[];
  targets: string[];
}

// =============================================================================
// Conflict Detection Types
// =============================================================================

/**
 * Product summary for comparison
 */
export interface ProductSummary {
  platform: string;
  item_id: string;
  sku: string;
  name: string;
  price: number;
  stock: number;
  images: string[];
}

/**
 * Single field difference
 */
export interface Difference {
  field: string;
  source_value: string;
  target_value: string;
}

/**
 * Adjustment information
 */
export interface AdjustmentInfo {
  title_will_truncate: boolean;
  desc_will_truncate: boolean;
  original_title?: string;
  adjusted_title?: string;
  title_limit: number;
  desc_limit: number;
}

/**
 * Clone conflict detection result
 */
export interface ConflictResult {
  has_conflict: boolean;
  source_product?: ProductSummary;
  target_product?: ProductSummary;
  differences?: Difference[];
  adjustments?: AdjustmentInfo;
}

// =============================================================================
// Product Data Types
// =============================================================================

/**
 * Product variant for cloning
 */
export interface ProductVariant {
  sku: string;
  name: string;
  price: number;
  stock: number;
}

/**
 * Product data for cloning
 */
export interface ProductData {
  item_id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  category_id: string;
  images: string[];
  variants?: ProductVariant[];
  attributes?: Record<string, string>;
  save_as_draft?: boolean;
}

// =============================================================================
// Response Wrappers
// =============================================================================

/**
 * Wrapper for clone operation response
 */
export interface CloneResponse {
  success: boolean;
  data?: CloneResult;
  error?: string;
}

/**
 * Wrapper for batch clone response
 */
export interface BatchCloneResponse {
  success: boolean;
  data?: BatchCloneResult;
  error?: string;
}

/**
 * Wrapper for clone status response
 */
export interface CloneStatusResponse {
  success: boolean;
  data?: CloneResult;
  error?: string;
}

/**
 * Wrapper for product data response
 */
export interface ProductDataResponse {
  success: boolean;
  product?: ProductData;
  error?: string;
}

/**
 * Wrapper for available targets response
 */
export interface AvailableTargetsResponse {
  success: boolean;
  sku?: string;
  status?: PlatformStatus;
  sources?: string[];
  targets?: string[];
  error?: string;
}

/**
 * Wrapper for preview/conflict response
 */
export interface PreviewResponse {
  success: boolean;
  data?: ConflictResult;
  error?: string;
}
