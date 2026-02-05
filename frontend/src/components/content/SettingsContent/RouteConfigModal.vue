<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="modal-overlay"
      @click.self="$emit('close')"
      role="dialog"
      aria-modal="true"
      aria-labelledby="route-config-title"
    >
      <div class="modal-content">
        <div class="modal-header">
          <h3 id="route-config-title">
            <Icon
              :name="isEdit ? 'edit' : 'plus'"
              size="sm"
              class="inline-block"
            />
            {{ isEdit ? "Edit Route" : "Add New Route" }}
          </h3>
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
            <label for="route-key">Route Key *</label>
            <input
              id="route-key"
              v-model="form.route_key"
              type="text"
              placeholder="e.g., update_stock"
              :disabled="isEdit"
              class="form-input"
            />
            <span class="form-hint">Unique identifier (no spaces)</span>
          </div>

          <div class="form-group">
            <label for="route-name">Route Name *</label>
            <input
              id="route-name"
              v-model="form.route_name"
              type="text"
              placeholder="e.g., Update Stock"
              class="form-input"
            />
          </div>

          <div class="form-group">
            <label for="route-desc">Description</label>
            <input
              id="route-desc"
              v-model="form.description"
              type="text"
              placeholder="e.g., Update stock to all platforms"
              class="form-input"
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="exec-mode">Execution Mode *</label>
              <select
                id="exec-mode"
                v-model="form.execution_mode"
                class="form-select"
              >
                <option value="queue">Queue</option>
                <option value="direct">Direct</option>
              </select>
            </div>

            <div class="form-group">
              <label for="priority">Priority</label>
              <select
                id="priority"
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
              <label for="icon">Icon</label>
              <input
                id="icon"
                v-model="form.icon"
                type="text"
                placeholder=""
                class="form-input form-input-small"
              />
            </div>

            <div class="form-group">
              <label for="category">Category</label>
              <select id="category" v-model="form.category" class="form-select">
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
          <button
            @click="$emit('close')"
            class="btn btn-secondary"
            type="button"
          >
            Cancel
          </button>
          <button
            @click="handleSave"
            :disabled="!isValid || isSaving"
            class="btn btn-primary"
            type="button"
          >
            {{ isSaving ? "Saving..." : isEdit ? "Update" : "Create" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import type { RouteExecutionConfig } from "@/types/routeExecutionConfig";
import Icon from "@/components/ui/Icon.vue";

const props = defineProps<{
  show: boolean;
  config: RouteExecutionConfig | null;
}>();

const emit = defineEmits<{
  close: [];
  save: [data: RouteExecutionConfig];
}>();

const isSaving = ref(false);

const form = ref({
  route_key: "",
  route_name: "",
  description: "",
  execution_mode: "queue" as "queue" | "direct",
  priority: "normal" as "low" | "normal" | "high",
  icon: "",
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
        icon: newConfig.icon || "",
        category: newConfig.category || "general",
      };
    } else {
      form.value = {
        route_key: "",
        route_name: "",
        description: "",
        execution_mode: "queue",
        priority: "normal",
        icon: "",
        category: "general",
      };
    }
  },
  { immediate: true },
);

function handleSave() {
  if (!isValid.value) return;
  isSaving.value = true;
  emit("save", { ...form.value } as RouteExecutionConfig);
  setTimeout(() => {
    isSaving.value = false;
  }, 500);
}

// ESC key handler
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.show) {
    emit("close");
  }
};

// Body scroll lock
watch(
  () => props.show,
  (isOpen) => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
  },
);

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});
</script>

<style src="./RouteConfigModal.styles.css" scoped></style>
