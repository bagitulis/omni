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
              v-model="formData.routePath"
              type="text"
              placeholder="/api/orders"
              class="form-input"
              :readonly="!!route?.id"
            />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Method</label>
              <select v-model="formData.routeMethod" class="form-select">
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
          <h5>⚙️ Caching</h5>
          <label
            ><input v-model="formData.cachingEnabled" type="checkbox" />
            Enable</label
          >
          <div v-if="formData.cachingEnabled" class="form-row">
            <div class="form-group">
              <label>TTL (seconds)</label>
              <input
                v-model.number="formData.cacheTTL"
                type="number"
                min="0"
                max="3600"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Strategy</label>
              <select v-model="formData.cacheStrategy" class="form-select">
                <option value="standard">Standard</option>
                <option value="aggressive">Aggressive</option>
                <option value="minimal">Minimal</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Queue Section -->
        <div class="form-section">
          <h5>📋 Queue</h5>
          <label
            ><input v-model="formData.queueEnabled" type="checkbox" />
            Enable</label
          >
          <div v-if="formData.queueEnabled" class="form-row">
            <div class="form-group">
              <label>Max Size</label>
              <input
                v-model.number="formData.queueMaxSize"
                type="number"
                min="1"
                max="1000"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Max Concurrent</label>
              <input
                v-model.number="formData.maxConcurrent"
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
          <h5>⏱️ Rate Limit</h5>
          <label
            ><input v-model="formData.rateLimitEnabled" type="checkbox" />
            Enable</label
          >
          <div v-if="formData.rateLimitEnabled" class="form-row">
            <div class="form-group">
              <label>Window (sec)</label>
              <input
                v-model.number="formData.rateLimitWindow"
                type="number"
                min="1"
                max="3600"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Max Requests</label>
              <input
                v-model.number="formData.rateLimitMax"
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
          <button type="submit" class="btn btn-primary" :disabled="isSaving">
            {{ isSaving ? "Saving..." : "💾 Save" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from "vue";
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
    routePath: "",
    routeMethod: "GET",
    enabled: true,
    cachingEnabled: true,
    cacheTTL: 0,
    cacheStrategy: "standard",
    queueEnabled: true,
    queueMaxSize: 100,
    queuePriority: "normal",
    maxConcurrent: 5,
    rateLimitEnabled: false,
    rateLimitWindow: 60,
    rateLimitMax: 100,
    minIntervalMs: 0,
    timeout: 30000,
    retryEnabled: true,
    maxRetries: 3,
    retryDelayMs: 1000,
  }
);

const saveForm = () => {
  emit("save", formData);
};
</script>

<style src="./RouteManagementModal.styles.css" scoped></style>
