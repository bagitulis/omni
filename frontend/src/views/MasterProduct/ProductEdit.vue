<template>
  <div class="master-product-page">
    <div class="main-layout">
      <LeftSidebar
        :collapsed="uiStore.leftSidebarCollapsed"
        :active-tab="'master-products'"
        :active-platform="uiStore.activePlatform"
        @toggle="uiStore.toggleLeftSidebar"
        @tab-change="handleTabChange"
        @platform-change="handlePlatformChange"
      />
      <main class="main-content">
        <div class="product-edit-page">
          <div class="page-header">
            <router-link to="/master-products" class="back-link">
              ← Kembali
            </router-link>
            <h1>Edit Produk</h1>
          </div>

          <!-- Loading State -->
          <div v-if="loading" class="loading-state">
            <div class="spinner"></div>
            <p>Memuat data produk...</p>
          </div>

          <!-- Error State -->
          <div v-else-if="loadError" class="error-state">
            <p>{{ loadError }}</p>
            <button @click="loadProduct" class="btn-retry">Coba Lagi</button>
          </div>

          <!-- Edit Form -->
          <form v-else @submit.prevent="handleSubmit" class="product-form">
            <!-- Section 1: Basic Info -->
            <section
              class="form-section"
              :class="{ collapsed: sections.basic }"
            >
              <div class="section-header" @click="toggleSection('basic')">
                <h2>1. Informasi Dasar</h2>
                <Icon
                  :name="sections.basic ? 'chevron-down' : 'chevron-up'"
                  size="sm"
                />
              </div>
              <div class="section-content" v-show="!sections.basic">
                <div class="form-group">
                  <label>Nama Produk <span class="required">*</span></label>
                  <input v-model="form.title" maxlength="120" required />
                  <span
                    class="char-count"
                    :class="{ warning: form.title.length > 100 }"
                  >
                    {{ form.title.length }}/120
                  </span>
                </div>
                <div class="form-group">
                  <label>Deskripsi <span class="required">*</span></label>
                  <textarea
                    v-model="form.description"
                    maxlength="5000"
                    rows="6"
                    required
                  ></textarea>
                  <span class="char-count"
                    >{{ form.description.length }}/5000</span
                  >
                </div>
              </div>
            </section>

            <!-- Section 2: Images -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('images')">
                <h2>2. Gambar Produk</h2>
                <span class="image-count">{{ form.images.length }}/8</span>
              </div>
              <div class="section-content" v-show="!sections.images">
                <div class="image-grid">
                  <div
                    v-for="(img, idx) in form.images"
                    :key="idx"
                    class="image-slot"
                  >
                    <img :src="img" alt="" loading="lazy" />
                    <button
                      type="button"
                      @click="removeImage(idx)"
                      class="btn-remove"
                    >
                      ×
                    </button>
                  </div>
                  <div
                    v-if="form.images.length < 8"
                    class="image-slot add-slot"
                    @click="triggerImageUpload"
                    title="Upload dari Komputer"
                  >
                    <span>+ Upload</span>
                  </div>
                  <div
                    v-if="form.images.length < 8"
                    class="image-slot gallery-slot"
                    @click="openGallery"
                    title="Pilih dari Galeri"
                  >
                    <span>+ Galeri</span>
                  </div>
                </div>
                <input
                  type="file"
                  ref="imageInput"
                  @change="handleImageUpload"
                  accept="image/*"
                  hidden
                />
              </div>
            </section>

            <!-- Section 3: Variants & SKUs (Read-only display) -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('variants')">
                <h2>3. Varian & SKU</h2>
                <span class="sku-count"
                  >{{ product?.skus?.length || 0 }} SKU</span
                >
              </div>
              <div class="section-content" v-show="!sections.variants">
                <div class="info-box">
                  <p>
                    <Icon name="warning" size="sm" class="text-yellow-500" />
                    SKU tidak dapat diedit. Untuk mengubah SKU, silakan buat
                    produk baru.
                  </p>
                </div>
                <div class="sku-list-readonly">
                  <div
                    v-for="sku in product?.skus"
                    :key="sku.id"
                    class="sku-row-readonly"
                  >
                    <div class="sku-field">
                      <label>Seller SKU</label>
                      <span>{{ sku.seller_sku }}</span>
                    </div>
                    <div class="sku-field">
                      <label>Varian</label>
                      <span>{{ sku.variant_name || "-" }}</span>
                    </div>
                    <div class="sku-field">
                      <label>Harga</label>
                      <span>{{ formatPrice(sku.price) }}</span>
                    </div>
                    <div class="sku-field">
                      <label>Stok</label>
                      <span>{{ sku.stock }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </section>

            <!-- Section 4: Sync Status -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('sync')">
                <h2>4. Status Sinkronisasi</h2>
              </div>
              <div class="section-content" v-show="!sections.sync">
                <div class="sync-platforms">
                  <div
                    v-for="platform in platforms"
                    :key="platform.name"
                    class="platform-card"
                  >
                    <div class="platform-header">
                      <Icon
                        :name="platform.icon"
                        size="md"
                        class="platform-icon"
                      />
                      <span class="platform-name">{{ platform.label }}</span>
                    </div>
                    <div class="platform-status">
                      <span
                        class="status-badge"
                        :class="getSyncStatusClass(platform.name)"
                      >
                        {{ getSyncStatusLabel(platform.name) }}
                      </span>
                      <span
                        v-if="getLastSyncTime(platform.name)"
                        class="sync-time"
                      >
                        Terakhir: {{ getLastSyncTime(platform.name) }}
                      </span>
                    </div>
                    <button
                      type="button"
                      @click="syncToPlatform(platform.name)"
                      class="btn-sync"
                      :disabled="syncing[platform.name]"
                    >
                      {{
                        syncing[platform.name]
                          ? "Menyinkronkan..."
                          : "Sync Sekarang"
                      }}
                    </button>
                  </div>
                </div>
              </div>
            </section>

            <!-- Section 5: Pricing Summary -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('pricing')">
                <h2>5. Ringkasan Harga</h2>
              </div>
              <div class="section-content" v-show="!sections.pricing">
                <div class="pricing-summary">
                  <div class="price-item">
                    <span>Harga Terendah:</span>
                    <span class="price">{{ formatPrice(minPrice) }}</span>
                  </div>
                  <div class="price-item">
                    <span>Harga Tertinggi:</span>
                    <span class="price">{{ formatPrice(maxPrice) }}</span>
                  </div>
                  <div class="price-item">
                    <span>Total Stok:</span>
                    <span>{{ totalStock }} unit</span>
                  </div>
                </div>
              </div>
            </section>

            <!-- Section 6: Status -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('status')">
                <h2>6. Status Produk</h2>
              </div>
              <div class="section-content" v-show="!sections.status">
                <div class="status-options">
                  <label class="radio-option">
                    <input type="radio" v-model="form.status" value="draft" />
                    <span>Draft</span>
                  </label>
                  <label class="radio-option">
                    <input type="radio" v-model="form.status" value="active" />
                    <span>Aktif</span>
                  </label>
                  <label class="radio-option">
                    <input
                      type="radio"
                      v-model="form.status"
                      value="inactive"
                    />
                    <span>Tidak Aktif</span>
                  </label>
                </div>
              </div>
            </section>

            <!-- Submit -->
            <div class="form-actions">
              <button
                type="submit"
                class="btn-submit"
                :disabled="submitting || !isValid"
              >
                {{ submitting ? "Menyimpan..." : "Simpan Perubahan" }}
              </button>
            </div>
          </form>

          <!-- Messages -->
          <div v-if="error" class="message error">{{ error }}</div>
          <div v-if="success" class="message success">{{ success }}</div>

          <!-- Image Gallery Modal -->
          <ImageGalleryPicker
            v-model:visible="showGallery"
            :max-images="8 - form.images.length"
            @select="handleGallerySelect"
          />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useUIStore } from "@/store/ui";
