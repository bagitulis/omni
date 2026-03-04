<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>
          <i class="pi pi-info-circle" aria-hidden="true"></i>
          Product Details - Item ID: {{ itemId }}
        </h3>
        <button
          class="btn-close"
          @click="$emit('close')"
          aria-label="Close modal"
          title="Close"
        >
          <i class="pi pi-times" aria-hidden="true"></i>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="loading" class="loading-state">
          <i class="pi pi-spin pi-spinner" aria-hidden="true"></i>
          Loading product details...
        </div>

        <div v-else-if="productBase">
          <ProductDetailInfo :product="productBase" :images="productImages" />
          <ProductDetailModels v-if="productBase.has_model" :models="models" />
          <ProductDetailVariations :variations="groupedVariations" />
        </div>

        <div v-else class="error-state">
          <i class="pi pi-exclamation-circle" aria-hidden="true"></i>
          <p>Failed to load product details</p>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="handleCopyItemId" class="btn btn-secondary">
          <i class="pi pi-copy" aria-hidden="true"></i> Copy Item ID
        </button>
        <button @click="$emit('close')" class="btn btn-primary">
          <i class="pi pi-times" aria-hidden="true"></i> Close
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import ProductDetailInfo from "./ProductDetailInfo.vue";
import ProductDetailModels from "./ProductDetailModels.vue";
import ProductDetailVariations from "./ProductDetailVariations.vue";
import { useProductData } from "@/composables/useProductData";
import {
  useProductImages,
  useProductVariations,
} from "@/composables/useProductFormatters";

interface Props {
  itemId: number;
}

const props = defineProps<Props>();
defineEmits<{
  close: [];
}>();

const {
  loading,
  productBase,
  models,
  variations,
  fetchProductDetails,
  copyItemId,
} = useProductData(props.itemId);

const productImages = useProductImages(productBase);
const groupedVariations = useProductVariations(variations);

const handleCopyItemId = () => copyItemId();

onMounted(() => {
  fetchProductDetails();
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 100%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    transform: translateY(-50px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 2px solid #ecf0f1;
}

.modal-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  color: #2c3e50;
  font-size: 1.2rem;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #95a5a6;
  padding: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.btn-close:hover {
  background: #f8f9fa;
  color: #2c3e50;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.loading-state,
.error-state {
  text-align: center;
  padding: 60px 20px;
  color: #95a5a6;
}

.loading-state i,
.error-state i {
  font-size: 3rem;
  margin-bottom: 15px;
  display: block;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 15px 20px;
  border-top: 1px solid #ecf0f1;
  background: #f8f9fa;
}

.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover {
  background: #2980b9;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn-secondary:hover {
  background: #7f8c8d;
}

.modal-body::-webkit-scrollbar {
  width: 8px;
}

.modal-body::-webkit-scrollbar-track {
  background: #f1f1f1;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #bdc3c7;
  border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: #95a5a6;
}

@media (max-width: 768px) {
  .modal-content {
    max-width: 95%;
  }

  .modal-footer {
    flex-direction: column;
  }

  .btn {
    justify-content: center;
  }
}
</style>
