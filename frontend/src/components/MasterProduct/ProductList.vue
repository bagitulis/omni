<template>
  <div class="product-list">
    <!-- Loading State -->
    <div v-if="loading" class="loading-spinner">
      <div class="spinner"></div>
      <p>Memuat produk...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="products.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <h3>Belum ada produk</h3>
      <p>Mulai dengan mengimpor produk dari Shopee atau tambah produk baru</p>
      <div class="empty-actions">
        <router-link to="/master-products/import" class="btn btn-secondary">
          Import dari Shopee
        </router-link>
        <router-link to="/master-products/add" class="btn btn-primary">
          Tambah Produk
        </router-link>
      </div>
    </div>

    <!-- Product Cards -->
    <div v-else class="product-cards">
      <div v-for="product in products" :key="product.id" class="product-card">
        <!-- Card Header: Image + Info -->
        <div class="card-header">
          <!-- Product Thumbnail -->
          <div class="product-thumbnail">
            <img
              v-if="getFirstImage(product)"
              :src="getFirstImage(product)"
              :alt="product.title"
              class="thumbnail-image"
              loading="lazy"
            />
            <div v-else class="thumbnail-placeholder">
              <span>📷</span>
            </div>
          </div>

          <!-- Product Info -->
          <div class="product-info">
            <h3 class="product-title">{{ product.title }}</h3>
            <div class="product-meta">
              <span class="sku-count">
                {{ product.skus?.length || 0 }} SKU
              </span>
              <!-- Show price/stock for single-SKU products inline -->
              <template v-if="product.skus?.length === 1">
                <span class="product-price-inline">{{
                  formatPrice(product.skus[0].price)
                }}</span>
                <span class="product-stock-inline"
                  >Stok: {{ product.skus[0].stock }}</span
                >
              </template>
              <span
                class="product-status"
                :class="getStatusClass(product.status)"
              >
                {{ getStatusLabel(product.status) }}
              </span>
            </div>
            <div class="product-actions">
              <router-link
                :to="`/master-products/${product.id}`"
                class="btn-action btn-edit"
              >
                Edit
              </router-link>
              <button
                class="btn-action btn-delete"
                @click.stop="$emit('delete', product)"
              >
                Hapus
              </button>
            </div>
          </div>

          <!-- Expand Toggle - Only show for products with more than 1 SKU -->
          <button
            v-if="product.skus && product.skus.length > 1"
            class="expand-toggle"
            :class="{ expanded: expandedProducts.has(product.id) }"
            @click="toggleExpand(product.id)"
            :aria-label="
              expandedProducts.has(product.id)
                ? 'Collapse variants'
                : 'Expand variants'
            "
          >
            <span class="chevron">▼</span>
          </button>
        </div>

        <!-- Expandable Variant Table -->
        <transition name="expand">
          <div
            v-if="expandedProducts.has(product.id)"
            class="variant-table-wrapper"
          >
            <table class="variant-table">
              <thead>
                <tr>
                  <th>SKU</th>
                  <th>Varian</th>
                  <th>Harga</th>
                  <th>Stok</th>
                  <th>Platform</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="sku in product.skus"
                  :key="sku.id"
                  class="variant-row"
                >
                  <td class="sku-cell">{{ sku.seller_sku }}</td>
                  <td class="variant-cell">{{ sku.variant_name || "-" }}</td>
                  <td
                    class="price-cell editable-cell"
                    :class="{
                      editing: isEditing(sku.id, 'price'),
                      saving:
                        savingCell?.skuId === sku.id &&
                        savingCell?.field === 'price',
                    }"
                    @click="startEdit(sku.id, 'price', sku.price)"
                  >
                    <template v-if="isEditing(sku.id, 'price')">
                      <input
                        type="number"
                        v-model.number="editedValue"
                        class="inline-edit-input"
                        @keydown.enter="saveEdit(sku, product)"
                        @keydown.escape="cancelEdit"
                        @blur="saveEdit(sku, product)"
                        ref="editInputRef"
                        min="0"
                        step="1000"
                      />
                    </template>
                    <template v-else>
                      <span class="cell-value">{{
                        formatPrice(sku.price)
                      }}</span>
                      <span
                        class="edit-hint"
                        v-if="
                          savingCell?.skuId !== sku.id ||
                          savingCell?.field !== 'price'
                        "
                        >✏️</span
                      >
                      <span
                        class="saving-indicator"
                        v-if="
                          savingCell?.skuId === sku.id &&
                          savingCell?.field === 'price'
                        "
                        >💾</span
                      >
                    </template>
                  </td>
                  <td
                    class="stock-cell editable-cell"
                    :class="{
                      editing: isEditing(sku.id, 'stock'),
                      saving:
                        savingCell?.skuId === sku.id &&
                        savingCell?.field === 'stock',
                    }"
                    @click="startEdit(sku.id, 'stock', sku.stock)"
                  >
                    <template v-if="isEditing(sku.id, 'stock')">
                      <input
                        type="number"
                        v-model.number="editedValue"
                        class="inline-edit-input"
                        @keydown.enter="saveEdit(sku, product)"
                        @keydown.escape="cancelEdit"
                        @blur="saveEdit(sku, product)"
                        ref="editInputRef"
                        min="0"
                        step="1"
                      />
                    </template>
                    <template v-else>
                      <span class="cell-value">{{ sku.stock }}</span>
                      <span
                        class="edit-hint"
                        v-if="
                          savingCell?.skuId !== sku.id ||
                          savingCell?.field !== 'stock'
                        "
                        >✏️</span
                      >
                      <span
                        class="saving-indicator"
                        v-if="
                          savingCell?.skuId === sku.id &&
                          savingCell?.field === 'stock'
                        "
                        >💾</span
                      >
                    </template>
                  </td>
                  <td class="platform-cell">
                    <div class="platform-icons">
                      <div
                        v-for="platform in ['shopee', 'tiktok', 'lazada']"
                        :key="platform"
                        class="platform-item"
                      >
                        <!-- Sync Button (only when needed) -->
                        <button
                          v-if="
                            isPlatformLinked(sku, platform) &&
                            needsSync(sku.id, platform) &&
                            !getSyncState(sku.id, platform)
                          "
                          class="sync-btn"
                          @click.stop="
                            syncToPlatform(product.id, sku.id, platform)
                          "
                          title="Sinkronisasi perubahan"
                        >
                          🔄
                        </button>

                        <!-- Loading Spinner -->
                        <div
                          v-else-if="
                            getSyncState(sku.id, platform) === 'syncing'
                          "
                          class="sync-spinner"
                        ></div>

                        <!-- Success Indicator -->
                        <span
                          v-else-if="
                            getSyncState(sku.id, platform) === 'success'
                          "
                          class="sync-success"
                          >✅</span
                        >

                        <!-- Error Indicator -->
                        <span
                          v-else-if="getSyncState(sku.id, platform) === 'error'"
                          class="sync-error"
                          title="Gagal sinkronisasi"
                          @click="syncToPlatform(product.id, sku.id, platform)"
                          >⚠️</span
                        >

                        <!-- Platform Icon -->
                        <span
                          class="platform-icon"
                          :class="[
                            platform,
                            {
                              linked: isPlatformLinked(sku, platform),
                              'has-update': needsSync(sku.id, platform),
                              'is-loading': isPlatformLoading(sku.id, platform),
                            },
                          ]"
                          :title="getPlatformTitle(platform, sku)"
                          @click.stop="handlePlatformClick(sku, platform)"
                        >
                          {{
                            isPlatformLoading(sku.id, platform)
                              ? "⏳"
                              : getPlatformEmoji(platform)
                          }}
                        </span>
                      </div>
                    </div>
                  </td>
                </tr>
                <tr v-if="!product.skus?.length">
                  <td colspan="5" class="no-skus">Belum ada SKU</td>
                </tr>
              </tbody>
            </table>
          </div>
        </transition>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button
        class="btn-pagination"
        :disabled="currentPage <= 1"
        @click="$emit('page-change', currentPage - 1)"
      >
        ← Sebelumnya
      </button>
      <span class="page-info">
        Halaman {{ currentPage }} dari {{ totalPages }}
      </span>
      <button
        class="btn-pagination"
        :disabled="currentPage >= totalPages"
        @click="$emit('page-change', currentPage + 1)"
      >
        Selanjutnya →
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from "vue";
import masterProductService from "@/services/masterProductService";

