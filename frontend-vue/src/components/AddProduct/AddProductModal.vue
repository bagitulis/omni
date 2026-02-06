<template>
  <Modal :is-open="show" :title="modalTitle" @close="closeModal">
    <ProgressSteps :steps="steps" :current-step="currentStep" />

    <div class="step-content-container">
      <!-- Step 1: Basic Info + Brand -->
      <BasicInfoStep
        v-if="currentStep === 0"
        v-model:title="form.title"
        v-model:description="form.description"
        v-model:categoryId="form.categoryId"
        v-model:selectedCategoryName="selectedCategoryName"
        v-model:brandId="form.brandId"
        :categories="categories"
        :recommended-category="recommendedCategory"
        :popular-categories="popularCategories"
        :categories-loading="loading"
        :brands="brands"
        :brands-loading="loading"
        @search-categories="searchCategories"
        @category-selected="handleCategorySelected"
        @fetch-recommended="handleFetchRecommended"
      />

      <!-- Step 2: Category Attributes -->
      <CategoryAttributesStep
        v-if="currentStep === 1"
        :attributes="categoryAttributes"
        :model-value="form.attributes"
        :loading="attributesLoading"
        @update:model-value="form.attributes = $event"
      />

      <!-- Step 3: Images & SKU -->
      <ImagesSkuStep
        v-if="currentStep === 2"
        :images="form.images"
        :video-url="form.videoUrl"
        :skus="form.skus"
        :has-variants="form.hasVariants"
        :variant-types="form.variantTypes"
        :platform="activePlatform"
        @add-image="addImage"
        @remove-image="removeImage"
        @reorder-images="reorderImages"
        @update:video-url="form.videoUrl = $event"
        @update:has-variants="form.hasVariants = $event"
        @update:variant-types="form.variantTypes = $event"
        @add-sku="addSku"
        @remove-sku="removeSku"
        @update-sku="updateSkuField"
        @generate-skus="generateSkusFromVariants"
        @bulk-update-skus="bulkUpdateSkus"
      />

      <!-- Step 4: Shipping + Delivery -->
      <ShippingStep
        v-if="currentStep === 3"
        v-model:packageWeight="form.packageWeight"
        v-model:dimensions="form.packageDimensions"
        v-model:saveMode="form.saveMode"
        v-model:deliveryOptionIds="form.deliveryOptionIds"
        :delivery-options="deliveryOptions"
        :delivery-loading="loading"
      />

      <div v-if="error" class="message error" role="alert">
        <i class="pi pi-exclamation-circle" aria-hidden="true"></i>
        {{ error }}
      </div>

      <div v-if="loading" class="loading-overlay">
        <i class="pi pi-spin pi-spinner" aria-hidden="true"></i>
        <span>Processing...</span>
      </div>
    </div>

    <template #footer>
      <button 
        v-if="currentStep > 0" 
        type="button"
        class="btn btn-secondary"
        :disabled="loading"
        @click="prevStep"
      >
        <i class="pi pi-arrow-left" aria-hidden="true"></i> Back
      </button>
      <div class="spacer"></div>
      <button type="button" class="btn btn-secondary" :disabled="loading" @click="closeModal">
        Cancel
      </button>
      <button
        v-if="currentStep < steps.length - 1"
        type="button"
        class="btn btn-primary"
        :disabled="!canProceed || loading"
        @click="nextStep"
      >
        Next <i class="pi pi-arrow-right" aria-hidden="true"></i>
      </button>
      <button
        v-else
        type="button"
        :class="['btn', platformButtonClass]"
        :disabled="!isFormValid || loading"
        @click="handleSubmit"
      >
        <i class="pi pi-check" aria-hidden="true"></i> Create Product
      </button>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import Modal from "@/components/Modal.vue";
import ProgressSteps from "./ProgressSteps.vue";
import BasicInfoStep from "./BasicInfoStep.vue";
import CategoryAttributesStep from "./CategoryAttributesStep.vue";
import ImagesSkuStep from "./ImagesSkuStep.vue";
import ShippingStep from "./ShippingStep.vue";
import { useProductCreate, type Platform } from "@/composables/useProductCreate";

const props = defineProps<{
  show: boolean;
  platform?: Platform;
}>();

const emit = defineEmits<{
  close: [];
  success: [productId: string];
}>();

