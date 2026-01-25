<template>
  <div class="route-table-wrapper">
    <div class="routes-table-container">
      <table class="routes-table" aria-label="API route configurations">
        <thead>
          <tr>
            <th scope="col">
              <input
                type="checkbox"
                :checked="
                  selectedIds.length === routes.length && routes.length > 0
                "
                @change="emitSelectAll"
                aria-label="Select all routes"
              />
            </th>
            <th scope="col">Path</th>
            <th scope="col">Category</th>
            <th scope="col">Method</th>
            <th scope="col">Enabled</th>
            <th scope="col">Cache (TTL)</th>
            <th scope="col">Queue</th>
            <th scope="col">Concurrency</th>
            <th scope="col">Timeout</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="route in routes" :key="route.id" class="route-row">
            <td>
              <input
                type="checkbox"
                :checked="selectedIds.includes(route.id!)"
                @change="emitToggleSelection(route.id!)"
              />
            </td>
            <td class="cell-path">
              <span class="path-text" :title="route.routePath">{{
                route.routePath
              }}</span>
            </td>
            <td>
              <span
                class="badge"
                :class="`badge-${route.category || 'default'}`"
              >
                {{ route.category || "N/A" }}
              </span>
            </td>
            <td>
              <span
                class="method-badge"
                :class="`method-${route.routeMethod.toLowerCase()}`"
              >
                {{ route.routeMethod }}
              </span>
            </td>
            <td>
              <button
                @click="emitToggleEnabled(route.id!, route.enabled)"
                :class="['toggle-btn', route.enabled ? 'enabled' : 'disabled']"
              >
                {{ route.enabled ? "✅" : "❌" }}
              </button>
            </td>
            <td class="cell-cache">
              <span v-if="route.cachingEnabled" class="cache-enabled"
                >{{ route.cacheTTL }}s</span
              >
              <span v-else class="cache-disabled">Disabled</span>
            </td>
            <td>
              <span v-if="route.queueEnabled" class="badge badge-success">
                Max: {{ route.queueMaxSize }}
              </span>
              <span v-else class="badge badge-danger">Disabled</span>
            </td>
            <td class="cell-center">{{ route.maxConcurrent }}</td>
            <td class="cell-center">{{ route.timeout }}ms</td>
            <td class="cell-actions">
              <button
                @click="emitEdit(route)"
                class="btn-icon"
                title="Edit"
                aria-label="Edit route"
              >
                <span aria-hidden="true">✏️</span>
              </button>
              <button
                @click="emitDelete(route.id!)"
                class="btn-icon btn-danger"
                title="Delete"
                aria-label="Delete route"
              >
                <span aria-hidden="true">🗑️</span>
              </button>
            </td>
          </tr>
          <tr v-if="routes.length === 0" class="empty-row">
            <td colspan="10">
              <div class="empty-state">
                <span class="empty-icon">📭</span>
                <p>No routes found. Try adjusting filters.</p>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { RouteConfig } from "./types/routeManagement";

defineProps<{
  routes: RouteConfig[];
  selectedIds: string[];
}>();

const emit = defineEmits<{
  "toggle-selection": [routeId: string];
  "select-all": [];
  "toggle-enabled": [routeId: string, currentState: boolean];
  edit: [route: RouteConfig];
  delete: [routeId: string];
}>();

const emitToggleSelection = (routeId: string) => {
  emit("toggle-selection", routeId);
};

const emitSelectAll = () => {
  emit("select-all");
};

const emitToggleEnabled = (routeId: string, currentState: boolean) => {
  emit("toggle-enabled", routeId, currentState);
};

const emitEdit = (route: RouteConfig) => {
  emit("edit", route);
};

const emitDelete = (routeId: string) => {
  emit("delete", routeId);
};
</script>

<style scoped lang="css">
.routes-table-container {
  overflow-x: auto;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  background: white;
}

.routes-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.routes-table thead {
  background: #f9f9f9;
  border-bottom: 2px solid #e0e0e0;
}

.routes-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #333;
}

.routes-table td {
  padding: 12px;
  border-bottom: 1px solid #e0e0e0;
}

.route-row:hover {
  background: #f5f5f5;
}

.cell-path {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.path-text {
  font-family: monospace;
  font-size: 12px;
}

.badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.badge-orders {
  background: #e3f2fd;
  color: #1976d2;
}
.badge-products {
  background: #f3e5f5;
  color: #7b1fa2;
}
.badge-inventory {
  background: #e8f5e9;
  color: #388e3c;
}
.badge-default {
  background: #f5f5f5;
  color: #666;
}
.badge-success {
  background: #c8e6c9;
  color: #2e7d32;
}
.badge-danger {
  background: #ffcdd2;
  color: #c62828;
}

.method-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  color: white;
}

.method-get {
  background: #4caf50;
}
.method-post {
  background: #2196f3;
}
.method-put {
  background: #ff9800;
}
.method-delete {
  background: #f44336;
}
.method-patch {
  background: #9c27b0;
}

.toggle-btn {
  padding: 4px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  background: white;
  font-size: 12px;
}

.toggle-btn.enabled {
  background: #e8f5e9;
  color: #2e7d32;
  border-color: #4caf50;
}
.toggle-btn.disabled {
  background: #ffebee;
  color: #c62828;
  border-color: #f44336;
}

.cache-enabled {
  color: #2e7d32;
  font-weight: 600;
}
.cache-disabled {
  color: #6b7280;
}

.cell-cache,
.cell-center,
.cell-actions {
  text-align: center;
}

.cell-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.btn-icon {
  padding: 6px 10px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: #f5f5f5;
  border-color: #6b7280;
}
.btn-icon.btn-danger:hover {
  background: #ffebee;
  border-color: #f44336;
}

.empty-row {
  height: 200px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 12px;
  color: #6b7280;
}

.empty-icon {
  font-size: 48px;
  opacity: 0.5;
}
.empty-state p {
  margin: 0;
  font-size: 14px;
}
</style>
