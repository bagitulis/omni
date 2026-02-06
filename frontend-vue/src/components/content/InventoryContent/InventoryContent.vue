<template>
  <div class="inventory-content">
    <!-- Header -->
    <div class="inventory-header">
      <div class="header-top"><h1>📦 Inventory Management</h1></div>
      <InventoryHeader
        :search-query="searchQuery"
        :syncing="syncing"
        :updating-stock="updatingStock"
        :updating-price="updatingPrice"
        :has-active-filters="hasActiveFiltersComputed"
        @search-changed="searchQuery = $event"
        @search-submit="filterMethods?.handleSearch"
        @clear-filters="filterMethods?.clearAllFilters"
        @check-platform-status="batchCheckMethods?.handleBatchCheckPlatform"
        @sync-from="syncMethods?.handleSyncFromSheets"
        @sync-to="syncMethods?.handleSyncToSheets"
        @update-stock="stockUpdateMethods?.handleUpdateStock"
        @update-price="priceUpdateMethods?.handleUpdatePrice"
        @delete-wholesale="
          wholesaleMethods.handleDeleteWholesale(inventoryList)
        "
        @update-wholesale="
          wholesaleMethods.handleUpdateWholesale(filteredInventoryList)
        "
        @open-mpq-modal="
          wholesaleMethods.handleOpenMpqModal(filteredInventoryList)
        "
        @add-item="showAddForm = true"
      />
    </div>

    <!-- Filter & Stats -->
    <InventoryFilterPanel
      :schema-columns="schemaColumns"
      :column-visibility="columnVisibility"
      :column-filters="columnFilters"
      :full-inventory-list="inventoryList"
      @visibility-changed="handleVisibilityChange"
      @filter-changed="handleFilterChange"
      @columns-saved="handleColumnsSaved"
      @sort="handleSort"
    />
    <InventoryStats
      :stats="stats"
      :sync-status="syncStatus?.status || 'unknown'"
    />
    <InventoryAlert
      :alert="syncAlert"
      :show="!!syncAlert"
      @close-alert="syncAlert = null"
    />

    <!-- Form & Table -->
    <InventoryForm
      :schema-columns="schemaColumns"
      :form-data="formData"
      :show="showAddForm"
      :editing-item="editingItem"
      @save-item="formMethods.saveItem"
      @close-form="formMethods.closeAddForm"
    />
    <InventoryTable
      :filtered-list="filteredInventoryList"
      :full-inventory-list="inventoryList"
      :schema-columns="schemaColumns"
      :column-visibility="columnVisibility"
      :loading="loading"
      :batch-results="batchCheckResults"
      :lock-stock-map="lockStock.map"
      @cell-edit-save="cellEditMethods?.handleCellEditSave"
      @edit-item="itemEditMethods?.handleEditItem"
      @delete-item="itemEditMethods?.handleDeleteItem"
      @filter-changed="handleFilterChange"
      @sort="handleSort"
      @open-marketplace-settings="showMarketplaceSettings = true"
      @clone-product="handleCloneProduct"
    />

    <!-- Pagination -->
    <InventoryPagination
      v-if="inventoryList.length > 0"
      :current-offset="currentOffset"
      :page-size="pageSize"
      :total-records="totalRecords"
      @previous="paginationMethods.previousPage()"
      @next="paginationMethods.nextPage()"
      @page-size-changed="handlePageSizeChange"
    />

    <!-- Modals -->
    <InventoryModals
      :sync-history="syncHistory"
      :show-sync-history="showSyncHistory"
      :show-marketplace-settings="showMarketplaceSettings"
      :schema-columns="schemaColumns"
      :show-wholesale-modal="showWholesaleModal"
      :wholesale-skus="wholesaleSkus"
      :show-wholesale-update-modal="showWholesaleUpdateModal"
      :wholesale-update-items="wholesaleUpdateItems"
      :show-wholesale-mpq-modal="showWholesaleMpqModal"
      :wholesale-mpq-items="wholesaleMpqItems"
      @close-sync-history="showSyncHistory = false"
      @close-marketplace-settings="showMarketplaceSettings = false"
      @marketplace-saved="$forceUpdate()"
      @close-wholesale-modal="showWholesaleModal = false"
      @wholesale-completed="wholesaleMethods.handleWholesaleCompleted"
      @close-wholesale-update="showWholesaleUpdateModal = false"
      @wholesale-update-completed="
        wholesaleMethods.handleWholesaleUpdateCompleted
      "
      @close-mpq-modal="showWholesaleMpqModal = false"
      @mpq-completed="wholesaleMethods.handleWholesaleMpqCompleted"
    />

    <!-- Clone Product Modal -->
    <CloneProductModal
      :show="showCloneModal"
      :sku="cloneSourceSku"
      :source-platform="cloneSourcePlatform"
      :loading="cloneLoading"
      :error="cloneError"
      :product-data="cloneProductData"
      :available-targets="cloneAvailableTargets"
      :selected-targets="cloneSelectedTargets"
      @close="closeCloneModal"
      @toggle-target="toggleCloneTarget"
      @start-clone="startCloneToTarget"
    />
  </div>
