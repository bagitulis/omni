/**
 * Inventory Wholesale Operations Composable
 * Handles wholesale, MPQ, and related modal operations
 */

import { extractSKU } from "./useInventoryItemHelpers";
import { getAllocationForItem } from "./useMarketplaceAllocation";
import { useMarketplaceSettings } from "./useMarketplaceSettings";
import { useInventoryAlerts } from "./useInventoryAlerts";

export interface WholesaleItem {
  sku: string;
  price: number;
  platform?: string;
}

interface InventoryItem {
  data?: string | Record<string, any>;
  checkbox?: boolean | string | number;
  selected?: boolean;
  [key: string]: any;
}

interface WholesaleState {
  showWholesaleModal: boolean;
  wholesaleSkus: string[];
  showWholesaleUpdateModal: boolean;
  wholesaleUpdateItems: WholesaleItem[];
  showWholesaleMpqModal: boolean;
  wholesaleMpqItems: WholesaleItem[];
}

/**
 * Check if item is selected/checked
 */
function isItemChecked(item: InventoryItem): boolean {
  return (
    item.checkbox === true ||
    item.checkbox === "true" ||
    item.checkbox === 1 ||
    item.selected === true
  );
}

/**
 * Extract HARGA (price) from item
 */
function getHargaFromItem(item: InventoryItem): number {
  try {
    // Handle both string (Node.js) and object (Go) formats
    let data = item;
    if (item.data !== undefined && item.data !== null) {
      if (typeof item.data === "string") {
        data = JSON.parse(item.data);
      } else if (typeof item.data === "object") {
        data = item.data;
      }
    }
    const priceKeys = ["HARGA", "harga", "Harga", "price", "Price", "PRICE"];

    for (const key of priceKeys) {
      if (data[key] !== undefined && data[key] !== null) {
        const price = parseFloat(String(data[key]).replace(/[^0-9.]/g, ""));
        if (!isNaN(price) && price > 0) return price;
      }
    }
    return 0;
  } catch {
    return 0;
  }
}

/**
 * Extract platform from item data (legacy - kept for backward compat)
 */
export function getPlatformFromItem(item: InventoryItem): string {
  try {
    // Handle both string (Node.js) and object (Go) formats
    let data = item;
    if (item.data !== undefined && item.data !== null) {
      if (typeof item.data === "string") {
        data = JSON.parse(item.data);
      } else if (typeof item.data === "object") {
        data = item.data;
      }
    }
    const platformKeys = ["PLATFORM", "platform", "Platform"];

    for (const key of platformKeys) {
      if (data[key]) {
        const platform = String(data[key]).toLowerCase();
        if (platform.includes("shopee")) return "shopee";
        if (platform.includes("tiktok")) return "tiktok";
        if (platform.includes("lazada")) return "lazada";
        return platform;
      }
    }
    return "unknown";
  } catch {
    return "unknown";
  }
}

