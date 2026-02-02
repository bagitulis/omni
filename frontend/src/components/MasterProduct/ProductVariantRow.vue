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
    <td
      class="price-cell editable-cell text-left"
      :class="{
        editing: isEditing('price'),
        saving: isSaving('price'),
      }"
      @click="$emit('start-edit', sku.id, 'price', sku.price)"
    >
      <template v-if="isEditing('price')">
        <input
          type="number"
          :value="editedValue"
          @input="
            $emit(
              'update:editedValue',
              ($event.target as HTMLInputElement).valueAsNumber,
            )
          "
          class="inline-edit-input"
          @keydown.enter="$emit('save-edit', sku)"
          @keydown.escape="$emit('cancel-edit')"
          @blur="$emit('save-edit', sku)"
          min="0"
          step="1000"
          autoFocus
        />
      </template>
      <template v-else>
        <span class="cell-value">{{ formatPrice(sku.price) }}</span>
        <span class="edit-hint" v-if="!isSaving('price')">✏️</span>
        <span class="saving-indicator" v-if="isSaving('price')">💾</span>
      </template>
    </td>

    <!-- Stock Cell -->
    <td
      class="stock-cell editable-cell text-left"
      :class="{
        editing: isEditing('stock'),
        saving: isSaving('stock'),
      }"
      @click="$emit('start-edit', sku.id, 'stock', sku.stock)"
    >
      <template v-if="isEditing('stock')">
        <input
          type="number"
          :value="editedValue"
          @input="
            $emit(
              'update:editedValue',
              ($event.target as HTMLInputElement).valueAsNumber,
            )
          "
          class="inline-edit-input"
          @keydown.enter="$emit('save-edit', sku)"
          @keydown.escape="$emit('cancel-edit')"
          @blur="$emit('save-edit', sku)"
          min="0"
          step="1"
          autoFocus
        />
      </template>
      <template v-else>
        <span class="cell-value">{{ sku.stock }}</span>
        <span class="edit-hint" v-if="!isSaving('stock')">✏️</span>
        <span class="saving-indicator" v-if="isSaving('stock')">💾</span>
      </template>
    </td>

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

.text-left {
  text-align: left;
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

.editable-cell {
  cursor: pointer;
  position: relative;
  transition: all 0.2s ease;
  min-width: 100px;
}

.editable-cell:hover:not(.editing):not(.saving) {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%) !important;
  border-radius: 0.25rem;
}

.editable-cell .cell-value {
  display: inline-block;
}

.editable-cell .edit-hint {
  opacity: 0;
  margin-left: 0.5rem;
  font-size: 0.75rem;
  transition: opacity 0.2s ease;
}

.editable-cell:hover .edit-hint {
  opacity: 0.7;
}

.editable-cell.editing {
  padding: 0.5rem !important;
  background: #fffbeb !important;
}

.editable-cell.saving {
  opacity: 0.7;
  pointer-events: none;
}

.saving-indicator {
  margin-left: 0.5rem;
  animation: pulse 1s ease-in-out infinite;
}

.inline-edit-input {
  width: 100%;
  max-width: 120px;
  padding: 0.375rem 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  border: 2px solid #3b82f6;
  border-radius: 0.375rem;
  background: white;
  color: inherit;
  outline: none;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  transition: all 0.2s ease;
}

.inline-edit-input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.15);
}

.inline-edit-input::-webkit-outer-spin-button,
.inline-edit-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.inline-edit-input[type="number"] {
  -moz-appearance: textfield;
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

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