// Types
interface PlatformLink {
  id: number;
  master_product_id: number;
  master_sku_id?: number;
  platform: string;
  platform_product_id?: string;
  platform_sku_id?: string;
  sync_status: string;
}

interface ProductSku {
  id: number;
  master_product_id: number;
  seller_sku: string;
  variant_name?: string;
  variant_data?: Record<string, any>;
  price: number;
  stock: number;
  platform_links?: PlatformLink[];
}

interface MasterProduct {
  id: number;
  tenant_id: string;
  title: string;
  description?: string;
  images?: string[];
  status: string;
  skus?: ProductSku[];
  created_at: string;
  updated_at: string;
}

// Props
const props = defineProps<{
  products: MasterProduct[];
  loading?: boolean;
  currentPage?: number;
  totalPages?: number;
}>();

// Emits
const emit = defineEmits<{
  delete: [product: MasterProduct];
  "page-change": [page: number];
  "sku-updated": [sku: ProductSku];
}>();

// State
const expandedProducts = ref<Set<number>>(new Set());

// Sync State
const skusNeedingSync = ref<Map<number, Set<string>>>(new Map());
const syncStatus = ref<Map<string, "syncing" | "success" | "error">>(new Map());
const syncErrorMessages = ref<Map<string, string>>(new Map());