</template>

<script>
import { defineComponent } from "vue";
import { useInventoryInit } from "./composables/useInventoryInit";
import { useInventoryConfig } from "./composables/useInventoryConfig";
import { useInventoryAlerts } from "./composables/useInventoryAlerts";
import { saveSortState } from "./composables/useSortPersistence";
import {
  loadPageSize,
  savePageSize,
} from "./composables/usePageSizePersistence";

import InventoryHeader from "./InventoryHeader.vue";
import InventoryFilterPanel from "./InventoryFilterPanel.vue";
import InventoryTable from "./InventoryTable.vue";
import InventoryForm from "./InventoryForm.vue";
import InventoryStats from "./InventoryStats.vue";
import InventoryAlert from "./InventoryAlert.vue";
import InventoryPagination from "./InventoryPagination.vue";
import InventoryModals from "./InventoryModals.vue";
import CloneProductModal from "./CloneProductModal.vue";
import { useProductClone } from "./composables/useProductClone";

export default defineComponent({
  name: "InventoryContent",
  components: {
    InventoryHeader,
    InventoryFilterPanel,
    InventoryTable,
    InventoryForm,
    InventoryStats,
    InventoryAlert,
    InventoryPagination,
    InventoryModals,
    CloneProductModal,
  },
  props: { spreadsheetId: String, sheetName: String },

  data() {
    return {
      inventoryList: [],
      schemaColumns: [],
      syncHistory: [],
      stats: {
        total_records: 0,
        total_columns: 0,
        columns: [],
        last_sync: null,
        db_size_kb: 0,
      },
      loading: false,
      syncing: false,
      showAddForm: false,
      showSyncHistory: false,
      showMarketplaceSettings: false,
      isInitialized: false,
      searchQuery: "",
      columnVisibility: {},
      columnFilters: {},
      currentOffset: 0,
      pageSize: loadPageSize(),
      totalRecords: 0,
      sortColumn: null,
      sortDirection: null,
      lockStock: { loading: false, error: null, map: {} },
      formData: {},
      editingItem: null,
      keyColumn: "",
      syncAlert: null,
      syncStatus: null,
      batchCheckResults: [],
      updatingStock: false,
      updatingPrice: false,
      showWholesaleModal: false,
      wholesaleSkus: [],
      showWholesaleUpdateModal: false,
      wholesaleUpdateItems: [],
      showWholesaleMpqModal: false,
      wholesaleMpqItems: [],
      // Clone product state
      showCloneModal: false,
      cloneSourceSku: "",
      cloneSourcePlatform: "",
      cloneLoading: false,
      cloneError: null,
      cloneProductData: null,
      cloneAvailableTargets: [],
      cloneSelectedTargets: [],
      configData: {
        spreadsheet_id: "",
        sheet_name: "",
        selected_columns: [],
        key_column: "",
        header_row: 1,
        data_start_row: 2,
        auto_sync: false,
        sync_interval_seconds: 300,
        last_sync_timestamp: null,
      },
    };
  },

  computed: {
    filteredInventoryList() {
      if (!this.inventoryList?.length) return [];
      const list = this.filterMethods?.getFilteredInventoryList?.() || [
        ...this.inventoryList,
      ];
      const sorted =
        this.sortColumn && this.sortDirection ? this.sortList(list) : list;
      return this.paginationMethods.paginateList(sorted);
    },
    hasActiveFiltersComputed() {
      const hasFilters = Object.values(this.columnFilters || {}).some(
        (v) => v && String(v).trim(),
      );
      return !!(hasFilters || this.searchQuery?.trim());
    },
  },

  watch: {
    schemaColumns(cols) {
      cols?.forEach((col) => {
        if (!(col.column_name in this.columnVisibility))
          this.columnVisibility[col.column_name] = true;
        if (!(col.column_name in this.columnFilters))
          this.columnFilters[col.column_name] = "";
      });
    },
    searchQuery(q) {
      if (this._timer) clearTimeout(this._timer);
      this._timer = setTimeout(() => {
        useInventoryConfig().setSearchQuery(q);
        this.currentOffset = 0;
      }, 300);
    },
  },

  created() {
    const initHelper = useInventoryInit(this.$data);
    Object.assign(this, initHelper.initializeMethods());
    initHelper.loadPersistedSort();
  },

  async mounted() {
    this.loading = true;
    try {
      const initHelper = useInventoryInit(this.$data);
      await initHelper.initializeInventory(this);
      initHelper.showSetupWarningIfNeeded();
    } catch (e) {
      useInventoryAlerts().showAlertError("⚠️ Initialization Error", String(e));
    } finally {
      this.loading = false;
    }
  },

  beforeUnmount() {
    useInventoryInit(this.$data).cleanup(this._timer);
  },

  methods: {
    sortList(list) {
      const parse = (item) => {
        // Handle both string (Node.js) and object (Go) formats
        if (item?.data !== undefined && item?.data !== null) {
          if (typeof item.data === "string") {
            try {
              return JSON.parse(item.data);
            } catch {
              return item;
            }
          } else if (typeof item.data === "object") {
            return item.data;
          }
        }
        return item;
      };
      return [...list].sort((a, b) => {
        const cmp = String(parse(a)[this.sortColumn] || "")
          .toLowerCase()
          .localeCompare(String(parse(b)[this.sortColumn] || "").toLowerCase());
        return this.sortDirection === "asc" ? cmp : -cmp;
      });
    },
    handleSort({ column, direction }) {
      this.sortColumn = direction ? column : null;
      this.sortDirection = direction;
      saveSortState({ column: this.sortColumn, direction: this.sortDirection });
      this.currentOffset = 0;
    },
    handleVisibilityChange({ column, visible }) {
      this.columnVisibility[column] = visible;
      // Note: Don't call loadStats here as it's not needed for visibility change
      // and can cause unnecessary re-renders
    },
    handleFilterChange({ column, values }) {
      if (values?.length) this.columnFilters[column] = JSON.stringify(values);
      else delete this.columnFilters[column];
      this.currentOffset = 0;
      this.columnFilters = { ...this.columnFilters };
    },
    handleColumnsSaved() {
      this.dataMethods?.loadStats();
      this.dataMethods?.loadInventoryData();
    },
    handlePageSizeChange(s) {
      this.pageSize = s;
      savePageSize(s);
      this.currentOffset = 0;
    },
    // Clone product methods
    async handleCloneProduct({ sku, platform }) {
      this.showCloneModal = true;
      this.cloneSourceSku = sku;
      this.cloneSourcePlatform = platform;
      this.cloneLoading = true;
      this.cloneError = null;
      this.cloneProductData = null;
      this.cloneAvailableTargets = [];
      this.cloneSelectedTargets = [];

      try {
        const productClone = useProductClone();
        await productClone.fetchCloneData(sku, platform);
        this.cloneProductData = productClone.productData.value;
        this.cloneAvailableTargets = productClone.availableTargets.value;
        this.cloneError = productClone.cloneError.value;
      } catch (err) {
        this.cloneError = String(err);
      } finally {
        this.cloneLoading = false;
      }
    },
    closeCloneModal() {
      this.showCloneModal = false;
      this.cloneSourceSku = "";
      this.cloneSourcePlatform = "";
      this.cloneProductData = null;
      this.cloneAvailableTargets = [];
      this.cloneSelectedTargets = [];
      this.cloneError = null;
    },
    toggleCloneTarget(platform) {
      const idx = this.cloneSelectedTargets.indexOf(platform);
      if (idx >= 0) {
        this.cloneSelectedTargets.splice(idx, 1);
      } else {
        this.cloneSelectedTargets.push(platform);
      }
    },
    startCloneToTarget(targetPlatform) {
      // Open AddProductModal with pre-filled data (to be implemented via router or modal)
      console.log(
        `Clone ${this.cloneSourceSku} from ${this.cloneSourcePlatform} to ${targetPlatform}`,
        this.cloneProductData,
      );
      // For now, navigate to product creation with query params
      this.$router.push({
        path: "/products/add",
        query: {
          platform: targetPlatform,
          cloneSku: this.cloneSourceSku,
          cloneFrom: this.cloneSourcePlatform,
        },
      });
      this.closeCloneModal();
    },
  },
});
</script>

<style scoped>
@import "./InventoryContent.styles.css";
</style>
