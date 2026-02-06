<template>
  <tr class="variant-row" :class="{ 'is-selected': selected }">
    <td class="checkbox-col">
      <input
        type="checkbox"
        :checked="selected"
        @change="$emit('toggle-row', sku.id)"
      />
    </td>
    <td class="sku-cell">{{ sku.seller_sku }}</td>
    <td class="variant-cell">{{ sku.variant_name || "-" }}</td>

    <!-- Price Cell -->
    <EditableCell
      class="price-cell"
      :value="sku.price"
      :editing-value="editedValue"
      :is-editing="isEditing('price')"
      :is-saving="isSaving('price')"
      :format-fn="formatPrice"
      :min="0"
      :step="1000"
      @start-edit="$emit('start-edit', sku.id, 'price', sku.price)"
      @update:editing-value="$emit('update:editedValue', $event)"
      @save-edit="$emit('save-edit', sku)"
      @cancel-edit="$emit('cancel-edit')"
    />

    <!-- Stock Cell -->
    <EditableCell
      class="stock-cell"
      :value="sku.stock"
      :editing-value="editedValue"
      :is-editing="isEditing('stock')"
      :is-saving="isSaving('stock')"
      :min="0"
      :step="1"
      @start-edit="$emit('start-edit', sku.id, 'stock', sku.stock)"
      @update:editing-value="$emit('update:editedValue', $event)"
      @save-edit="$emit('save-edit', sku)"
      @cancel-edit="$emit('cancel-edit')"
    />

    <!-- Platform Cell -->
    <td class="platform-cell">
      <div class="platform-icons">
        <ProductPlatformAction
          v-for="platform in ['shopee', 'tiktok', 'lazada']"
          :key="platform"
          :platform="platform"
          :linked="isPlatformLinked(platform)"
          :hasUpdate="needsSync(platform)"
          :loading="isPlatformLoading(platform)"
          :syncState="getSyncState(platform)"
          :title="getPlatformTitle(platform)"
          :emoji="getPlatformEmoji(platform)"
          @sync="$emit('sync', sku.id, platform)"
          @click="$emit('platform-click', sku, platform)"
        />
      </div>
    </td>
  </tr>
</template>

<script setup lang="ts">
import { type ProductSku, type PlatformLink } from "./ProductList.types";
import ProductPlatformAction from "./ProductPlatformAction.vue";
import EditableCell from "./EditableCell.vue";

const props = defineProps<{
  sku: ProductSku;
  selected: boolean;
  editingCell: { skuId: number; field: "price" | "stock" } | null;
  savingCell: { skuId: number; field: "price" | "stock" } | null;
  editedValue: number;
  skusNeedingSync: Map<number, Set<string>>;
  syncStatus: Map<string, "syncing" | "success" | "error">;
  syncErrorMessages: Map<string, string>;
  platformLoading: Map<string, boolean>;
}>();

defineEmits<{
  "toggle-row": [skuId: number];
  "start-edit": [skuId: number, field: "price" | "stock", value: number];
  "save-edit": [sku: ProductSku];
  "cancel-edit": [];
  "update:editedValue": [value: number];
  sync: [skuId: number, platform: string];
  "platform-click": [sku: ProductSku, platform: string];
}>();

// Helpers
const isEditing = (field: "price" | "stock"): boolean => {
  return (
    props.editingCell?.skuId === props.sku.id &&
    props.editingCell?.field === field
  );
};

const isSaving = (field: "price" | "stock"): boolean => {
  return (
    props.savingCell?.skuId === props.sku.id &&
    props.savingCell?.field === field
  );
};

const formatPrice = (price: number): string => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(price);
};

const isPlatformLinked = (platform: string): boolean => {
  return (
    props.sku.platform_links?.some(
      (link: PlatformLink) => link.platform === platform,
    ) ?? false
  );
};

const needsSync = (platform: string): boolean => {
  return props.skusNeedingSync.get(props.sku.id)?.has(platform) ?? false;
};

const getSyncState = (platform: string) => {
  return props.syncStatus.get(`${props.sku.id}-${platform}`);
};

const isPlatformLoading = (platform: string): boolean => {
  return props.platformLoading.get(`${props.sku.id}-${platform}`) ?? false;
};

const getPlatformTitle = (platform: string): string => {
  const syncKey = `${props.sku.id}-${platform}`;
  const errorMsg = props.syncErrorMessages.get(syncKey);
  if (errorMsg) return errorMsg;

  const linked = isPlatformLinked(platform);
  const platformName = platform.charAt(0).toUpperCase() + platform.slice(1);
  return linked
    ? `Terhubung ke ${platformName}`
    : `Belum terhubung ke ${platformName}`;
};

const getPlatformEmoji = (platform: string): string => {
  switch (platform) {
    case "shopee":
      return "🟠";
    case "tiktok":
      return "⬛";
    case "lazada":
      return "🔵";
    default:
      return "⚪";
  }
};
</script>

<style scoped>
.variant-row td {
  padding: 1rem;
  color: var(--text-primary, #1a1a1a);
  border-bottom: 1px solid #f3f4f6;
  transition: background-color 0.15s ease;
}

.variant-row:hover td {
  background-color: #f8f9fa;
}

.variant-row:last-child td {
  border-bottom: none;
}

.variant-row.is-selected {
  background: #eff6ff !important;
}

.variant-row.is-selected td {
  background: #eff6ff !important;
}

.sku-cell {
  font-family: "SF Mono", "Monaco", "Cascadia Code", "Consolas", monospace;
  font-size: 0.8125rem;
  color: #2563eb;
  font-weight: 500;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  padding: 0.5rem 0.75rem !important;
  border-radius: 0.25rem;
  display: inline-block;
  min-width: 100px;
}

.variant-cell {
  font-weight: 500;
  color: #1f2937;
}

.price-cell {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: #059669;
}

.stock-cell {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: #dc2626;
}

.platform-cell {
  width: 140px;
}

.platform-icons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.checkbox-col {
  width: 40px;
  text-align: center;
}

.checkbox-col input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #3b82f6;
}
</style>
