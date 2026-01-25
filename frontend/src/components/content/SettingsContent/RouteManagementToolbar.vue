<template>
  <div class="toolbar-wrapper">
    <div class="filters-section">
      <label for="route-search" class="sr-only">Search routes</label>
      <input
        id="route-search"
        :value="searchQuery"
        type="text"
        placeholder="🔍 Search routes..."
        class="search-input"
        @input="$emit('update-search', ($event.target as any).value)"
      />
      <label for="route-category-filter" class="sr-only"
        >Filter by category</label
      >
      <select
        id="route-category-filter"
        :value="selectedCategory"
        class="filter-select"
        @change="$emit('update-category', ($event.target as any).value)"
      >
        <option value="">All Categories</option>
        <option v-for="cat in categories" :key="cat" :value="cat">
          {{ cat || "Uncategorized" }}
        </option>
      </select>
      <button @click="$emit('toggle-all-enabled')" class="btn btn-secondary">
        {{ allEnabled ? "🔓 Disable All" : "🔒 Enable All" }}
      </button>
    </div>

    <!-- Bulk Actions -->
    <div v-if="selectedCount > 0" class="bulk-actions">
      <span class="selection-info">{{ selectedCount }} routes selected</span>
      <div class="action-buttons">
        <button @click="$emit('apply-preset', 'balanced')" class="btn btn-info">
          ⚡ Apply Preset
        </button>
        <button @click="$emit('bulk-delete')" class="btn btn-danger">
          🗑️ Delete Selected
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  searchQuery: string;
  selectedCategory: string;
  categories: string[];
  allEnabled: boolean;
  selectedCount: number;
}>();

import type { PresetType } from "./types/routeManagement";

defineEmits<{
  "update-search": [value: string];
  "update-category": [value: string];
  "toggle-all-enabled": [];
  "bulk-delete": [];
  "apply-preset": [preset: PresetType];
}>();
</script>

<style scoped lang="css">
.filters-section {
  display: flex;
  gap: 12px;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
  flex-wrap: wrap;
}

.search-input,
.filter-select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
}

.search-input {
  flex: 1;
  min-width: 200px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary {
  background: #666;
  color: white;
}
.btn-secondary:hover {
  background: #555;
}

.btn-info {
  background: #00bcd4;
  color: white;
}
.btn-info:hover {
  background: #0097a7;
}

.btn-danger {
  background: #f44336;
  color: white;
}
.btn-danger:hover {
  background: #d32f2f;
}

.bulk-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  background: #e3f2fd;
  border-radius: 8px;
  border-left: 4px solid #2196f3;
}

.selection-info {
  font-weight: 600;
  color: #1976d2;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
