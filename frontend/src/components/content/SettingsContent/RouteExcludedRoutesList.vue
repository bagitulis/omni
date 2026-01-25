<template>
  <div class="excluded-routes-manager">
    <!-- Add New Route Section -->
    <div class="add-route-section">
      <h4>➕ Add Excluded Route</h4>
      <div class="add-route-form">
        <input
          v-model="newRoute"
          type="text"
          placeholder="e.g., /api/inventory"
          class="route-input"
          @keyup.enter="addRoute"
        />
        <button @click="addRoute" class="btn btn-add" :disabled="!newRoute.trim()">
          Add Route
        </button>
      </div>
      <p class="form-hint">Enter full API path (e.g., /api/inventory, /api/orders)</p>
    </div>

    <!-- Routes List Section -->
    <div class="routes-list-section">
      <h4>🚫 Currently Excluded Routes</h4>
      
      <div v-if="excludedRoutesList.length === 0" class="empty-state">
        <p>No routes excluded from cache. All routes will be cached.</p>
      </div>

      <div v-else class="routes-table">
        <div class="routes-header">
          <span class="route-col">Route Path</span>
          <span class="action-col">Action</span>
        </div>
        
        <div
          v-for="(route, idx) in excludedRoutesList"
          :key="idx"
          class="route-row"
        >
          <span class="route-col">
            <code>{{ route }}</code>
          </span>
          <button
            @click="removeRoute(idx)"
            class="btn btn-remove"
            title="Remove from excluded list"
          >
            🗑️ Remove
          </button>
        </div>
      </div>

      <div v-if="excludedRoutesList.length > 0" class="routes-footer">
        <p>{{ excludedRoutesList.length }} route(s) excluded from caching</p>
        <button @click="clearAllExcluded" class="btn btn-secondary-small">
          Clear All
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

interface Props {
  excludeRoutes?: string[];
}

interface Emits {
  (e: 'update', routes: string[]): void;
}

const props = withDefaults(defineProps<Props>(), {
  excludeRoutes: () => [],
});

const emit = defineEmits<Emits>();

const newRoute = ref('');

const excludedRoutesList = computed(() => {
  return (props.excludeRoutes || []).filter(r => r.trim().length > 0);
});

const addRoute = () => {
  const route = newRoute.value.trim();
  if (!route) return;
  
  // Prevent duplicates
  if (excludedRoutesList.value.includes(route)) {
    alert(`Route "${route}" is already excluded.`);
    return;
  }
  
  const updated = [...excludedRoutesList.value, route];
  emit('update', updated);
  newRoute.value = '';
};

const removeRoute = (idx: number) => {
  const updated = excludedRoutesList.value.filter((_, i) => i !== idx);
  emit('update', updated);
};

const clearAllExcluded = () => {
  if (confirm('Clear all excluded routes?')) {
    emit('update', []);
  }
};
</script>

<style scoped lang="css">
.excluded-routes-manager {
  padding: 20px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
}

.add-route-section {
  margin-bottom: 28px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0f0f0;
}

.add-route-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.add-route-form {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}

.route-input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  font-family: 'Courier New', monospace;
  transition: border-color 0.2s;
}

.route-input:focus {
  outline: none;
  border-color: #2196f3;
  box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.1);
}

.btn-add {
  padding: 10px 16px;
  background: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.btn-add:hover:not(:disabled) {
  background: #388e3c;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.btn-add:disabled {
  background: #ccc;
  cursor: not-allowed;
  opacity: 0.6;
}

.form-hint {
  font-size: 12px;
  color: #666;
  margin: 8px 0 0 0;
  line-height: 1.4;
}

.routes-list-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.empty-state {
  padding: 24px;
  text-align: center;
  background: #f9f9f9;
  border-radius: 4px;
  color: #6b7280;
  font-size: 13px;
}

.routes-table {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
}

.routes-header {
  display: grid;
  grid-template-columns: 1fr 120px;
  gap: 12px;
  padding: 12px;
  background: #f5f5f5;
  font-weight: 600;
  font-size: 12px;
  color: #666;
  border-bottom: 1px solid #e0e0e0;
}

.route-col {
  display: flex;
  align-items: center;
}

.action-col {
  text-align: right;
}

.route-row {
  display: grid;
  grid-template-columns: 1fr 120px;
  gap: 12px;
  padding: 12px;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
  transition: background 0.2s;
}

.route-row:hover {
  background: #f9f9f9;
}

.route-row:last-child {
  border-bottom: none;
}

.route-row code {
  display: block;
  padding: 6px 8px;
  background: #f5f5f5;
  border-radius: 3px;
  font-size: 12px;
  color: #d32f2f;
  word-break: break-all;
}

.btn-remove {
  padding: 6px 10px;
  background: #f44336;
  color: white;
  border: none;
  border-radius: 3px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-remove:hover {
  background: #d32f2f;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.routes-footer {
  padding: 12px;
  background: #f9f9f9;
  border-top: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #666;
}

.btn-secondary-small {
  padding: 6px 12px;
  background: #757575;
  color: white;
  border: none;
  border-radius: 3px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary-small:hover {
  background: #616161;
}

@media (max-width: 600px) {
  .add-route-form {
    flex-direction: column;
  }

  .btn-add {
    width: 100%;
  }

  .routes-header,
  .route-row {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .action-col {
    text-align: left;
  }

  .routes-footer {
    flex-direction: column;
    gap: 8px;
  }

  .btn-secondary-small {
    width: 100%;
  }
}
</style>
