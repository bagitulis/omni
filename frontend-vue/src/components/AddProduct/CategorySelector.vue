<template>
  <div class="category-selector">
    <label for="category-input">Category *</label>
    
    <!-- Recommended Category Banner -->
    <div v-if="recommendedCategory && !selectedCategory" class="recommended-banner">
      <div class="recommended-content">
        <i class="pi pi-lightbulb" aria-hidden="true"></i>
        <div class="recommended-text">
          <span class="rec-label">Recommended:</span>
          <span class="rec-name">{{ recommendedCategory.name }}</span>
        </div>
      </div>
      <button type="button" class="btn btn-sm btn-primary" @click="applyRecommended">
        Apply
      </button>
    </div>

    <!-- Search Input -->
    <div class="search-wrapper">
      <i class="pi pi-search search-icon" aria-hidden="true"></i>
      <input
        id="category-input"
        ref="searchInput"
        v-model="searchQuery"
        type="text"
        placeholder="Search or select category..."
        autocomplete="off"
        @focus="showDropdown = true"
        @input="handleSearch"
      />
      <button
        v-if="selectedCategory"
        type="button"
        class="clear-btn"
        aria-label="Clear selection"
        @click="clearSelection"
      >
        <i class="pi pi-times" aria-hidden="true"></i>
      </button>
      <i
        class="pi pi-chevron-down dropdown-icon"
        :class="{ open: showDropdown }"
        aria-hidden="true"
        @click="toggleDropdown"
      ></i>
    </div>

    <!-- Selected Category Display -->
    <div v-if="selectedCategory" class="selected-display">
      <i class="pi pi-check-circle" aria-hidden="true"></i>
      <span>{{ selectedCategory.name }}</span>
      <button type="button" class="change-btn" @click="focusSearch">Change</button>
    </div>

    <!-- Dropdown -->
    <div v-if="showDropdown" class="dropdown-panel">
      <div v-if="loading" class="dropdown-loading">
        <i class="pi pi-spin pi-spinner" aria-hidden="true"></i>
        Loading categories...
      </div>

      <template v-else>
        <!-- Recent/Popular Categories -->
        <div v-if="!searchQuery && (popularCategories?.length ?? 0) > 0" class="dropdown-section">
          <div class="section-header">Popular Categories</div>
          <div
            v-for="cat in (popularCategories ?? [])"
            :key="cat.id"
            class="dropdown-item"
            @click="selectCategory(cat)"
          >
            <span>{{ cat.name }}</span>
            <i v-if="cat.isLeaf" class="pi pi-check" aria-hidden="true"></i>
          </div>
        </div>

        <!-- Search Results -->
        <div v-if="categories.length" class="dropdown-section">
          <div v-if="searchQuery" class="section-header">Search Results</div>
          <div
            v-for="cat in categories"
            :key="cat.id"
            class="dropdown-item"
            :class="{ 'is-leaf': cat.isLeaf, selected: selectedCategory?.id === cat.id }"
            @click="selectCategory(cat)"
          >
            <span class="cat-name">{{ cat.name }}</span>
            <span v-if="!cat.isLeaf" class="has-children">
              <i class="pi pi-chevron-right" aria-hidden="true"></i>
            </span>
            <i v-else class="pi pi-check leaf-icon" aria-hidden="true"></i>
          </div>
        </div>

        <!-- No Results -->
        <div v-if="searchQuery && !categories.length && !loading" class="no-results">
          No categories found for "{{ searchQuery }}"
        </div>
      </template>
    </div>

    <!-- Click outside to close -->
    <div v-if="showDropdown" class="backdrop" @click="showDropdown = false"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue";

interface Category {
  id: string;
  name: string;
  isLeaf: boolean;
  parentId?: string;
}

const props = defineProps<{
  modelValue: string;
  categories: Category[];
  recommendedCategory?: Category | null;
  popularCategories?: Category[];
  loading?: boolean;
  selectedCategoryName?: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [id: string];
  "update:selectedCategoryName": [name: string];
  "search": [keyword: string];
  "fetch-recommended": [];
  "category-selected": [categoryId: string];
}>();

