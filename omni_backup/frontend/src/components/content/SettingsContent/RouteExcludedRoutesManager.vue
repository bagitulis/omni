<template>
  <div class="excluded-routes-manager">
    <!-- Header -->
    <div class="manager-header">
      <h3>🔒 Exclusion List Manager</h3>
      <p class="header-desc">Routes listed here will NOT be cached</p>
    </div>

    <!-- Add Route Section -->
    <div class="add-route-card">
      <label class="label">Add New Route</label>
      <div class="add-route-input">
        <input
          v-model="newRoute"
          type="text"
          placeholder="e.g., /api/inventory"
          class="route-input"
          @keyup.enter="addRoute"
        />
        <button
          @click="addRoute"
          class="btn btn-add"
          :disabled="!newRoute.trim()"
        >
          Add Route
        </button>
      </div>
      <small class="input-hint"
        >Enter the route path to exclude from caching</small
      >
    </div>

    <!-- Excluded Routes List -->
    <div class="excluded-routes-card">
      <div class="routes-header">
        <label class="label"
          >Excluded Routes ({{ excludedRoutes.length }})</label
        >
      </div>

      <div v-if="excludedRoutes.length === 0" class="empty-state">
        <p>No routes excluded. All routes will be cached by default.</p>
      </div>

      <div v-else class="routes-grid">
        <div
          v-for="(route, index) in excludedRoutes"
          :key="index"
          class="route-item"
        >
          <div class="route-content">
            <span class="route-icon">📍</span>
            <span class="route-name">{{ route }}</span>
          </div>
          <button
            @click="removeRoute(index as number)"
            class="btn-icon btn-remove"
            title="Remove"
            aria-label="Remove route"
          >
            ×
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

const props = defineProps<{
  cacheManager: any;
}>();

const emit = defineEmits<{
  "update-exclude-routes": [routes: string[]];
}>();

const newRoute = ref("");

const excludedRoutes = computed(() => {
  return (
    (props.cacheManager.cacheConfig.value as any).excludeRoutes || []
  ).filter((r: string) => r.trim().length > 0);
});

const addRoute = () => {
  const route = newRoute.value.trim();
  if (!route) return;

  // Check if already exists
  if (excludedRoutes.value.includes(route)) {
    alert(`"${route}" is already in the exclusion list`);
    return;
  }

  const updated = [...excludedRoutes.value, route];
  emit("update-exclude-routes", updated);
  newRoute.value = "";
};

const removeRoute = (index: number) => {
  const updated = excludedRoutes.value.filter(
    (_: string, i: number) => i !== index,
  );
  emit("update-exclude-routes", updated);
};
</script>

<style scoped lang="css">
.excluded-routes-manager {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.manager-header {
  padding-bottom: 16px;
  border-bottom: 2px solid #f0f0f0;
}

.manager-header h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 700;
  color: #1a1a1a;
}

.header-desc {
  margin: 0;
  font-size: 13px;
  color: #666;
}

.label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.add-route-card {
  padding: 16px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
}

.add-route-input {
  display: flex;
  gap: 10px;
  margin-bottom: 8px;
}

.route-input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  font-family: "Monaco", "Courier New", monospace;
  background: white;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.route-input:focus {
  outline: none;
  border-color: #2196f3;
  box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.08);
}

.input-hint {
  display: block;
  font-size: 12px;
  color: #6b7280;
  margin: 0;
}

.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-add {
  background: #4caf50;
  color: white;
}

.btn-add:hover:not(:disabled) {
  background: #388e3c;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.excluded-routes-card {
  padding: 16px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
}

.routes-header {
  margin-bottom: 12px;
}

.empty-state {
  padding: 24px;
  text-align: center;
  background: white;
  border: 1px dashed #ddd;
  border-radius: 4px;
  color: #6b7280;
  font-size: 13px;
}

.empty-state p {
  margin: 0;
}

.routes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
}

.route-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: white;
  border: 1px solid #e0e0e0;
  border-left: 3px solid #2196f3;
  border-radius: 4px;
  transition: all 0.2s;
}

.route-item:hover {
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  border-left-color: #1976d2;
}

.route-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.route-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.route-name {
  font-family: "Monaco", "Courier New", monospace;
  font-size: 12px;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-icon {
  background: none;
  border: none;
  padding: 4px 8px;
  margin-left: 8px;
  font-size: 18px;
  color: #f44336;
  cursor: pointer;
  flex-shrink: 0;
  transition: color 0.2s;
  border-radius: 3px;
}

.btn-remove:hover {
  color: #d32f2f;
  background: rgba(244, 67, 54, 0.05);
}

/* Responsive */
@media (max-width: 768px) {
  .routes-grid {
    grid-template-columns: 1fr;
  }

  .add-route-input {
    flex-direction: column;
  }

  .btn-add {
    width: 100%;
  }
}
</style>