// Platform Connect/Disconnect State
const platformLoading = ref<Map<string, boolean>>(new Map());

// Inline editing state
const editingCell = ref<{ skuId: number; field: "price" | "stock" } | null>(
  null,
);
const editedValue = ref<number>(0);
const originalValue = ref<number>(0);
const savingCell = ref<{ skuId: number; field: "price" | "stock" } | null>(
  null,
);
const editInputRef = ref<HTMLInputElement | null>(null);

// Inline editing methods
const isEditing = (skuId: number, field: "price" | "stock"): boolean => {
  return (
    editingCell.value?.skuId === skuId && editingCell.value?.field === field
  );
};

const startEdit = (skuId: number, field: "price" | "stock", value: number) => {
  // Don't start edit if already saving
  if (savingCell.value) return;

  editingCell.value = { skuId, field };
  editedValue.value = value;
  originalValue.value = value;

  // Focus the input after Vue updates the DOM
  nextTick(() => {
    const input = document.querySelector(
      ".inline-edit-input",
    ) as HTMLInputElement;
    if (input) {
      input.focus();
      input.select();
    }
  });
};

const cancelEdit = () => {
  editingCell.value = null;
  editedValue.value = 0;
  originalValue.value = 0;
};

const syncToPlatform = async (
  productId: number,
  skuId: number,
  platform: string,
) => {
  const syncKey = `${skuId}-${platform}`;
  syncStatus.value.set(syncKey, "syncing");
  syncErrorMessages.value.delete(syncKey);

  try {
    await masterProductService.sync(productId, platform);
    syncStatus.value.set(syncKey, "success");

    // Remove from needing sync list
    const platformSet = skusNeedingSync.value.get(skuId);
    if (platformSet) {
      platformSet.delete(platform);
      if (platformSet.size === 0) {
        skusNeedingSync.value.delete(skuId);
      }
    }

    // Clear success status after 3 seconds
    setTimeout(() => {
      syncStatus.value.delete(syncKey);
    }, 3000);
  } catch (error: any) {
    console.error(`Sync to ${platform} failed:`, error);
    syncStatus.value.set(syncKey, "error");
    syncErrorMessages.value.set(
      syncKey,
      error.response?.data?.error || "Gagal sinkronisasi",
    );
  }
};

const saveEdit = async (sku: ProductSku, product: MasterProduct) => {
  if (!editingCell.value) return;

  const { skuId, field } = editingCell.value;

  // Don't save if value hasn't changed
  if (editedValue.value === originalValue.value) {
    cancelEdit();
    return;
  }

  // Validate value
  if (editedValue.value < 0) {
    editedValue.value = 0;
  }

  // Set saving state
  savingCell.value = { skuId, field };
  editingCell.value = null;

  try {
    const updateInput =
      field === "price"
        ? { price: editedValue.value }
        : { stock: editedValue.value };

    const updatedSku = await masterProductService.updateSku(skuId, updateInput);

    // Update the local SKU data
    if (product.skus) {
      const skuIndex = product.skus.findIndex((s) => s.id === skuId);
      if (skuIndex !== -1) {
        product.skus[skuIndex] = { ...product.skus[skuIndex], ...updatedSku };
      }
    }

    // Emit event for parent
    emit("sku-updated", updatedSku);

    // Mark as needing sync for all linked platforms
    if (sku.platform_links?.length) {
      if (!skusNeedingSync.value.has(sku.id)) {
        skusNeedingSync.value.set(sku.id, new Set());
      }
      const platformSet = skusNeedingSync.value.get(sku.id)!;
      sku.platform_links.forEach((link) => {
        // Only if currently linked
        if (link.platform && link.platform_sku_id) {
          platformSet.add(link.platform);
        }
      });
    }
  } catch (error) {
    console.error("Failed to update SKU:", error);
    // Revert to original value on error
    if (field === "price") {
      sku.price = originalValue.value;
    } else {
      sku.stock = originalValue.value;
    }
  } finally {
    savingCell.value = null;
    editedValue.value = 0;
    originalValue.value = 0;
  }
};