const searchInput = ref<HTMLInputElement | null>(null);
const searchQuery = ref("");
const showDropdown = ref(false);
const selectedCategory = ref<Category | null>(null);

// Initialize from prop
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.categories.length) {
    const found = props.categories.find(c => c.id === newVal);
    if (found) selectedCategory.value = found;
  }
}, { immediate: true });

// Debounced search
let searchTimeout: ReturnType<typeof setTimeout>;
function handleSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    if (searchQuery.value.length >= 2) {
      emit("search", searchQuery.value);
    }
  }, 300);
}

function selectCategory(cat: Category) {
  if (cat.isLeaf) {
    selectedCategory.value = cat;
    searchQuery.value = "";
    showDropdown.value = false;
    emit("update:modelValue", cat.id);
    emit("update:selectedCategoryName", cat.name);
    emit("category-selected", cat.id);
  } else {
    // Navigate into subcategory (trigger fetch children)
    emit("search", cat.name);
  }
}

function applyRecommended() {
  if (props.recommendedCategory) {
    selectCategory(props.recommendedCategory);
  }
}

function clearSelection() {
  selectedCategory.value = null;
  searchQuery.value = "";
  emit("update:modelValue", "");
  emit("update:selectedCategoryName", "");
}

function toggleDropdown() {
  showDropdown.value = !showDropdown.value;
}

function focusSearch() {
  showDropdown.value = true;
  searchInput.value?.focus();
}

onMounted(() => {
  emit("fetch-recommended");
});
</script>

<style scoped>
.category-selector {
  position: relative;
}

.category-selector > label {
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
  font-size: 14px;
  color: #374151;
}

.recommended-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  margin-bottom: 8px;
}

.recommended-content {
  display: flex;
  align-items: center;
  gap: 10px;
}

.recommended-content i {
  color: #f59e0b;
  font-size: 18px;
}

.recommended-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.rec-label {
  font-size: 11px;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.rec-name {
  font-weight: 600;
  font-size: 13px;
  color: #1d4ed8;
}

.search-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #9ca3af;
  font-size: 14px;
}

.search-wrapper input {
  width: 100%;
  padding: 10px 36px 10px 36px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
}

.search-wrapper input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.clear-btn, .dropdown-icon {
  position: absolute;
  right: 10px;
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
  padding: 4px;
}

.clear-btn { right: 32px; }
.clear-btn:hover { color: #ef4444; }
.dropdown-icon { transition: transform 0.2s; }
.dropdown-icon.open { transform: rotate(180deg); }

.selected-display {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 10px 12px;
  background: #f0fdf4;
  border-radius: 6px;
  color: #166534;
  font-size: 13px;
}

.selected-display i { color: #22c55e; }

.change-btn {
  margin-left: auto;
  background: none;
  border: none;
  color: #3b82f6;
  font-size: 12px;
  cursor: pointer;
  text-decoration: underline;
}

.dropdown-panel {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 280px;
  overflow-y: auto;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1);
  z-index: 100;
}

.dropdown-loading {
  padding: 16px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
}

.dropdown-section { padding: 8px 0; }
.section-header {
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  cursor: pointer;
  transition: background 0.15s;
}

.dropdown-item:hover { background: #f3f4f6; }
.dropdown-item.selected { background: #dbeafe; }
.dropdown-item.is-leaf { font-weight: 500; }

.cat-name { flex: 1; }
.has-children { color: #9ca3af; font-size: 12px; }
.leaf-icon { color: #22c55e; font-size: 12px; }

.no-results {
  padding: 16px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
}

.backdrop {
  position: fixed;
  inset: 0;
  z-index: 99;
}

.btn { padding: 6px 12px; border: none; border-radius: 4px; cursor: pointer; font-weight: 600; font-size: 12px; }
.btn-sm { padding: 6px 12px; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover { background: #2563eb; }
</style>
