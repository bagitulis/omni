<template>
  <div class="modal-overlay">
    <div class="modal-content modal-large">
      <div class="modal-header">
        <h4>{{ route?.id ? "Edit Route" : "Add New Route" }}</h4>
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
            <label>Route Path *</label>
            <input
              v-model="formData.route_path"
              type="text"
              placeholder="/api/orders"
              class="form-input"
              :readonly="!!route?.id"
            />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Method</label>
              <select v-model="formData.route_method" class="form-select">
                <option>GET</option>
                <option>POST</option>
                <option>PUT</option>
                <option>DELETE</option>
                <option>PATCH</option>
              </select>
            </div>
            <div class="form-group">
              <label>Category</label>
              <input
                v-model="formData.category"
                type="text"
                placeholder="orders, products, etc"
                class="form-input"
              />
            </div>
          </div>
          <div class="form-group">
            <label>Description</label>
            <input
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
          <label
            ><input v-model="formData.caching_enabled" type="checkbox" />
            Enable</label
          >
          <div v-if="formData.caching_enabled" class="form-row">
            <div class="form-group">
              <label>TTL (seconds)</label>
              <input
                v-model.number="formData.cache_ttl"
                type="number"
                min="0"
                max="3600"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Strategy</label>
              <select v-model="formData.cache_strategy" class="form-select">
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
          <label
            ><input v-model="formData.queue_enabled" type="checkbox" />
            Enable</label
          >
          <div v-if="formData.queue_enabled" class="form-row">
            <div class="form-group">
              <label>Max Size</label>
              <input
                v-model.number="formData.queue_max_size"
                type="number"
                min="1"
                max="1000"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Max Concurrent</label>
              <input
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
          <label
            ><input v-model="formData.rate_limit_enabled" type="checkbox" />
            Enable</label
          >
          <div v-if="formData.rate_limit_enabled" class="form-row">
            <div class="form-group">
              <label>Window (sec)</label>
              <input
                v-model.number="formData.rate_limit_window"
                type="number"
                min="1"
                max="3600"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Max Requests</label>
              <input
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
</template>

<script setup lang="ts">
import { reactive } from "vue";
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
</script>

<style src="./RouteManagementModal.styles.css" scoped></style>
