<template>
  <div class="brand-selector">
    <label :for="inputId" class="form-label">
      Brand
      <span v-if="required" class="required">*</span>
    </label>
    
    <div class="select-wrapper">
      <input
        :id="inputId"
        v-model="searchQuery"
        type="text"
        class="form-input"
        :placeholder="selectedBrandName || 'Search brand...'"
        @focus="showDropdown = true"
        @input="handleSearch"
      />
      <i 
        v-if="modelValue" 
        class="pi pi-times clear-btn" 
        @click="clearSelection"
        aria-label="Clear selection"
      ></i>
      <i v-else class="pi pi-chevron-down dropdown-icon"></i>
    </div>

    <div v-if="showDropdown && filteredBrands.length > 0" class="dropdown-list">
      <div
        v-for="brand in filteredBrands"
        :key="brand.id"
        class="dropdown-item"
        :class="{ selected: brand.id === modelValue }"
        @click="selectBrand(brand)"
      >
        {{ brand.name }}
      </div>
    </div>

    <div v-if="showDropdown && filteredBrands.length === 0 && searchQuery" class="dropdown-empty">
      No brands found
    </div>

    <div v-if="loading" class="loading-indicator">
      <i class="pi pi-spin pi-spinner"></i> Loading brands...
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";

interface Brand {
  id: string;
  name: string;
}

const props = defineProps<{
  modelValue: string;
  brands: Brand[];
  loading?: boolean;
  required?: boolean;
  inputId?: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
  search: [query: string];
}>();

const searchQuery = ref("");
const showDropdown = ref(false);

const selectedBrandName = computed(() => {
  const brand = props.brands.find((b) => b.id === props.modelValue);
  return brand?.name || "";
});

const filteredBrands = computed(() => {
  if (!searchQuery.value) return props.brands.slice(0, 20);
  const query = searchQuery.value.toLowerCase();
  return props.brands.filter((b) => 
    b.name.toLowerCase().includes(query)
  ).slice(0, 20);
});

function handleSearch() {
  emit("search", searchQuery.value);
}

function selectBrand(brand: Brand) {
  emit("update:modelValue", brand.id);
  searchQuery.value = "";
  showDropdown.value = false;
}

function clearSelection() {
  emit("update:modelValue", "");
  searchQuery.value = "";
}

// Close dropdown on outside click
function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement;
  if (!target.closest(".brand-selector")) {
    showDropdown.value = false;
  }
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});

watch(() => props.modelValue, () => {
  if (props.modelValue) {
    showDropdown.value = false;
  }
});
</script>

<style scoped>
.brand-selector {
  position: relative;
}

.form-label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  color: #374151;
}

.required {
  color: #ef4444;
}

.select-wrapper {
  position: relative;
}

.form-input {
  width: 100%;
  padding: 10px 36px 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.dropdown-icon,
.clear-btn {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
  font-size: 12px;
}

.clear-btn {
  cursor: pointer;
  padding: 4px;
}

.clear-btn:hover {
  color: #ef4444;
}

.dropdown-list {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  max-height: 200px;
  overflow-y: auto;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  margin-top: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
}

.dropdown-item {
  padding: 10px 12px;
  cursor: pointer;
  transition: background 0.15s;
}

.dropdown-item:hover {
  background: #f3f4f6;
}

.dropdown-item.selected {
  background: #eff6ff;
  color: #1d4ed8;
  font-weight: 500;
}

.dropdown-empty {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  padding: 12px;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  margin-top: 4px;
  text-align: center;
  color: #6b7280;
  font-size: 14px;
  z-index: 100;
}

.loading-indicator {
  margin-top: 8px;
  font-size: 13px;
  color: #6b7280;
}

.loading-indicator i {
  margin-right: 6px;
}
</style>
