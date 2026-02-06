<template>
  <div class="step-content">
    <div class="form-group">
      <label for="product-title">Product Title *</label>
      <input
        id="product-title"
        :value="title"
        type="text"
        placeholder="Enter product title"
        required
        @input="handleTitleChange"
      />
    </div>

    <div class="form-group">
      <label for="product-description">Description *</label>
      <textarea
        id="product-description"
        :value="description"
        rows="4"
        placeholder="Enter product description"
        required
        @input="$emit('update:description', ($event.target as HTMLTextAreaElement).value)"
      ></textarea>
    </div>

    <!-- Category Selector with Dropdown and Recommendations -->
    <CategorySelector
      :model-value="categoryId"
      :categories="categories"
      :recommended-category="recommendedCategory"
      :popular-categories="popularCategories"
      :loading="categoriesLoading"
      :selected-category-name="selectedCategoryName"
      @update:model-value="$emit('update:categoryId', $event)"
      @update:selected-category-name="$emit('update:selectedCategoryName', $event)"
      @search="$emit('search-categories', $event)"
      @fetch-recommended="handleFetchRecommended"
      @category-selected="$emit('category-selected', $event)"
    />

    <!-- Brand Selector (shown when category is selected) -->
    <div v-if="categoryId" class="form-group">
      <BrandSelector
        :model-value="brandId"
        :brands="brands"
        :loading="brandsLoading"
        input-id="brand-selector"
        @update:model-value="$emit('update:brandId', $event)"
        @search="$emit('search-brands', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import CategorySelector from "./CategorySelector.vue";
import BrandSelector from "./BrandSelector.vue";

interface Category {
  id: string;
  name: string;
  isLeaf: boolean;
}

interface Brand {
  id: string;
  name: string;
}

defineProps<{
  title: string;
  description: string;
  categoryId: string;
  categories: Category[];
  selectedCategoryName: string;
  recommendedCategory?: Category | null;
  popularCategories?: Category[];
  categoriesLoading?: boolean;
  brandId: string;
  brands: Brand[];
  brandsLoading?: boolean;
}>();

const emit = defineEmits<{
  "update:title": [value: string];
  "update:description": [value: string];
  "update:categoryId": [value: string];
  "update:selectedCategoryName": [value: string];
  "update:brandId": [value: string];
  "search-categories": [keyword: string];
  "search-brands": [keyword: string];
  "category-selected": [categoryId: string];
  "fetch-recommended": [title: string];
}>();

function handleTitleChange(event: Event) {
  const value = (event.target as HTMLInputElement).value;
  emit("update:title", value);
}

function handleFetchRecommended() {
  // Trigger fetch recommended when component mounts or title changes
  emit("fetch-recommended", "");
}
</script>

<style scoped>
@import "./AddProductForm.styles.css";
</style>
