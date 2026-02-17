/**
 * Re-export shared types for component-level imports
 * Convenience bridge to avoid long import paths
 */

export type {
  // Platform types
  Platform,
  PlatformSyncState,
  PlatformLinkStatus,
  PlatformIndicatorData,
  // Inline edit types
  InlineEditMode,
  InlineEditCellProps,
  // Batch action types
  BatchActionType,
  BatchActionItem,
  // Column manager types
  ColumnConfig,
  ColumnManagerProps,
  // Filter types
  ProductFilterValues,
  // Stock sync types
  StockSyncMode,
  StockSyncItem,
  // Unified product types
  UnifiedProductRow,
  // Image gallery types
  GalleryImage,
  GalleryPaginationMeta,
  // Marketplace sync history types
  SyncOperation,
  SyncResultStatus,
  MarketplaceSyncHistoryEntry,
  MarketplaceSyncHistoryListResult,
  MarketplaceSyncHistoryFilter,
} from "../../types/shared";

export { getImageUrl } from "../../types/shared";
