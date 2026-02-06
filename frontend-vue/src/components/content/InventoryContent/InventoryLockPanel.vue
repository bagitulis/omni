<template>
  <div class="lock-panel">
    <div class="filter-section">
      <div class="section-header" @click="isExpanded = !isExpanded">
        <h4>🔒 Lock Column Edit</h4>
        <span class="toggle-icon" :class="{ 'is-expanded': isExpanded }"
          >▼</span
        >
      </div>

      <div v-if="isExpanded" class="section-content">
        <div class="lock-info">
          <p>
            Select columns that should NOT be editable. Locked columns will
            display 🔒 saat di-hover.
          </p>
        </div>

        <div v-if="editableColumns.length === 0" class="empty-state">
          <p>No columns available for configuration.</p>
        </div>

        <div v-else class="lock-grid">
          <div
            v-for="column in editableColumns"
            :key="column"
            class="lock-item"
          >
            <label class="checkbox-label">
              <input
                type="checkbox"
                :checked="isLocked(column)"
                @change="wrappedToggleLock(column)"
                class="lock-checkbox"
              />
              <span class="column-name">{{ column }}</span>
              <span v-if="isLocked(column)" class="lock-badge">🔒 Locked</span>
            </label>
          </div>
        </div>

        <div class="lock-actions">
          <button @click="wrappedLockAll" class="btn-lock">Lock All</button>
          <button @click="wrappedUnlockAll" class="btn-unlock">
            Unlock All
          </button>
        </div>

        <div class="locked-summary">
          <p>
            <strong>Locked Columns ({{ lockedCount }}):</strong>
          </p>
          <div v-if="lockedCount > 0" class="locked-list">
            <span v-for="col in getLockedList" :key="col" class="locked-tag">
              {{ col }}
            </span>
          </div>
          <div v-else class="empty-locked">
            <p>All columns are editable</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue";
import { useInventoryLockManager } from "./composables/useInventoryLockManager";

const props = defineProps<{
  schemaColumns: any[];
  columnVisibility: Record<string, boolean>;
}>();

const isExpanded = ref(true);

const {
  editableColumns,
  lockedCount,
  getLockedList,
  isLocked,
  toggleLock,
  lockAll,
  unlockAll,
  lockedColumns,
} = useInventoryLockManager(props.schemaColumns, props.columnVisibility);

// Watch for lock changes
watch(
  lockedColumns,
  () => {
    // Lock state updated
  },
  { deep: true }
);

// Toggle lock wrapper
const wrappedToggleLock = (column: string) => {
  toggleLock(column);
};

// Lock all button wrapper
const wrappedLockAll = () => {
  lockAll();
};

// Unlock all button wrapper
const wrappedUnlockAll = () => {
  unlockAll();
};
</script>

<style scoped>
.lock-panel {
  width: 100%;
}

.filter-section {
  background-color: #fff;
  padding: 0;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  margin-bottom: 15px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  padding: 12px 20px;
  transition: background-color 0.2s;
  border-radius: 8px 8px 0 0;
}

.section-header:hover {
  background-color: #f9f9f9;
}

.section-header h4 {
  margin: 0;
  flex: 1;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}

.toggle-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  font-size: 12px;
  color: #6b7280;
  transition: transform 0.2s;
  transform: rotate(-90deg);
}

.toggle-icon.is-expanded {
  transform: rotate(0deg);
}

.section-content {
  padding: 20px;
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    max-height: 0;
  }
  to {
    opacity: 1;
    max-height: 1000px;
  }
}

.lock-info {
  background-color: #f0f8ff;
  padding: 10px 12px;
  border-radius: 4px;
  border-left: 3px solid #3498db;
  margin-bottom: 15px;
}

.lock-info p {
  margin: 0;
  color: #555;
  font-size: 12px;
  line-height: 1.5;
}

.empty-state {
  text-align: center;
  padding: 20px;
  color: #6b7280;
  font-size: 12px;
}

.lock-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-bottom: 15px;
}

.lock-item {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
  user-select: none;
}

.checkbox-label:hover {
  background-color: #f5f5f5;
}

.lock-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #e74c3c;
}

.column-name {
  font-size: 13px;
  color: #333;
  font-weight: 500;
}

.lock-badge {
  font-size: 11px;
  color: #e74c3c;
  font-weight: 600;
  margin-left: 4px;
}

.lock-actions {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.btn-lock,
.btn-unlock {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background-color: white;
}

.btn-lock {
  border-color: #e74c3c;
  color: #e74c3c;
}

.btn-lock:hover {
  background-color: #ffe0e0;
}

.btn-unlock {
  border-color: #27ae60;
  color: #27ae60;
}

.btn-unlock:hover {
  background-color: #e8f8f5;
}

.locked-summary {
  background-color: #f9f9f9;
  padding: 12px;
  border-radius: 4px;
  border: 1px solid #e8e8e8;
}

.locked-summary p {
  margin: 0 0 8px 0;
  font-size: 12px;
  color: #333;
}

.locked-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.locked-tag {
  background-color: #ffe0e0;
  color: #e74c3c;
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
}

.empty-locked {
  color: #6b7280;
  font-size: 12px;
}

.empty-locked p {
  margin: 0;
}
</style>
