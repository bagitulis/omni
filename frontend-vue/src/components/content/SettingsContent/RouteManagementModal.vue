<template>
  <Teleport to="body">
    <div
      class="modal-overlay"
      @click.self="$emit('close')"
      role="dialog"
      aria-modal="true"
      aria-labelledby="route-modal-title"
    >
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h4 id="route-modal-title">
            {{ route?.id ? "Edit Route" : "Add New Route" }}
          </h4>
          <button
            @click="$emit('close')"
            class="btn-close"
            type="button"
            aria-label="Close modal"
          >
            ✕
          </button>
        </div>

        <form @submit.prevent="saveForm">
          <!-- Basic Info Section -->
          <div class="form-section">
            <div class="form-group">
              <label for="route-path">Route Path *</label>
              <input
                id="route-path"
                v-model="formData.route_path"
                type="text"
                placeholder="/api/orders"
                class="form-input"
                :readonly="!!route?.id"
              />
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="route-method">Method</label>
                <select
                  id="route-method"
                  v-model="formData.route_method"
                  class="form-select"
                >
                  <option>GET</option>
                  <option>POST</option>
                  <option>PUT</option>
                  <option>DELETE</option>
                  <option>PATCH</option>
                </select>
              </div>
              <div class="form-group">
                <label for="route-category">Category</label>
                <input
                  id="route-category"
                  v-model="formData.category"
                  type="text"
                  placeholder="orders, products, etc"
                  class="form-input"
                />
              </div>
            </div>
            <div class="form-group">
              <label for="route-description">Description</label>
              <input
                id="route-description"
                v-model="formData.description"
                type="text"
                placeholder="Route description"
                class="form-input"
              />
            </div>
          </div>

          <!-- Caching Section -->
          <div class="form-section">
            <h5><Icon name="settings" size="sm" /> Caching</h5>
            <label>
              <input v-model="formData.caching_enabled" type="checkbox" />
              Enable caching
            </label>
            <div v-if="formData.caching_enabled" class="form-row">
              <div class="form-group">
                <label for="cache-ttl">TTL (seconds)</label>
                <input
                  id="cache-ttl"
                  v-model.number="formData.cache_ttl"
                  type="number"
                  min="0"
                  max="3600"
                  class="form-input"
                />
              </div>
              <div class="form-group">
                <label for="cache-strategy">Strategy</label>
                <select
                  id="cache-strategy"
                  v-model="formData.cache_strategy"
                  class="form-select"
                >
                  <option value="standard">Standard</option>
                  <option value="aggressive">Aggressive</option>
                  <option value="minimal">Minimal</option>
                </select>
              </div>
            </div>
          </div>

          <!-- Queue Section -->
          <div class="form-section">
            <h5><Icon name="document" size="sm" /> Queue</h5>
            <label>
              <input v-model="formData.queue_enabled" type="checkbox" />
              Enable queue
            </label>
            <div v-if="formData.queue_enabled" class="form-row">
              <div class="form-group">
                <label for="queue-max-size">Max Size</label>
                <input
                  id="queue-max-size"
                  v-model.number="formData.queue_max_size"
                  type="number"
                  min="1"
                  max="1000"
                  class="form-input"
                />
              </div>
              <div class="form-group">
                <label for="max-concurrent">Max Concurrent</label>
                <input
                  id="max-concurrent"
                  v-model.number="formData.max_concurrent"
                  type="number"
                  min="1"
                  max="50"
                  class="form-input"
                />
              </div>
            </div>
          </div>

          <!-- Rate Limit Section -->
          <div class="form-section">
            <h5><Icon name="clock" size="sm" /> Rate Limit</h5>
            <label>
              <input v-model="formData.rate_limit_enabled" type="checkbox" />
              Enable rate limiting
            </label>
            <div v-if="formData.rate_limit_enabled" class="form-row">
              <div class="form-group">
                <label for="rate-limit-window">Window (sec)</label>
                <input
                  id="rate-limit-window"
                  v-model.number="formData.rate_limit_window"
                  type="number"
                  min="1"
                  max="3600"
                  class="form-input"
                />
              </div>
              <div class="form-group">
                <label for="rate-limit-max">Max Requests</label>
                <input
                  id="rate-limit-max"
                  v-model.number="formData.rate_limit_max"
                  type="number"
                  min="1"
                  max="10000"
                  class="form-input"
                />
              </div>
            </div>
          </div>

          <!-- Form Actions -->
          <div class="modal-footer">
            <button
              type="button"
              @click="$emit('close')"
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="btn btn-primary"
              :disabled="isSaving"
              aria-label="Save route"
            >
              <Icon name="download" size="sm" />
              {{ isSaving ? "Saving..." : "Save" }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { reactive, onMounted, onUnmounted } from "vue";
import Icon from "@/components/ui/Icon.vue";
import type { RouteConfig } from "./types/routeManagement";

interface Props {
  route: RouteConfig | null;
  isSaving: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  save: [route: RouteConfig];
  close: [];
}>();

const formData = reactive<RouteConfig>(
  props.route || {
    route_path: "",
    route_method: "GET",
    enabled: true,
    caching_enabled: true,
    cache_ttl: 0,
    cache_strategy: "standard",
    queue_enabled: true,
    queue_max_size: 100,
    queue_priority: "normal",
    max_concurrent: 5,
    rate_limit_enabled: false,
    rate_limit_window: 60,
    rate_limit_max: 100,
    min_interval_ms: 0,
    timeout: 30000,
    retry_enabled: true,
    max_retries: 3,
    retry_delay_ms: 1000,
  },
);

const saveForm = () => {
  emit("save", formData);
};

// ESC key handler
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape") {
    emit("close");
  }
};

// Body scroll lock
onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
  document.body.style.overflow = "hidden";
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});
</script>

<style src="./RouteManagementModal.styles.css" scoped></style>