import Icon from "@/components/ui/Icon.vue";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import masterProductService, {
  type MasterProduct,
  type MasterProductPlatformLink,
} from "@/services/masterProductService";
import ImageGalleryPicker from "@/components/MasterProduct/ImageGalleryPicker.vue";
import type { GalleryImage } from "@/services/imageService";

const router = useRouter();
const route = useRoute();
const uiStore = useUIStore();

// Navigation Handlers
const handleTabChange = (tab: string) => {
  if (tab === "master-products") return;

  if (["settings", "product-management", "order-management"].includes(tab)) {
    router.push({ path: "/dashboard", query: { tab } });
    uiStore.setActiveTab(tab);
  } else {
    router.push({ name: tab });
  }
};

const handlePlatformChange = (platform: string) => {
  uiStore.setActivePlatform(platform);
};

// State
const loading = ref(true);
const loadError = ref("");
const product = ref<MasterProduct | null>(null);
const showGallery = ref(false);

// Form state
const form = ref({
  title: "",
  description: "",
  images: [] as string[],
  status: "draft",
});

// Section collapse state
const sections = ref({
  basic: false,
  images: false,
  variants: false,
  sync: false,
  pricing: false,
  status: false,
});

const submitting = ref(false);
const error = ref("");
const success = ref("");
const imageInput = ref<HTMLInputElement | null>(null);

// Sync state
const syncing = ref<Record<string, boolean>>({
  shopee: false,
  tiktok: false,
  lazada: false,
});