// Methods
const toggleExpand = (productId: number) => {
  if (expandedProducts.value.has(productId)) {
    expandedProducts.value.delete(productId);
  } else {
    expandedProducts.value.add(productId);
  }
};

const getFirstImage = (product: MasterProduct): string | null => {
  if (!product.images || product.images.length === 0) return null;
  return product.images[0] || null;
};

const getStatusClass = (status: string): string => {
  switch (status) {
    case "active":
      return "status-active";
    case "draft":
      return "status-draft";
    case "archived":
      return "status-archived";
    default:
      return "status-draft";
  }
};

const getStatusLabel = (status: string): string => {
  switch (status) {
    case "active":
      return "Aktif";
    case "draft":
      return "Draft";
    case "archived":
      return "Diarsipkan";
    default:
      return status;
  }
};

const formatPrice = (price: number): string => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(price);
};

const isPlatformLinked = (sku: ProductSku, platform: string): boolean => {
  return (
    sku.platform_links?.some((link) => link.platform === platform) ?? false
  );
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

const getPlatformTitle = (platform: string, sku: ProductSku): string => {
  const syncKey = `${sku.id}-${platform}`;
  const errorMsg = syncErrorMessages.value.get(syncKey);
  if (errorMsg) return errorMsg;

  const linked = isPlatformLinked(sku, platform);
  const platformName = platform.charAt(0).toUpperCase() + platform.slice(1);
  return linked
    ? `Terhubung ke ${platformName}`
    : `Belum terhubung ke ${platformName}`;
};

const needsSync = (skuId: number, platform: string): boolean => {
  return skusNeedingSync.value.get(skuId)?.has(platform) ?? false;
};

const getSyncState = (skuId: number, platform: string) => {
  return syncStatus.value.get(`${skuId}-${platform}`);
};

// Platform Connect/Disconnect Methods
const isPlatformLoading = (skuId: number, platform: string): boolean => {
  return platformLoading.value.get(`${skuId}-${platform}`) ?? false;
};

const handlePlatformClick = async (sku: ProductSku, platform: string) => {
  const loadingKey = `${sku.id}-${platform}`;
  if (platformLoading.value.get(loadingKey)) return;

  const isLinked = isPlatformLinked(sku, platform);

  if (isLinked) {
    // Unlink Flow
    if (confirm(`Putuskan koneksi dari ${platform}?`)) {
      platformLoading.value.set(loadingKey, true);
      try {
        await masterProductService.unlinkSku(sku.id, platform);

        // Update local state: remove the link
        if (sku.platform_links) {
          sku.platform_links = sku.platform_links.filter(
            (link) => link.platform !== platform,
          );
        }
      } catch (error: any) {
        console.error(`Failed to unlink ${platform}:`, error);
        alert(
          `Gagal memutuskan koneksi: ${error.response?.data?.error || error.message}`,
        );
      } finally {
        platformLoading.value.delete(loadingKey);
      }
    }
  } else {
    // Link Flow
    const platformItemId = prompt(
      `Masukkan ID Produk untuk ${platform} (Item ID):`,
    );
    if (!platformItemId) return;

    // Optional: Ask for SKU ID if needed
    // const platformSkuId = prompt(`Masukkan ID SKU (Variant ID) jika ada (opsional):`);

    platformLoading.value.set(loadingKey, true);
    try {
      await masterProductService.manualLink(
        sku.id,
        platform,
        platformItemId,
        // platformSkuId || undefined
      );

      // Update local state: add the link
      if (!sku.platform_links) {
        sku.platform_links = [];
      }

      // Add a mock link so UI updates immediately (proper data comes on refresh)
      sku.platform_links.push({
        id: Date.now(), // Temporary ID
        master_product_id: sku.master_product_id,
        master_sku_id: sku.id,
        platform: platform,
        platform_product_id: platformItemId,
        sync_status: "linked",
      });

      // Mark for sync if needed
      if (!skusNeedingSync.value.has(sku.id)) {
        skusNeedingSync.value.set(sku.id, new Set());
      }
      skusNeedingSync.value.get(sku.id)!.add(platform);
    } catch (error: any) {
      console.error(`Failed to link ${platform}:`, error);
      alert(
        `Gagal menghubungkan: ${error.response?.data?.error || error.message}`,
      );
    } finally {
      platformLoading.value.delete(loadingKey);
    }
  }
};
</script>

<style scoped>
.product-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Loading State */
.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: var(--text-secondary, #6b7280);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border, #e5e7eb);
  border-top-color: var(--primary, #3b82f6);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  background: var(--surface, white);
  border-radius: 0.5rem;
  border: 1px dashed var(--border, #e5e7eb);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.empty-state h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin: 0 0 0.5rem 0;
}

.empty-state p {
  color: var(--text-secondary, #6b7280);
  margin: 0 0 1.5rem 0;
}

.empty-actions {
  display: flex;
  gap: 0.75rem;
}

/* Product Cards */
.product-cards {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.product-card {
  background: var(--surface, white);
  border-radius: 0.625rem;
  border: 1px solid var(--border, #e5e7eb);
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
}

.product-card:hover {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
  border-color: #d1d5db;
  transform: translateY(-1px);
}

/* Card Header */
.card-header {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  align-items: flex-start;
}

.product-thumbnail {
  width: 80px;
  height: 80px;
  border-radius: 0.375rem;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--surface-hover, #f3f4f6);
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumbnail-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  color: var(--text-secondary, #6b7280);
}

.product-info {
  flex: 1;
  min-width: 0;
}

.product-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin: 0 0 0.5rem 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-meta {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 0.75rem;
}

.sku-count {
  font-size: 0.875rem;
  color: #6b7280;
  background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  padding: 0.25rem 0.625rem;
  border-radius: 0.375rem;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

/* Inline price/stock for single-SKU products */
.product-price-inline {
  font-size: 0.875rem;
  font-weight: 600;
  color: #059669;
  padding: 0.25rem 0.5rem;
  background: #ecfdf5;
  border-radius: 0.25rem;
}

.product-stock-inline {
  font-size: 0.875rem;
  font-weight: 500;
  color: #6b7280;
  padding: 0.25rem 0.5rem;
  background: #f9fafb;
  border-radius: 0.25rem;
}

.product-status {
  font-size: 0.75rem;
  font-weight: 500;
  padding: 0.125rem 0.5rem;
  border-radius: 9999px;
}

.status-active {
  background: #dcfce7;
  color: #166534;
}

.status-draft {
  background: #fef3c7;
  color: #92400e;
}

.status-archived {
  background: #f3f4f6;
  color: #6b7280;
}

.product-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  border: none;
}

.btn-edit {
  background: var(--primary, #3b82f6);
  color: white;
}

.btn-edit:hover {
  background: var(--primary-hover, #2563eb);
}

.btn-delete {
  background: transparent;
  color: #dc2626;
  border: 1px solid #dc2626;
}

.btn-delete:hover {
  background: #fef2f2;
}

/* Expand Toggle */
.expand-toggle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 2px solid var(--border, #e5e7eb);
  background: linear-gradient(135deg, #ffffff 0%, #f9fafb 100%);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.expand-toggle:hover {
  background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  border-color: #d1d5db;
  transform: scale(1.05);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.expand-toggle:active {
  transform: scale(0.95);
}

.chevron {
  font-size: 0.875rem;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  color: #6b7280;
  font-weight: bold;
}

.expand-toggle.expanded {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-color: #2563eb;
}

.expand-toggle.expanded .chevron {
  transform: rotate(180deg);
  color: white;
}

.expand-toggle.expanded:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  border-color: #1d4ed8;
}

/* Variant Table */
.variant-table-wrapper {
  border-top: 1px solid var(--border, #e5e7eb);
  overflow: hidden;
  background: linear-gradient(to bottom, #fafbfc 0%, #ffffff 100%);
}

.variant-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.variant-table thead {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  backdrop-filter: blur(10px);
}

.variant-table th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-weight: 600;
  color: #495057;
  border-bottom: 2px solid var(--border, #e5e7eb);
  text-transform: uppercase;
  letter-spacing: 0.025em;
  font-size: 0.75rem;
}

.variant-table tbody {
  background: white;
}

.variant-table td {
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

.text-right {
  text-align: right;
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

/* Inline Editing Styles */
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

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
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

.platform-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  position: relative;
}

.sync-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  padding: 0;
  line-height: 1;
  margin-bottom: -2px;
  animation: bounce 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes bounce {
  0% {
    transform: scale(0);
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
  }
}

.sync-spinner {
  width: 12px;
  height: 12px;
  border: 2px solid #e5e7eb;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 2px;
}

.sync-success {
  font-size: 0.75rem;
  margin-bottom: -2px;
  animation: popIn 0.3s ease-out;
}

.sync-error {
  font-size: 0.75rem;
  margin-bottom: -2px;
  cursor: pointer;
  animation: shake 0.5s ease-in-out;
}

@keyframes popIn {
  from {
    transform: scale(0);
  }
  to {
    transform: scale(1);
  }
}

@keyframes shake {
  0%,
  100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-2px);
  }
  75% {
    transform: translateX(2px);
  }
}

.platform-icon {
  font-size: 1.125rem;
  opacity: 0.2;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  filter: grayscale(100%);
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.platform-icon:hover {
  opacity: 0.8;
  filter: grayscale(0%);
  transform: scale(1.2);
}

.platform-icon.linked {
  opacity: 1;
  filter: grayscale(0%);
  transform: scale(1.1);
}

.platform-icon.is-loading {
  opacity: 1;
  filter: grayscale(0%);
  animation: pulse 1.5s infinite;
  cursor: wait;
}

.platform-icon.has-update {
  position: relative;
}

.platform-icon.has-update::after {
  content: "!";
  position: absolute;
  top: -4px;
  right: -4px;
  background: #f59e0b;
  color: white;
  font-size: 0.5rem;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid white;
  font-weight: bold;
}

.platform-icon.linked::after {
  content: "✓";
  position: absolute;
  bottom: -4px;
  right: -4px;
  font-size: 0.5rem;
  background: #10b981;
  color: white;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1.5px solid white;
  font-weight: bold;
}

.platform-icon:hover {
  transform: scale(1.2);
}

.no-skus {
  text-align: center;
  color: var(--text-secondary, #6b7280);
  padding: 2rem !important;
  font-style: italic;
}

/* Expand Animation */
.expand-enter-active {
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
  max-height: 800px;
}

.expand-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.6, 1);
  max-height: 800px;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
  transform: translateY(-10px);
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  transform: translateY(0);
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  margin-top: 1rem;
}

.btn-pagination {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border, #e5e7eb);
  background: var(--surface, white);
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.btn-pagination:hover:not(:disabled) {
  background: var(--surface-hover, #f3f4f6);
}

.btn-pagination:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.875rem;
  color: var(--text-secondary, #6b7280);
}

/* Buttons */
.btn {
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.btn-primary {
  background: var(--primary, #3b82f6);
  color: white;
  border: none;
}

.btn-primary:hover {
  background: var(--primary-hover, #2563eb);
}

.btn-secondary {
  background: var(--surface, white);
  color: var(--text-primary, #1a1a1a);
  border: 1px solid var(--border, #e5e7eb);
}

.btn-secondary:hover {
  background: var(--surface-hover, #f3f4f6);
}

/* Responsive */
@media (max-width: 640px) {
  .card-header {
    flex-wrap: wrap;
  }

  .product-thumbnail {
    width: 60px;
    height: 60px;
  }

  .variant-table-wrapper {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .variant-table {
    font-size: 0.75rem;
    min-width: 600px;
  }

  .variant-table th,
  .variant-table td {
    padding: 0.625rem 0.5rem;
  }

  .sku-cell {
    font-size: 0.75rem;
    min-width: 80px;
  }

  .platform-icon {
    font-size: 1rem;
  }

  .platform-icon.linked::after {
    width: 10px;
    height: 10px;
    font-size: 0.45rem;
  }
}
</style>
