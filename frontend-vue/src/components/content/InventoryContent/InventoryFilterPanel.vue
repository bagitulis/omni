<template>
  <div class="inventory-filter-panel" v-if="schemaColumns.length > 0">
    <!-- Column Manager Section (Visibility + Lock + Filter) -->
    <InventoryColumnManager
      :schema-columns="schemaColumns"
      :full-inventory-list="fullInventoryList"
      @visibility-changed="handleVisibilityChange"
      @columns-saved="$emit('columns-saved')"
      @filter-changed="handleFilterChanged"
      @sort="$emit('sort', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import InventoryColumnManager from "./InventoryColumnManager.vue";

interface Column {
  column_name: string;
  column_type: string;
  is_key: boolean;
}

defineProps<{
  schemaColumns: Column[];
  columnVisibility: Record<string, boolean>;
  columnFilters: Record<string, string>;
  fullInventoryList: Record<string, any>[];
}>();

const emits = defineEmits<{
  "visibility-changed": [data: { column: string; visible: boolean }];
  "filter-changed": [data: { column: string; values: string[] }];
  "columns-saved": [];
  sort: [data: { column: string; direction: "asc" | "desc" | null }];
}>();

const handleVisibilityChange = (data: { column: string; visible: boolean }) => {
  emits("visibility-changed", data);
};

const handleFilterChanged = (data: { column: string; values: string[] }) => {
  emits("filter-changed", data);
};
</script>

<style scoped>
.inventory-filter-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  box-sizing: border-box;
}
</style>
