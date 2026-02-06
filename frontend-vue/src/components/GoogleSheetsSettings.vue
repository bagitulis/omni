<template>
  <div class="google-sheets-settings">
    <div class="settings-header">
      <h2>🔗 Google Sheets Configuration</h2>
      <p>
        Link your Google Sheets spreadsheets for inventory, wallet, shipping,
        and orders
      </p>
    </div>

    <SpreadsheetLinkField
      type="inventory"
      label="📦 Inventory Spreadsheet"
      helper-text="Data inventory dan stok barang"
      v-model="links.inventory"
      :validating="validating.inventory"
      :validated="validated.inventory"
      :error="errors.inventory"
      :metadata="metadata.inventory"
      :selected-sheet="inventorySheet"
      :locked="locked.inventory"
      @check="checkLink('inventory')"
      @lock="lockLink('inventory')"
      @select-sheet="
        (sheet) =>
          updateInventorySheet(
            sheet,
            metadata.inventory?.spreadsheetId || '',
            metadata.inventory?.sheets || []
          )
      "
      @unlock="unlock('inventory')"
    />

    <SpreadsheetLinkField
      type="wallet"
      label="💰 Wallet Report Spreadsheet"
      helper-text="Spreadsheet untuk menyimpan laporan wallet transactions dari Shopee"
      v-model="links.wallet"
      :validating="validating.wallet"
      :validated="validated.wallet"
      :error="errors.wallet"
      :metadata="metadata.wallet"
      :locked="locked.wallet"
      @check="checkLink('wallet')"
      @lock="lockLink('wallet')"
      @unlock="unlock('wallet')"
    />

    <SpreadsheetLinkField
      type="shipping"
      label="🚚 Shipping Fee Report Spreadsheet"
      helper-text="Spreadsheet untuk menyimpan laporan perhitungan shipping fee dari Shopee"
      v-model="links.shipping"
      :validating="validating.shipping"
      :validated="validated.shipping"
      :error="errors.shipping"
      :metadata="metadata.shipping"
      :locked="locked.shipping"
      @check="checkLink('shipping')"
      @lock="lockLink('shipping')"
      @unlock="unlock('shipping')"
    />

    <SpreadsheetLinkField
      type="order"
      label="🛒 Order Export Spreadsheet"
      helper-text="Spreadsheet untuk export data pesanan"
      v-model="links.order"
      :validating="validating.order"
      :validated="validated.order"
      :error="errors.order"
      :metadata="metadata.order"
      :locked="locked.order"
      @check="checkLink('order')"
      @lock="lockLink('order')"
      @unlock="unlock('order')"
    />

    <div class="section actions-section">
      <button @click="saveLinks" :disabled="!hasChanges" class="btn-save">
        <span>💾 Save</span>
      </button>
      <button @click="clearSheetCache" class="btn-secondary">
        <span>🗑️ Clear Cache</span>
      </button>
    </div>

    <div v-if="statusMsg" :class="['status-msg', statusType]">
      {{ statusMsg }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import SpreadsheetLinkField from "./SpreadsheetLinkField.vue";
import { useGoogleSheetsLinks } from "../composables/useGoogleSheetsLinks";
import { useGoogleSheetsSettings } from "../composables/useGoogleSheetsSettings";
import { useCacheManagement } from "../composables/useCacheManagement";

const { clearType } = useCacheManagement();

const {
  links,
  validating,
  validated,
  errors,
  metadata,
  locked,
  hasChanges,
  checkLink,
  saveLinks,
  unlock,
} = useGoogleSheetsLinks();

const lockLink = (type: "inventory" | "wallet" | "shipping" | "order") => {
  locked.value[type] = true;
};

const clearSheetCache = async () => {
  const result = await clearType("googleSheets");
  if (result.success) {
    alert("✅ Google Sheets cache cleared. Refresh to load fresh data.");
    window.location.reload();
  }
};

const {
  inventorySheet,
  statusMsg,
  statusType,
  loadSettings,
  loadDetailedSettings,
  processDetailedSettings,
  loadFromCacheAsFallback,
  updateInventorySheet,
  metadata: settingsMetadata,
} = useGoogleSheetsSettings();

const loadSavedLinks = async () => {
  try {
    const linksData = await loadSettings();
    if (
      linksData &&
      (linksData.inventory ||
        linksData.wallet ||
        linksData.shipping ||
        linksData.order)
    ) {
      links.value = {
        inventory: linksData.inventory || "",
        wallet: linksData.wallet || "",
        shipping: linksData.shipping || "",
        order: linksData.order || "",
      };
      // Set validated and locked based on saved links
      Object.keys(links.value).forEach((key) => {
        if (links.value[key as keyof typeof links.value]) {
          validated.value[key as keyof typeof validated.value] = true;
          locked.value[key as keyof typeof locked.value] = true;
        }
      });
    }
  } catch (error) {
    console.error("Failed to load links:", error);
  }

  try {
    const detailedData = await loadDetailedSettings();
    processDetailedSettings(detailedData);

    // Sync metadata from settings to links composable
    Object.assign(metadata.value, settingsMetadata.value);
  } catch (error) {
    console.warn("Failed to load detailed settings:", error);
    // Try fallback to cache if API fails
    loadFromCacheAsFallback();

    // Sync cached metadata to links composable
    Object.assign(metadata.value, settingsMetadata.value);
  }
};

onMounted(() => {
  loadSavedLinks();
});
</script>

<style scoped>
.google-sheets-settings {
  max-width: 700px;
}

.settings-header {
  margin-bottom: 30px;
}

.settings-header h2 {
  font-size: 24px;
  margin: 0 0 8px;
  color: #333;
}

.settings-header p {
  color: #666;
  margin: 0;
  font-size: 14px;
}

.section {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.actions-section {
  display: flex;
  gap: 12px;
  justify-content: flex-start;
}

.btn-save,
.btn-secondary {
  padding: 10px 20px;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-save {
  background: #4285f4;
}

.btn-save:hover:not(:disabled) {
  background: #3367d6;
  box-shadow: 0 2px 8px rgba(66, 133, 244, 0.3);
}

.btn-save:disabled {
  background: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.btn-secondary {
  background: #666;
}

.btn-secondary:hover {
  background: #555;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.status-msg {
  padding: 12px 16px;
  border-radius: 6px;
  margin-top: 16px;
  font-size: 14px;
  font-weight: 500;
}

.status-msg.success {
  background: #e8f5e9;
  color: #2e7d32;
  border: 1px solid #c8e6c9;
}

.status-msg.error {
  background: #ffebee;
  color: #c62828;
  border: 1px solid #ffcdd2;
}
</style>
