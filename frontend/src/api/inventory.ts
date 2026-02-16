/**
 * Inventory API Module (Barrel File)
 *
 * This file re-exports all inventory-related functions from specialized modules
 * for backward compatibility. Existing imports like `import { getInventory } from '@/api/inventory'`
 * continue to work without changes.
 */

// Types
export type {
	BatchCheckResult,
	InventoryConfig,
	InventoryListResult,
	InventoryRecord,
	InventoryStats,
	SkuCheckResult,
	SyncHistoryEntry,
} from "@/types/inventory";
// Check operations
export { batchCheckSku, checkPlatformStatus } from "./inventoryCheck";
// Column operations
export {
	getAvailableColumns,
	getSelectedColumns,
	saveSelectedColumns,
} from "./inventoryColumns";
export type { RawInventoryConfig } from "./inventoryConfig";
// Config operations
export {
	getInventoryConfig,
	normalizeInventoryConfig,
	normalizeSelectedColumns,
	updateInventoryConfig,
} from "./inventoryConfig";
export type { GetInventoryParams } from "./inventoryCore";
// Core operations
export {
	getInventory,
	getInventoryBySku,
	getInventoryStats,
	getSyncHistory,
	updateInventoryRecord,
} from "./inventoryCore";
// Price operations
export { updatePrice, updatePriceBatch } from "./inventoryPrice";
export type {
	BatchPriceUpdateResult,
	PriceUpdateItem,
	PriceUpdateResult,
} from "./inventoryPriceTypes";
// Sync operations
export {
	syncInventory,
	syncToSheets,
	updateStock,
	updateStockBatch,
} from "./inventorySync";
