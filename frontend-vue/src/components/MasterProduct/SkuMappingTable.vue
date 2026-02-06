<template>
  <div class="mapping-table-wrapper">
    <table class="mapping-table">
      <thead>
        <tr>
          <th>Seller SKU</th>
          <th>Variant</th>
          <th>🟠 Shopee</th>
          <th>⬛ TikTok</th>
          <th>🔵 Lazada</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="sku in skus" :key="sku.master_sku_id">
          <td class="sku-code">{{ sku.seller_sku }}</td>
          <td>{{ sku.variant_name || "-" }}</td>

          <td class="platform-cell">
            <span v-if="getPlatformLink(sku, 'shopee')" class="linked-badge"
              >✓ Linked</span
            >
            <button
              v-else
              @click="$emit('link', { sku, platform: 'shopee' })"
              class="btn-link"
            >
              Link
            </button>
          </td>
          <td class="platform-cell">
            <span v-if="getPlatformLink(sku, 'tiktok')" class="linked-badge"
              >✓ Linked</span
            >
            <button
              v-else
              @click="$emit('link', { sku, platform: 'tiktok' })"
              class="btn-link"
            >
              Link
            </button>
          </td>
          <td class="platform-cell">
            <span v-if="getPlatformLink(sku, 'lazada')" class="linked-badge"
              >✓ Linked</span
            >
            <button
              v-else
              @click="$emit('link', { sku, platform: 'lazada' })"
              class="btn-link"
            >
              Link
            </button>
          </td>

          <td class="actions-cell">
            <button
              @click="$emit('auto-map', sku)"
              title="Auto-map"
              class="btn-action"
            >
              🔍
            </button>
            <button
              v-if="hasAnyLink(sku)"
              @click="$emit('unlink', sku)"
              title="Unlink"
              class="btn-action"
            >
              🔗
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { SkuMappingInfo } from "@/services/masterProductService";

interface Props {
  skus: SkuMappingInfo[];
}
defineProps<Props>();

const emit = defineEmits<{
  (e: "link", payload: { sku: SkuMappingInfo; platform: string }): void;
  (e: "auto-map", sku: SkuMappingInfo): void;
  (e: "unlink", sku: SkuMappingInfo): void;
}>();

const getPlatformLink = (sku: SkuMappingInfo, platform: string) => {
  return sku.platform_links?.find((link) => link.platform === platform);
};

const hasAnyLink = (sku: SkuMappingInfo): boolean => {
  return (sku.platform_links?.length || 0) > 0;
};
</script>

<style scoped>
.mapping-table-wrapper {
  overflow-x: auto;
}

.mapping-table {
  width: 100%;
  border-collapse: collapse;
}

.mapping-table th,
.mapping-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.mapping-table th {
  background: #f9fafb;
  font-weight: 600;
  font-size: 0.875rem;
}

.sku-code {
  font-family: monospace;
  font-weight: 600;
}

.platform-cell {
  text-align: center;
}

.linked-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: #dcfce7;
  color: #166534;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.btn-link {
  padding: 0.25rem 0.75rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  cursor: pointer;
}

.btn-link:hover {
  background: #2563eb;
}

.actions-cell {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  width: 32px;
  height: 32px;
  border: 1px solid #e5e7eb;
  background: white;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 1rem;
}

.btn-action:hover {
  background: #f3f4f6;
}
</style>
