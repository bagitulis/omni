<template>
  <div class="export-dropdown" ref="exportDropdownRef">
    <button
      @click="toggleExportMenu"
      :disabled="loading || disabled"
      class="btn-export"
      title="Export options"
      aria-label="Export options"
    >
      <i class="pi pi-download" aria-hidden="true"></i>
      <span class="btn-text">Export</span>
      <i class="pi pi-chevron-down export-chevron" aria-hidden="true"></i>
    </button>

    <div v-if="showExportMenu" class="export-menu">
      <button
        class="export-menu-item"
        @click="handleExportToN8N"
        title="Send data to N8N for automation"
      >
        <i class="pi pi-send" aria-hidden="true"></i>
        <span>Export to N8N</span>
      </button>
      <button
        class="export-menu-item"
        @click="handleExportToCSV"
        title="Download CSV file"
      >
        <i class="pi pi-file" aria-hidden="true"></i>
        <span>Download CSV</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";

defineProps<{
  loading: boolean;
  disabled: boolean;
}>();

const emit = defineEmits<{
  (e: "export-n8n"): void;
  (e: "export-csv"): void;
}>();

const showExportMenu = ref(false);
const exportDropdownRef = ref<HTMLElement | null>(null);

const toggleExportMenu = () => {
  showExportMenu.value = !showExportMenu.value;
};

const closeExportMenu = () => {
  showExportMenu.value = false;
};

const handleExportToN8N = () => {
  closeExportMenu();
  emit("export-n8n");
};

const handleExportToCSV = () => {
  closeExportMenu();
  emit("export-csv");
};

const handleClickOutside = (event: MouseEvent) => {
  if (
    exportDropdownRef.value &&
    !exportDropdownRef.value.contains(event.target as Node)
  ) {
    closeExportMenu();
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<style scoped>
@import "./OrderManager.styles.css";

/* Export Dropdown Styles */
.export-dropdown {
  position: relative;
  display: inline-block;
}

.btn-export {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.export-chevron {
  font-size: 0.7rem;
  margin-left: 0.25rem;
}

.export-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 0.25rem;
  background: white;
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-md);
  box-shadow: var(--om-shadow-lg);
  min-width: 180px;
  z-index: 100;
  overflow: hidden;
}

.export-menu-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem 1rem;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: var(--om-font-sm);
  color: var(--om-text-primary);
  text-align: left;
  transition: background-color var(--om-transition-fast);
}

.export-menu-item:hover {
  background-color: var(--om-bg-secondary);
}

.export-menu-item:first-child {
  border-bottom: 1px solid var(--om-border);
}

.export-menu-item i {
  color: var(--om-text-secondary);
  font-size: 1rem;
}

.export-menu-item:hover i {
  color: var(--om-text-primary);
}
</style>