const activePlatform = computed(() => props.platform || "tiktok");

const modalTitle = computed(() => {
  const titles: Record<Platform, string> = {
    tiktok: "Add Product to TikTok Shop",
    shopee: "Add Product to Shopee",
    lazada: "Add Product to Lazada",
  };
  return titles[activePlatform.value];
});

const platformButtonClass = computed(() => {
  const classes: Record<Platform, string> = {
    tiktok: "btn-tiktok",
    shopee: "btn-shopee",
    lazada: "btn-lazada",
  };
  return classes[activePlatform.value];
});

const {
  loading, error, categories, categoryAttributes, brands, deliveryOptions,
  recommendedCategory, popularCategories, form, isFormValid,
  searchCategories, fetchCategoryAttributes, fetchRecommendedCategory, fetchPopularCategories,
  fetchBrands, fetchDeliveryOptions,
  addSku, removeSku, addImage, removeImage, reorderImages,
  generateSkusFromVariants, bulkUpdateSkus,
  createProduct, resetForm,
} = useProductCreate(activePlatform.value);

const currentStep = ref(0);
const selectedCategoryName = ref("");
const attributesLoading = ref(false);

// Updated steps array (4 steps now)
const steps = [
  { id: "basic", label: "Basic Info" },
  { id: "attributes", label: "Attributes" },
  { id: "images", label: "Images & SKU" },
  { id: "shipping", label: "Shipping" },
];

const canProceed = computed(() => {
  if (currentStep.value === 0) {
    return form.value.title && form.value.description && form.value.categoryId;
  }
  if (currentStep.value === 1) {
    // Check if all required attributes are filled
    const requiredAttrs = categoryAttributes.value.filter(a => a.required);
    return requiredAttrs.every(a => form.value.attributes[a.id]);
  }
  if (currentStep.value === 2) {
    return form.value.images.length > 0 && 
           form.value.skus.every(s => s.sellerSku && s.price > 0);
  }
  return true;
});

async function handleCategorySelected(categoryId: string) {
  attributesLoading.value = true;
  try {
    await Promise.all([
      fetchCategoryAttributes(categoryId),
      fetchBrands(categoryId),
      fetchDeliveryOptions(),
    ]);
  } finally {
    attributesLoading.value = false;
  }
}

async function handleFetchRecommended() {
  // Fetch popular categories on mount
  await fetchPopularCategories();
  // Fetch recommended if title exists
  if (form.value.title) {
    await fetchRecommendedCategory(form.value.title, form.value.description);
  }
}

function closeModal() {
  resetForm();
  currentStep.value = 0;
  selectedCategoryName.value = "";
  emit("close");
}

function nextStep() {
  if (currentStep.value < steps.length - 1) currentStep.value++;
}

function prevStep() {
  if (currentStep.value > 0) currentStep.value--;
}

function updateSkuField(index: number, field: string, value: string | number) {
  const sku = form.value.skus[index];
  if (field === "sellerSku") sku.sellerSku = value as string;
  else if (field === "price") sku.price = value as number;
  else if (field === "stock") sku.stock = value as number;
}

async function handleSubmit() {
  const result = await createProduct();
  if (result.success && result.productId) {
    emit("success", result.productId);
    closeModal();
  }
}

watch(() => props.show, (newVal) => {
  if (!newVal) currentStep.value = 0;
});
</script>

<style scoped>
.step-content-container {
  position: relative;
  min-height: 200px;
  padding-top: 16px;
}

.message {
  margin-top: 16px;
  padding: 12px;
  border-radius: 6px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.message.error {
  background: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font-size: 14px;
  color: #1d4ed8;
}

.loading-overlay i { font-size: 32px; }
.spacer { flex: 1; }

/* Buttons */
.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }

.btn-secondary { background: #e5e7eb; color: #374151; }
.btn-secondary:hover:not(:disabled) { background: #d1d5db; }

.btn-tiktok { background: #1d4ed8; color: white; }
.btn-tiktok:hover:not(:disabled) { background: #1e40af; }

.btn-shopee { background: #f53d2d; color: white; }
.btn-shopee:hover:not(:disabled) { background: #d93025; }

.btn-lazada { background: #0f1689; color: white; }
.btn-lazada:hover:not(:disabled) { background: #0a0f5c; }
</style>
