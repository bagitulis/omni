import { useBatchSkuCheck } from "@/composables/useBatchSkuCheck";
import { getAuthHeaders, getApiBaseUrl } from "../../../../utils/apiHeaders";

export function useInventoryBatchCheck(state: any, onAlert: any) {
  const extractSkuFromItem = (item: any): string => {
    if (!item) return "";

    // Handle nested data field - supports both string (Node.js) and object (Go) formats
    let itemData = item;
    if (item.data !== undefined && item.data !== null) {
      if (typeof item.data === "string") {
        try {
          itemData = JSON.parse(item.data);
        } catch {
          itemData = item;
        }
      } else if (typeof item.data === "object") {
        itemData = item.data;
      }
    }

    const skuFields = [
      "sku",
      "SKU",
      "Sku",
      "sku_id",
      "SKU_ID",
      "Sku_Id",
      "seller_sku",
      "sellerSku",
      "SellerSku",
      "product_sku",
      "productSku",
      "ProductSku",
    ];

    for (const field of skuFields) {
      const value = itemData[field];
      if (value) {
        const skuValue = String(value).trim();
        if (skuValue.length > 0) return skuValue;
      }
    }

    return "";
  };

  const fetchAllInventoryFromSheets = async (): Promise<any[]> => {
    try {
      const response = await fetch(getApiBaseUrl("/google/sheets/data"), {
        method: "GET",
        headers: getAuthHeaders(),
      });

      if (!response.ok) {
        console.warn(
          "⚠️ Google Sheets API failed, falling back to component data",
        );
        return state.inventoryList || [];
      }

      const data = await response.json();
      if (data.success && Array.isArray(data.data)) {
        console.log(`✅ Fetched ${data.data.length} items from Google Sheets`);
        return data.data;
      }

      return state.inventoryList || [];
    } catch (error) {
      console.error("⚠️ Error fetching from Google Sheets:", error);
      return state.inventoryList || [];
    }
  };

  const loadSavedPlatformStatus = async () => {
    try {
      const { loadSavedStatus } = useBatchSkuCheck();
      const savedResults = await loadSavedStatus();
      if (savedResults && savedResults.length > 0) {
        state.batchCheckResults = savedResults;
      }
    } catch {
      // Not critical
    }
  };

  const handleBatchCheckPlatform = async () => {
    try {
      state.batchCheckLoading = true;

      const allData = await fetchAllInventoryFromSheets();
      if (!allData || allData.length === 0) {
        onAlert.warning("⚠️ No Data", "Tidak ada data untuk di-check");
        return;
      }

      const skus = allData
        .map((item) => extractSkuFromItem(item))
        .filter((sku) => sku && sku.length > 0)
        .filter((sku, index, self) => self.indexOf(sku) === index);

      if (skus.length === 0) {
        onAlert.warning("⚠️ No SKU Found", "Tidak ada SKU yang ditemukan");
        return;
      }

      onAlert.info(
        "⏳ Checking",
        `Batch checking ${skus.length} unique SKU(s) dari ${allData.length} items...`,
      );

      const { checkSkus, saveResults } = useBatchSkuCheck();
      const results = await checkSkus(skus);

      if (results.length > 0) {
        state.batchCheckResults = results;

        onAlert.info("💾 Saving", "Menyimpan hasil check ke database...");
        const saveSucess = await saveResults(results);

        if (saveSucess) {
          onAlert.success(
            "✅ Platform Check & Save Complete",
            `Checked & saved ${results.length} SKU(s) from ${allData.length} items`,
          );
        } else {
          onAlert.warning(
            "⚠️ Check Complete, Save Failed",
            `Checked ${results.length} SKU(s) but failed to save to database`,
          );
        }
      } else {
        onAlert.error("❌ Check Failed", "Failed to check SKUs");
      }
    } catch (error) {
      onAlert.error("❌ Error", String(error));
    } finally {
      state.batchCheckLoading = false;
    }
  };

  return {
    extractSkuFromItem,
    fetchAllInventoryFromSheets,
    loadSavedPlatformStatus,
    handleBatchCheckPlatform,
  };
}
