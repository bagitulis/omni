<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>⚙️ Inventory Column Configuration</h3>
        <button
          @click="close"
          class="btn-close"
          type="button"
          aria-label="Close modal"
        >
          ✕
        </button>
      </div>

      <div class="modal-body">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <span>Loading columns...</span>
        </div>

        <div v-else class="columns-config">
          <div class="info-text">
            <p>
              Select which columns from Google Sheets to display in the
              inventory table.
            </p>
            <p class="text-sm">
              Total: {{ availableColumns.length }} (A-{{
                getLastColumnLetter(availableColumns.length)
              }})
            </p>
          </div>

          <!-- Selected Columns -->
          <div v-if="selectedColumns.length > 0" class="selected-section">
            <h4>Selected ({{ selectedColumns.length }})</h4>
            <div class="selected-list">
              <div
                v-for="colName in selectedColumns"
                :key="colName"
                class="selected-item"
              >
                <span>{{ colName }}</span>
                <button @click="removeColumn(colName)" class="btn-remove">
                  ✕
                </button>
              </div>
            </div>
          </div>

          <!-- Available Columns -->
          <div class="available-section">
            <h4>
              Available
              <button @click="selectAll" class="btn-small">Select All</button>
              <button @click="clearAll" class="btn-small btn-secondary">
                Clear All
              </button>
            </h4>

            <div v-if="availableColumns.length === 0" class="empty-state">
              <p>No columns available. Configure spreadsheet first.</p>
            </div>

            <div v-else class="columns-grid">
              <ColumnCheckbox
                v-for="column in availableColumns"
                :key="column.name"
                :column="column"
                :is-selected="isSelected(column.name)"
                @toggle="toggleColumn(column.name)"
              />
            </div>
          </div>

          <div v-if="errorMessage" class="alert alert-error">
            {{ errorMessage }}
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="close" class="btn btn-secondary">Cancel</button>
        <button
          @click="handleSave"
          :disabled="saving || loading"
          class="btn btn-primary"
        >
          {{ saving ? "Saving..." : "Save" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useColumnConfiguration } from "./composables/useColumnConfiguration";
import ColumnCheckbox from "./ColumnCheckbox.vue";

const props = defineProps<{ isOpen: boolean }>();
const emit = defineEmits<{ close: []; saved: [] }>();

const {
  loadAvailableColumns,
  loadSelectedColumns,
  saveColumnConfiguration,
  getLastColumnLetter,
} = useColumnConfiguration();

const loading = ref(false);
const saving = ref(false);
const availableColumns = ref<any[]>([]);
const selectedColumns = ref<string[]>([]);
const errorMessage = ref("");

watch(
  () => props.isOpen,
  async (newVal) => {
    if (newVal) {
      try {
        loading.value = true;
        errorMessage.value = "";
        const [available, selected] = await Promise.all([
          loadAvailableColumns(),
          loadSelectedColumns(),
        ]);
        availableColumns.value = available;
        selectedColumns.value = selected;
      } catch (error: any) {
        errorMessage.value = error.message;
      } finally {
        loading.value = false;
      }
    }
  }
);

function isSelected(columnName: string): boolean {
  return selectedColumns.value.includes(columnName);
}

function toggleColumn(columnName: string) {
  const index = selectedColumns.value.indexOf(columnName);
  if (index > -1) {
    selectedColumns.value.splice(index, 1);
  } else {
    selectedColumns.value.push(columnName);
  }
}

function removeColumn(columnName: string) {
  const index = selectedColumns.value.indexOf(columnName);
  if (index > -1) {
    selectedColumns.value.splice(index, 1);
  }
}

function selectAll() {
  selectedColumns.value = availableColumns.value.map((col) => col.name);
}

function clearAll() {
  selectedColumns.value = [];
}

async function handleSave() {
  try {
    saving.value = true;
    errorMessage.value = "";
    await saveColumnConfiguration(selectedColumns.value);
    emit("saved");
    close();
  } catch (error: any) {
    errorMessage.value = error.message;
  } finally {
    saving.value = false;
  }
}

function close() {
  emit("close");
}
</script>

<style scoped>
@import "./InventoryColumnsModal.styles.css";
</style>
