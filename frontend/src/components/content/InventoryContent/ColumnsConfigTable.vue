<template>
  <div class="columns-grid">
    <div
      v-for="(column, index) in sortedColumns"
      :key="column.name"
      class="column-card"
      :class="{
        'is-active': props.isSelected(column.name),
        'is-locked': props.isColumnLocked(column.name),
        'is-dragging': draggedIndex === index,
        'drag-over': dragOverIndex === index,
      }"
      draggable="true"
      @dragstart="handleDragStart(index, $event)"
      @dragover.prevent="handleDragOver(index)"
      @dragleave="handleDragLeave"
      @drop="handleDrop(index)"
      @dragend="handleDragEnd"
    >
      <span class="drag-handle" title="Drag to reorder">⋮⋮</span>

      <label class="card-checkbox">
        <input
          type="checkbox"
          :checked="props.isSelected(column.name)"
          @change="handleVisibilityChange(column.name)"
        />
      </label>

      <span class="card-name">{{ column.name }}</span>

      <button
        v-if="props.isSelected(column.name)"
        class="lock-btn"
        :class="{ locked: props.isColumnLocked(column.name) }"
        @click="
          handleLockChange(column.name, !props.isColumnLocked(column.name))
        "
        :title="props.isColumnLocked(column.name) ? 'Unlock' : 'Lock'"
      >
        {{ props.isColumnLocked(column.name) ? "🔒" : "🔓" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useInventoryConfig } from "./composables/useInventoryConfig";

const props = defineProps<{
  availableColumns: Array<{ name: string; [key: string]: any }>;
  isSelected: (colName: string) => boolean;
  isColumnLocked: (colName: string) => boolean;
  inventoryList?: Record<string, any>[];
}>();

const emits = defineEmits<{
  "toggle-visibility": [columnName: string];
  "columns-reordered": [columns: string[]];
}>();

// Use old config for backward compatibility
const { setColumnLocked, columnOrder, setColumnOrder } = useInventoryConfig();

// Local sorted columns for drag-and-drop
const localColumnOrder = ref<string[]>([]);

// Initialize order from saved config or props
const initializeOrder = () => {
  // Don't initialize if no columns available yet
  if (!props.availableColumns || props.availableColumns.length === 0) {
    return;
  }

  const savedOrder = columnOrder.value;
  if (savedOrder && savedOrder.length > 0) {
    // Use saved order, but include any new columns
    const existingNames = props.availableColumns.map((c) => c.name);
    const orderedNames = savedOrder.filter((name: string) =>
      existingNames.includes(name)
    );
    const newNames = existingNames.filter((name) => !savedOrder.includes(name));
    localColumnOrder.value = [...orderedNames, ...newNames];
  } else {
    localColumnOrder.value = props.availableColumns.map((c) => c.name);
  }
};

// Watch for changes in availableColumns
watch(
  () => props.availableColumns,
  () => initializeOrder(),
  { immediate: true }
);

// Computed sorted columns based on local order
const sortedColumns = computed(() => {
  return localColumnOrder.value
    .map((name) => props.availableColumns.find((c) => c.name === name))
    .filter(Boolean) as Array<{ name: string; [key: string]: any }>;
});

// Drag and drop state
const draggedIndex = ref<number | null>(null);
const dragOverIndex = ref<number | null>(null);

const handleDragStart = (index: number, event: DragEvent) => {
  draggedIndex.value = index;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
  }
};

const handleDragOver = (index: number) => {
  if (draggedIndex.value !== null && draggedIndex.value !== index) {
    dragOverIndex.value = index;
  }
};

const handleDragLeave = () => {
  dragOverIndex.value = null;
};

const handleDrop = (targetIndex: number) => {
  if (draggedIndex.value === null || draggedIndex.value === targetIndex) return;

  const newOrder = [...localColumnOrder.value];
  const [draggedItem] = newOrder.splice(draggedIndex.value, 1);
  newOrder.splice(targetIndex, 0, draggedItem);

  localColumnOrder.value = newOrder;
  setColumnOrder(newOrder);
  emits("columns-reordered", newOrder);

  dragOverIndex.value = null;
};

const handleDragEnd = () => {
  draggedIndex.value = null;
  dragOverIndex.value = null;
};

/**
 * Handle visibility change - emit to parent for coordinated update
 */
const handleVisibilityChange = (columnName: string) => {
  emits("toggle-visibility", columnName);
};

/**
 * Handle lock change - directly call setColumnLocked to update state
 */
const handleLockChange = (columnName: string, checked: boolean) => {
  console.log(
    `%c🔄 [ColumnsConfigTable] Lock changed for "${columnName}" to ${checked}`,
    "color: #9b59b6; font-weight: bold"
  );
  setColumnLocked(columnName, checked);
};
</script>

<style scoped>
.columns-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  width: 100%;
}

.column-card {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  transition: all 0.2s;
  cursor: grab;
}

.column-card:active {
  cursor: grabbing;
}

.column-card:hover {
  background: #f0f0f0;
}

.column-card.is-active {
  background: #e8f4fd;
  border-color: #3498db;
}

.column-card.is-active.is-locked {
  background: #fef3f2;
  border-color: #e74c3c;
}

.column-card.is-dragging {
  opacity: 0.5;
  transform: scale(0.98);
}

.column-card.drag-over {
  border-color: #27ae60;
  border-style: dashed;
  background: #e8f8f5;
}

.drag-handle {
  color: #6b7280;
  font-size: 12px;
  cursor: grab;
  user-select: none;
  flex-shrink: 0;
}

.drag-handle:active {
  cursor: grabbing;
}

.card-checkbox input {
  width: 14px;
  height: 14px;
  cursor: pointer;
  accent-color: #3498db;
  flex-shrink: 0;
}

.card-name {
  flex: 1;
  font-size: 12px;
  font-weight: 500;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lock-btn {
  padding: 2px 4px;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 12px;
  opacity: 0.6;
  transition: opacity 0.2s;
  flex-shrink: 0;
}

.lock-btn:hover {
  opacity: 1;
}

.lock-btn.locked {
  opacity: 1;
}

/* Responsive */
@media (max-width: 1024px) {
  .columns-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .columns-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .columns-grid {
    grid-template-columns: 1fr;
  }
}
</style>