// Platform definitions
const platforms = [
  { name: "shopee", label: "Shopee", icon: "store" },
  { name: "tiktok", label: "TikTok", icon: "chart" },
  { name: "lazada", label: "Lazada", icon: "shopping-bag" },
];

// Computed
const isValid = computed(() => {
  return (
    form.value.title.length > 0 &&
    form.value.title.length <= 120 &&
    form.value.description.length > 0
  );
});

const minPrice = computed(() => {
  if (!product.value?.skus || product.value.skus.length === 0) return 0;
  return Math.min(...product.value.skus.map((s) => s.price || 0));
});

const maxPrice = computed(() => {
  if (!product.value?.skus || product.value.skus.length === 0) return 0;
  return Math.max(...product.value.skus.map((s) => s.price || 0));
});

const totalStock = computed(() => {
  if (!product.value?.skus || product.value.skus.length === 0) return 0;
  return product.value.skus.reduce((sum, s) => sum + (s.stock || 0), 0);
});

// Methods
const toggleSection = (section: keyof typeof sections.value) => {
  sections.value[section] = !sections.value[section];
};

const loadProduct = async () => {
  loading.value = true;
  loadError.value = "";

  try {
    const productId = parseInt(route.params.id as string);
    if (isNaN(productId)) {
      loadError.value = "ID produk tidak valid";
      return;
    }

    const response = await masterProductService.getById(productId);
    if (!response.success || !response.data) {
      loadError.value = "Produk tidak ditemukan";
      return;
    }

    product.value = response.data;

    // Populate form with existing data
    form.value.title = product.value.title;
    form.value.description = product.value.description;
    form.value.status = product.value.status;

    // Images are already an array from the API
    if (product.value.images && Array.isArray(product.value.images)) {
      form.value.images = [...product.value.images].filter(Boolean);
    } else {
      form.value.images = [];
    }
  } catch (err: any) {
    loadError.value = err.message || "Gagal memuat data produk";
  } finally {
    loading.value = false;
  }
};

const triggerImageUpload = () => imageInput.value?.click();

const handleImageUpload = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (file && form.value.images.length < 8) {
    const reader = new FileReader();
    reader.onload = () => {
      form.value.images.push(reader.result as string);
    };
    reader.readAsDataURL(file);
  }
};

const removeImage = (idx: number) => form.value.images.splice(idx, 1);

const openGallery = () => {
  showGallery.value = true;
};

const handleGallerySelect = (images: GalleryImage[]) => {
  // Construct path: /uploads/images/{tenant_id}/{local_path}
  // Note: We need tenant_id. If not in product, we might need to get it from auth store or context.
  // But wait, the image object returned from API has tenant_id.

  images.forEach((img) => {
    // If form.images is full, stop
    if (form.value.images.length >= 8) return;

    // Check if image already added (by filename or path)
    // This is a simple check, might need more robust if path format varies
    const path = `/uploads/images/${img.tenant_id}/${img.local_path}`;
    const alreadyExists = form.value.images.some(
      (existing) => existing.includes(img.filename) || existing === path,
    );

    if (!alreadyExists) {
      form.value.images.push(path);
    }
  });
};

const formatPrice = (price: number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
};

const getSyncStatusClass = (platform: string) => {
  const links = product.value?.skus
    ?.flatMap((sku) => sku.platform_links || [])
    .filter((link) => link.platform === platform);

  if (!links || links.length === 0) return "status-not-synced";
  if (links.some((link) => link.sync_status === "synced"))
    return "status-synced";
  if (links.some((link) => link.sync_status === "pending"))
    return "status-pending";
  return "status-error";
};

const getSyncStatusLabel = (platform: string) => {
  const links = product.value?.skus
    ?.flatMap((sku) => sku.platform_links || [])
    .filter((link) => link.platform === platform);

  if (!links || links.length === 0) return "Belum Sync";
  if (links.some((link) => link.sync_status === "synced"))
    return "Tersinkronkan";
  if (links.some((link) => link.sync_status === "pending")) return "Pending";
  return "Error";
};

const getLastSyncTime = (platform: string): string | null => {
  const links = product.value?.skus
    ?.flatMap((sku) => sku.platform_links || [])
    .filter((link) => link.platform === platform && link.last_synced_at);

  if (!links || links.length === 0) return null;

  const latestSync = links.reduce((latest, link) => {
    const linkDate = new Date(link.last_synced_at!);
    return linkDate > latest ? linkDate : latest;
  }, new Date(0));

  if (latestSync.getTime() === 0) return null;

  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(latestSync);
};

