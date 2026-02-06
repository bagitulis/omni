<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click.self="$emit('close-form')">
        <div class="modal-content">
          <div class="modal-header">
            <h3>{{ isEditing ? "Edit Item" : "Add New Item" }}</h3>
            <button
              @click="$emit('close-form')"
              class="btn-close"
              type="button"
              aria-label="Close form"
            >
              ✕
            </button>
          </div>

          <div class="modal-body">
            <form @submit.prevent="submitForm">
              <div
                v-for="column in schemaColumns"
                :key="column.column_name"
                class="form-group"
              >
                <label :for="`field-${column.column_name}`">
                  {{ column.column_name }}
                  <span v-if="column.is_key" class="badge-key">Key</span>
                </label>
                <input
                  :id="`field-${column.column_name}`"
                  v-model="formData[column.column_name]"
                  :type="getInputType(column.column_type)"
                  :placeholder="`Enter ${column.column_name}`"
                  :disabled="column.is_key && isEditing"
                  class="form-input"
                />
              </div>

              <div class="modal-footer">
                <button
                  type="button"
                  @click="$emit('close-form')"
                  class="btn btn-secondary"
                >
                  Cancel
                </button>
                <button type="submit" class="btn btn-primary">
                  {{ isEditing ? "Update" : "Add" }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Column {
  column_name: string;
  column_type: string;
  is_key: boolean;
}

const props = defineProps<{
  schemaColumns: Column[];
  formData: Record<string, any>;
  show: boolean;
  editingItem?: Record<string, any> | null;
}>();

const emits = defineEmits<{
  "save-item": [data: Record<string, any>];
  "close-form": [];
}>();

const isEditing = computed(() => !!props.editingItem);

const getInputType = (columnType: string): string => {
  const type = columnType.toLowerCase();
  if (type.includes("int") || type.includes("number")) return "number";
  if (type.includes("date")) return "date";
  if (type.includes("email")) return "email";
  return "text";
};

const submitForm = () => {
  emits("save-item", props.formData);
};
</script>

<style scoped>
@import "./InventoryContent.styles.css";

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.btn-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #6b7280;
  transition: color 0.2s;
}

.btn-close:hover {
  color: #333;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.badge-key {
  display: inline-block;
  background-color: #e8f4f8;
  color: #0099cc;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;
}

.form-input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.1);
}

.form-input:disabled {
  background-color: #f5f5f5;
  color: #6b7280;
  cursor: not-allowed;
}

.modal-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid #e0e0e0;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background-color: #3498db;
  color: white;
}

.btn-primary:hover {
  background-color: #2980b9;
}

.btn-secondary {
  background-color: #e8e8e8;
  color: #333;
}

.btn-secondary:hover {
  background-color: #ddd;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .modal-overlay {
    padding: 10px;
  }

  .modal-content {
    max-width: 100%;
  }

  .modal-footer {
    flex-direction: column-reverse;
  }

  .btn {
    width: 100%;
  }
}
</style>
