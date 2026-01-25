<template>
  <div class="step-content">
    <!-- Image Uploader -->
    <ImageUploader
      :images="images"
      :max-images="9"
      @add-image="$emit('add-image', $event)"
      @remove-image="$emit('remove-image', $event)"
      @reorder-images="$emit('reorder-images', $event)"
    />

    <!-- Video Uploader -->
    <VideoUploader
      :video-url="videoUrl"
      :platform="platform"
      @update:video-url="$emit('update:videoUrl', $event)"
    />

    <!-- Variant Builder -->
    <VariantBuilder
      :has-variants="hasVariants"
      :variant-types="variantTypes"
      @update:has-variants="$emit('update:hasVariants', $event)"
      @update:variant-types="handleVariantTypesChange"
    />

    <!-- SKU Table -->
    <SkuTable
      :skus="skus"
      :has-variants="hasVariants"
      @add-sku="$emit('add-sku')"
      @remove-sku="$emit('remove-sku', $event)"
      @update-sku="handleUpdateSku"
      @bulk-update="handleBulkUpdate"
    />
  </div>
</template>

<script setup lang="ts">
import ImageUploader from "./ImageUploader.vue";
import VideoUploader from "./VideoUploader.vue";
import VariantBuilder from "./VariantBuilder.vue";
import SkuTable from "./SkuTable.vue";

interface ProductSku {
  sellerSku: string;
  price: number;
  stock: number;
  variantLabel?: string;
}

interface VariantType {
  name: string;
  values: string[];
}

defineProps<{
  images: string[];
  videoUrl: string;
  skus: ProductSku[];
  hasVariants: boolean;
  variantTypes: VariantType[];
  platform?: "tiktok" | "shopee" | "lazada";
}>();

const emit = defineEmits<{
  "add-image": [url: string];
  "remove-image": [index: number];
  "reorder-images": [images: string[]];
  "update:videoUrl": [url: string];
  "update:hasVariants": [value: boolean];
  "update:variantTypes": [value: VariantType[]];
  "add-sku": [];
  "remove-sku": [index: number];
  "update-sku": [index: number, field: string, value: string | number];
  "generate-skus": [combinations: string[][]];
  "bulk-update-skus": [field: string, value: number];
}>();

function handleVariantTypesChange(types: VariantType[]) {
  emit("update:variantTypes", types);
  
  // Generate SKU combinations from variants
  const validTypes = types.filter(t => t.name && t.values.length > 0);
  if (validTypes.length === 0) return;

  const combinations = validTypes.reduce<string[][]>(
    (acc, type) => {
      if (acc.length === 0) return type.values.map(v => [v]);
      return acc.flatMap(combo => type.values.map(v => [...combo, v]));
    },
    []
  );
  
  emit("generate-skus", combinations);
}

function handleBulkUpdate(field: string, value: number) {
  emit("bulk-update-skus", field, value);
}

function handleUpdateSku(index: number, field: string, value: string | number) {
  emit("update-sku", index, field, value);
}
</script>

<style scoped>
@import "./AddProductForm.styles.css";

.step-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
</style>