const syncToPlatform = async (platform: string) => {
  syncing.value[platform] = true;
  error.value = "";

  try {
    const productId = parseInt(route.params.id as string);
    const result = await masterProductService.sync(productId, platform);

    success.value = `Sync ke ${platform} berhasil! ${result.skus_synced} SKU disinkronkan`;
    setTimeout(() => (success.value = ""), 3000);

    // Reload product to get updated sync status
    await loadProduct();
  } catch (err: any) {
    error.value = `Gagal sync ke ${platform}: ${err.message || "Unknown error"}`;
  } finally {
    syncing.value[platform] = false;
  }
};

const handleSubmit = async () => {
  submitting.value = true;
  error.value = "";

  try {
    const productId = parseInt(route.params.id as string);

    await masterProductService.update(productId, {
      title: form.value.title,
      description: form.value.description,
      images: form.value.images,
      status: form.value.status,
    });

    success.value = "Produk berhasil diperbarui!";
    setTimeout(() => router.push("/master-products"), 1500);
  } catch (err: any) {
    error.value = err.message || "Gagal memperbarui produk";
  } finally {
    submitting.value = false;
  }
};

// Load product on mount
onMounted(() => {
  loadProduct();
});
</script>

<style scoped>
@import "../Dashboard.module.css";

.product-edit-page {
  padding: 1.5rem;
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.back-link {
  color: #6b7280;
  text-decoration: none;
  font-size: 0.875rem;
}

.page-header h1 {
  margin: 0.5rem 0 0;
  font-size: 1.5rem;
  font-weight: 700;
}

/* Loading & Error States */
.loading-state,
.error-state {
  text-align: center;
  padding: 3rem;
  background: white;
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e5e7eb;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.btn-retry {
  padding: 0.5rem 1.5rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  margin-top: 1rem;
}

/* Form Sections */
.form-section {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: #f9fafb;
  cursor: pointer;
  user-select: none;
}

.section-header h2 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.section-content {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.required {
  color: #ef4444;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
}

.char-count {
  display: block;
  text-align: right;
  font-size: 0.75rem;
  color: #6b7280;
  margin-top: 0.25rem;
}

.char-count.warning {
  color: #f59e0b;
}

/* Images */
.image-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

.image-slot {
  aspect-ratio: 1;
  border: 2px dashed #e5e7eb;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.image-slot img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.add-slot {
  cursor: pointer;
  color: #6b7280;
}

.add-slot:hover {
  border-color: #3b82f6;
  color: #3b82f6;
}

.gallery-slot {
  cursor: pointer;
  color: #6b7280;
  background: #f9fafb;
}

.gallery-slot:hover {
  border-color: #ff6b2c;
  color: #ff6b2c;
  background: #fff5f0;
}

.btn-remove {
  position: absolute;
  top: 0.25rem;
  right: 0.25rem;
  width: 24px;
  height: 24px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
}

/* SKU Read-only Display */
.info-box {
  padding: 1rem;
  background: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 0.375rem;
  margin-bottom: 1rem;
}

.info-box p {
  margin: 0;
  color: #92400e;
}

.sku-list-readonly {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.sku-row-readonly {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.375rem;
}

.sku-field label {
  display: block;
  font-size: 0.75rem;
  color: #6b7280;
  margin-bottom: 0.25rem;
}

.sku-field span {
  font-weight: 500;
}

/* Sync Platforms */
.sync-platforms {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.platform-card {
  padding: 1.25rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.platform-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.platform-icon {
  font-size: 1.5rem;
}

.platform-name {
  font-weight: 600;
}

.platform-status {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 1rem;
  font-size: 0.75rem;
  font-weight: 500;
  width: fit-content;
}

.status-not-synced {
  background: #f3f4f6;
  color: #6b7280;
}

.status-synced {
  background: #dcfce7;
  color: #166534;
}

.status-pending {
  background: #fef3c7;
  color: #92400e;
}

.status-error {
  background: #fee2e2;
  color: #991b1b;
}

.sync-time {
  font-size: 0.75rem;
  color: #6b7280;
}

.btn-sync {
  padding: 0.5rem 1rem;
  background: #eff6ff;
  color: #3b82f6;
  border: 1px solid #3b82f6;
  border-radius: 0.375rem;
  cursor: pointer;
  font-weight: 500;
}

.btn-sync:hover {
  background: #dbeafe;
}

.btn-sync:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Pricing Summary */
.pricing-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.price-item {
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.375rem;
  text-align: center;
}

.price-item .price {
  display: block;
  font-size: 1.25rem;
  font-weight: 700;
  color: #059669;
  margin-top: 0.5rem;
}

/* Status Options */
.status-options {
  display: flex;
  gap: 2rem;
}

.radio-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

.btn-submit {
  padding: 0.75rem 2rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-weight: 600;
  cursor: pointer;
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Messages */
.message {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 1rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 500;
  z-index: 1000;
}

.message.success {
  background: #dcfce7;
  color: #166534;
}

.message.error {
  background: #fee2e2;
  color: #991b1b;
}
</style>
