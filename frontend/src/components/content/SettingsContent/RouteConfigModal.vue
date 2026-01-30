<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ isEdit ? "✎ Edit Route" : "➕ Add New Route" }}</h3>
          <button
            @click="$emit('close')"
            class="btn-close"
            type="button"
            aria-label="Close modal"
          >
            ✕
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>Route Key *</label>
            <input
              v-model="form.route_key"
              type="text"
              placeholder="e.g., update_stock"
              :disabled="isEdit"
              class="form-input"
            />
            <span class="form-hint">Unique identifier (no spaces)</span>
          </div>

          <div class="form-group">
            <label>Route Name *</label>
            <input
              v-model="form.route_name"
              type="text"
              placeholder="e.g., Update Stock"
              class="form-input"
            />
          </div>

          <div class="form-group">
            <label>Description</label>
            <input
              v-model="form.description"
              type="text"
              placeholder="e.g., Update stok ke semua platform"
              class="form-input"
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Execution Mode *</label>
              <select v-model="form.execution_mode" class="form-select">
                <option value="queue">Queue</option>
                <option value="direct">Direct</option>
              </select>
            </div>

            <div class="form-group">
              <label>Priority</label>
              <select
                v-model="form.priority"
                class="form-select"
                :disabled="form.execution_mode !== 'queue'"
              >
                <option value="high">High</option>
                <option value="normal">Normal</option>
                <option value="low">Low</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Icon</label>
              <input
                v-model="form.icon"
                type="text"
                placeholder="📋"
                class="form-input form-input-small"
              />
            </div>

            <div class="form-group">
              <label>Category</label>
              <select v-model="form.category" class="form-select">
                <option value="inventory">Inventory</option>
                <option value="orders">Orders</option>
                <option value="auth">Auth</option>
                <option value="sync">Sync</option>
                <option value="general">General</option>
              </select>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="$emit('close')" class="btn btn-secondary">
            Cancel
          </button>
          <button
            @click="handleSave"
            :disabled="!isValid || isSaving"
            class="btn btn-primary"
          >
            {{ isSaving ? "Saving..." : isEdit ? "Update" : "Create" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import type { RouteExecutionConfig } from "@/types/routeExecutionConfig";

const props = defineProps<{
  show: boolean;
  config: RouteExecutionConfig | null;
}>();

const emit = defineEmits<{
  close: [];
  save: [data: any];
}>();

const isSaving = ref(false);

const form = ref({
  route_key: "",
  route_name: "",
  description: "",
  execution_mode: "queue" as "queue" | "direct",
  priority: "normal" as "low" | "normal" | "high",
  icon: "📋",
  category: "general",
});

const isEdit = computed(() => !!props.config);

const isValid = computed(() => {
  return form.value.route_key.trim() && form.value.route_name.trim();
});

watch(
  () => props.config,
  (newConfig) => {
    if (newConfig) {
      form.value = {
        route_key: newConfig.route_key,
        route_name: newConfig.route_name,
        description: newConfig.description || "",
        execution_mode: newConfig.execution_mode,
        priority: newConfig.priority,
        icon: newConfig.icon || "📋",
        category: newConfig.category || "general",
      };
    } else {
      form.value = {
        route_key: "",
        route_name: "",
        description: "",
        execution_mode: "queue",
        priority: "normal",
        icon: "📋",
        category: "general",
      };
    }
  },
  { immediate: true },
);

function handleSave() {
  if (!isValid.value) return;
  isSaving.value = true;
  emit("save", { ...form.value });
  setTimeout(() => {
    isSaving.value = false;
  }, 500);
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
}

.btn-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #6b7280;
}

.btn-close:hover {
  color: #333;
}

.modal-body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.form-group label {
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.form-input,
.form-select {
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.form-input:disabled {
  background: #f5f5f5;
  color: #6b7280;
}

.form-input-small {
  width: 80px;
}

.form-hint {
  font-size: 11px;
  color: #6b7280;
}

.form-row {
  display: flex;
  gap: 16px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e0e0e0;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary {
  background: #f0f0f0;
  color: #666;
}

.btn-secondary:hover {
  background: #e0e0e0;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover {
  background: #2980b9;
}

.btn-primary:disabled {
  background: #bdc3c7;
  cursor: not-allowed;
}
</style>
