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
        <div class="product-add-page">
          <div class="page-header">
            <router-link to="/master-products" class="back-link"
              >← Kembali</router-link
            >
            <h1>Tambah Produk Baru</h1>
          </div>

          <form @submit.prevent="handleSubmit" class="product-form">
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

            <!-- Section 2: Images (max 8) -->
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
                    <img :src="img" alt="" />
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
                  >
                    <span>+ Tambah Gambar</span>
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

            <!-- Section 3: Variants & SKUs -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('variants')">
                <h2>3. Varian & SKU</h2>
                <span class="sku-count">{{ form.skus.length }}/50 SKU</span>
              </div>
              <div class="section-content" v-show="!sections.variants">
                <div class="sku-list">
                  <div
                    v-for="(sku, idx) in form.skus"
                    :key="idx"
                    class="sku-row"
                  >
                    <input
                      v-model="sku.seller_sku"
                      placeholder="Seller SKU"
                      required
                    />
                    <input
                      v-model="sku.variant_name"
                      placeholder="Nama Varian"
                    />
                    <input
                      v-model.number="sku.price"
                      type="number"
                      placeholder="Harga"
                      min="0"
                      required
                    />
                    <input
                      v-model.number="sku.stock"
                      type="number"
                      placeholder="Stok"
                      min="0"
                      required
                    />
                    <button
                      type="button"
                      @click="removeSku(idx)"
                      class="btn-remove-sku"
                    >
                      ×
                    </button>
                  </div>
                </div>
                <button
                  type="button"
                  @click="addSku"
                  class="btn-add-sku"
                  :disabled="form.skus.length >= 50"
                >
                  + Tambah SKU
                </button>
              </div>
            </section>

            <!-- Section 4: Pricing Summary -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('pricing')">
                <h2>4. Ringkasan Harga</h2>
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

            <!-- Section 5: Status -->
            <section class="form-section">
              <div class="section-header" @click="toggleSection('status')">
                <h2>5. Status Produk</h2>
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
                </div>
              </div>
            </section>

            <!-- Submit -->
            <div class="form-actions">
              <button type="button" @click="saveDraft" class="btn-draft">
                Simpan Draft
              </button>
              <button
                type="submit"
                class="btn-submit"
                :disabled="submitting || !isValid"
              >
                {{ submitting ? "Menyimpan..." : "Simpan Produk" }}
              </button>
            </div>
          </form>

          <!-- Messages -->
          <div v-if="error" class="message error">{{ error }}</div>
          <div v-if="success" class="message success">{{ success }}</div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import Icon from "@/components/ui/Icon.vue";
import masterProductService from "@/services/masterProductService";

const router = useRouter();
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

// Form state
const form = ref({
  title: "",
  description: "",
  images: [] as string[],
  skus: [{ seller_sku: "", variant_name: "", price: 0, stock: 0 }],
  status: "draft",
});

// Section collapse state
const sections = ref({
  basic: false,
  images: false,
  variants: false,
  pricing: false,
  status: false,
});

const submitting = ref(false);
const error = ref("");
const success = ref("");
const imageInput = ref<HTMLInputElement | null>(null);

// Computed
const isValid = computed(() => {
  return (
    form.value.title.length > 0 &&
    form.value.title.length <= 120 &&
    form.value.description.length > 0 &&
    form.value.skus.length > 0 &&
    form.value.skus.every((s) => s.seller_sku && s.price >= 0)
  );
});

const minPrice = computed(() =>
  Math.min(...form.value.skus.map((s) => s.price || 0)),
);
const maxPrice = computed(() =>
  Math.max(...form.value.skus.map((s) => s.price || 0)),
);
const totalStock = computed(() =>
  form.value.skus.reduce((sum, s) => sum + (s.stock || 0), 0),
);

// Methods
const toggleSection = (section: keyof typeof sections.value) => {
  sections.value[section] = !sections.value[section];
};

const addSku = () => {
  if (form.value.skus.length < 50) {
    form.value.skus.push({
      seller_sku: "",
      variant_name: "",
      price: 0,
      stock: 0,
    });
  }
};

const removeSku = (idx: number) => {
  if (form.value.skus.length > 1) {
    form.value.skus.splice(idx, 1);
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

const formatPrice = (price: number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
};

const saveDraft = () => {
  localStorage.setItem("masterProductDraft", JSON.stringify(form.value));
  success.value = "Draft tersimpan!";
  setTimeout(() => (success.value = ""), 2000);
};

const handleSubmit = async () => {
  submitting.value = true;
  error.value = "";

  try {
    await masterProductService.create({
      title: form.value.title,
      description: form.value.description,
      images: form.value.images,
      status: form.value.status,
      skus: form.value.skus,
    });
    success.value = "Produk berhasil dibuat!";
    setTimeout(() => router.push("/master-products"), 1500);
  } catch (err: any) {
    error.value = err.message || "Gagal menyimpan produk";
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped>
@import "../Dashboard.module.css";

.product-add-page {
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

.sku-row {
  display: grid;
  grid-template-columns: 1fr 1fr 120px 100px 40px;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.sku-row input {
  padding: 0.5rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.25rem;
}

.btn-remove-sku {
  background: #fee2e2;
  color: #ef4444;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
}

.btn-add-sku {
  padding: 0.5rem 1rem;
  background: #eff6ff;
  color: #3b82f6;
  border: 1px solid #3b82f6;
  border-radius: 0.375rem;
  cursor: pointer;
  margin-top: 0.5rem;
}

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

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

.btn-draft {
  padding: 0.75rem 1.5rem;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  cursor: pointer;
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

.message {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 1rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 500;
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