export function useInventoryWholesale(state: WholesaleState) {
  const alertMethods = useInventoryAlerts();

  /**
   * Handle delete wholesale - extract SKUs from checked items
   */
  function handleDeleteWholesale(inventoryList: InventoryItem[]): void {
    if (!inventoryList || inventoryList.length === 0) {
      alertMethods.showAlertWarning(
        "⚠️ No Data",
        "Inventory list is empty. Please sync data first.",
      );
      return;
    }

    const checkedSkus = inventoryList
      .filter((item) => isItemChecked(item))
      .map((item) => extractSKU(item))
      .filter((sku) => sku && sku.length > 0);

    if (checkedSkus.length === 0) {
      alertMethods.showAlertWarning(
        "⚠️ No Items Selected",
        "Please check/select items to delete wholesale tiers",
      );
      return;
    }

    state.wholesaleSkus = checkedSkus;
    state.showWholesaleModal = true;
  }

  /**
   * Handle wholesale delete completed
   */
  function handleWholesaleCompleted(result: any): void {
    if (result.success) {
      alertMethods.showAlertSuccess(
        "✅ Wholesale Deleted",
        `Successfully deleted wholesale for ${result.data.processed} products`,
      );
    } else if (result.data.processed > 0) {
      alertMethods.showAlertWarning(
        "⚠️ Partial Success",
        `Deleted ${result.data.processed}/${result.data.uniqueItems} products. ${result.data.failed} failed.`,
      );
    } else {
      alertMethods.showAlertError(
        "❌ Delete Failed",
        result.message || "Failed to delete wholesale tiers",
      );
    }
  }

  /**
   * Open wholesale update modal for selected items
   */
  function handleUpdateWholesale(filteredList: InventoryItem[]): void {
    if (!filteredList || filteredList.length === 0) {
      alertMethods.showAlertWarning(
        "⚠️ No Data",
        "Inventory list is empty. Please sync data first.",
      );
      return;
    }

    const checkedItems = filteredList
      .filter((item) => isItemChecked(item))
      .map((item) => ({
        sku: extractSKU(item),
        price: getHargaFromItem(item),
      }))
      .filter((item) => item.sku && item.sku.length > 0 && item.price > 0);

    if (checkedItems.length === 0) {
      alertMethods.showAlertWarning(
        "⚠️ No Valid Items",
        "Please check items with valid SKU and HARGA (price > 0)",
      );
      return;
    }

    state.wholesaleUpdateItems = checkedItems;
    state.showWholesaleUpdateModal = true;
  }

  /**
   * Handle wholesale update completed
   */
  function handleWholesaleUpdateCompleted(result: any): void {
    if (result.success) {
      alertMethods.showAlertSuccess(
        "✅ Wholesale Updated",
        `Successfully updated wholesale for ${result.data.processed} products`,
      );
    } else if (result.data.processed > 0) {
      alertMethods.showAlertWarning(
        "⚠️ Partial Success",
        `Updated ${result.data.processed}/${result.data.uniqueItems} products. ${result.data.failed} failed.`,
      );
    } else {
      alertMethods.showAlertError(
        "❌ Update Failed",
        result.message || "Failed to update wholesale tiers",
      );
    }
  }

  /**
   * Open MPQ modal with marketplace allocation
   */
  function handleOpenMpqModal(filteredList: InventoryItem[]): void {
    if (!filteredList || filteredList.length === 0) {
      alertMethods.showAlertWarning(
        "⚠️ No Data",
        "Inventory list is empty. Please sync data first.",
      );
      return;
    }

    const { totalColumn, autoColumn } = useMarketplaceSettings();
    const expandedItems: WholesaleItem[] = [];

    filteredList
      .filter((item) => isItemChecked(item))
      .forEach((item) => {
        const sku = extractSKU(item);
        const price = getHargaFromItem(item);

        if (!sku || sku.length === 0 || price <= 0) return;

        const allocation = getAllocationForItem(
          item,
          totalColumn.value,
          autoColumn.value,
        );

        // Add entry for each platform that has allocation > 0
        if (allocation.shopee > 0) {
          expandedItems.push({ sku, price, platform: "shopee" });
        }
        if (allocation.tiktok > 0) {
          expandedItems.push({ sku, price, platform: "tiktok" });
        }
        if (allocation.lazada > 0) {
          expandedItems.push({ sku, price, platform: "lazada" });
        }

        // Default to shopee if no allocation
        if (
          allocation.shopee === 0 &&
          allocation.tiktok === 0 &&
          allocation.lazada === 0
        ) {
          expandedItems.push({ sku, price, platform: "shopee" });
        }
      });

    if (expandedItems.length === 0) {
      alertMethods.showAlertWarning(
        "⚠️ No Valid Items",
        "Please check items with valid SKU and HARGA (price > 0)",
      );
      return;
    }

    state.wholesaleMpqItems = expandedItems;
    state.showWholesaleMpqModal = true;
  }

  /**
   * Handle MPQ modal completed
   */
  function handleWholesaleMpqCompleted(result: any): void {
    if (result.success) {
      alertMethods.showAlertSuccess(
        "✅ Update Complete",
        "Bulk pricing update completed successfully",
      );
    } else {
      alertMethods.showAlertError(
        "❌ Update Failed",
        result.message || "Failed to update bulk pricing",
      );
    }
  }

  return {
    handleDeleteWholesale,
    handleWholesaleCompleted,
    handleUpdateWholesale,
    handleWholesaleUpdateCompleted,
    handleOpenMpqModal,
    handleWholesaleMpqCompleted,
    getHargaFromItem,
    getPlatformFromItem,
  };
}
