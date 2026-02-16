/**
 * Inventory API Module (Barrel File)
 *
 * This file re-exports all inventory-related functions from specialized modules
 * for backward compatibility. Existing imports like `import { getInventory } from '@/api/inventory'`
 * continue to work without changes.
 */

// Re-export all types
export type {
	BatchCheckResult,
	InventoryConfig,
	InventoryListResult,
	InventoryRecord,
	InventoryStats,
	SkuCheckResult,
	SyncHistoryEntry,
} from "@/types/inventory";
// Re-export check operations
export { batchCheckSku, checkPlatformStatus } from "./inventoryCheck";
// Re-export column operations
export {
	getAvailableColumns,
	getSelectedColumns,
	saveSelectedColumns,
} from "./inventoryColumns";
export type { RawInventoryConfig } from "./inventoryConfig";
// Re-export config operations
export {
	getInventoryConfig,
	normalizeInventoryConfig,
	normalizeSelectedColumns,
	updateInventoryConfig,
} from "./inventoryConfig";
export type { GetInventoryParams } from "./inventoryCore";
// Re-export core operations
export {
	getInventory,
	getInventoryBySku,
	getInventoryStats,
	getSyncHistory,
	updateInventoryRecord,
} from "./inventoryCore";
// Re-export price operations
export { updatePrice, updatePriceBatch } from "./inventoryPrice";
export type {
	BatchPriceUpdateResult,
	PriceUpdateItem,
	PriceUpdateResult,
} from "./inventoryPriceTypes";

// Re-export sync operations
export {
	syncInventory,
	syncToSheets,
	updateStock,
	updateStockBatch,
} from "./inventorySync";
