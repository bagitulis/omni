<template>
  <div class="route-table-section">
    <!-- Collapsible Header -->
    <div class="section-header" @click="isExpanded = !isExpanded">
      <div class="header-left">
        <span class="toggle-icon">{{ isExpanded ? '▼' : '▶' }}</span>
        <h3>📋 Routes Table</h3>
        <span class="route-count">({{ routeCount }} routes)</span>
      </div>
      <button @click.stop="onAddRoute" class="btn btn-add-route">
        ➕ Add Route
      </button>
    </div>

    <!-- Expandable Content -->
    <transition name="collapse">
      <div v-if="isExpanded" class="table-wrapper">
        <slot />
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

defineProps<{
  routeCount: number;
}>();

defineEmits<{
  'add-route': [];
}>();

const isExpanded = ref(true);

const onAddRoute = () => {
  const emit = defineEmits<{
    'add-route': [];
  }>();
  emit('add-route');
};
</script>

<style scoped lang="css">
.route-table-section {
  background: white;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  margin-bottom: 16px;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  cursor: pointer;
  transition: background 0.2s;
  user-select: none;
}

.section-header:hover {
  background: #f0f2f5;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.toggle-icon {
  font-size: 12px;
  color: #666;
  transition: transform 0.2s;
  display: inline-block;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
}

.route-count {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-add-route {
  background: #4caf50;
  color: white;
}

.btn-add-route:hover {
  background: #388e3c;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.table-wrapper {
  padding: 16px;
  border-top: 1px solid #e0e0e0;
}

/* Collapse Animation */
.collapse-enter-active,
.collapse-leave-active {
  transition: all 0.3s ease;
}

.collapse-enter-from {
  opacity: 0;
  max-height: 0;
}

.collapse-leave-to {
  opacity: 0;
  max-height: 0;
}

.collapse-enter-to,
.collapse-leave-from {
  opacity: 1;
  max-height: 1000px;
}

/* Responsive */
@media (max-width: 768px) {
  .section-header {
    flex-wrap: wrap;
    gap: 12px;
  }

  .header-left {
    width: 100%;
  }

  .btn-add-route {
    width: 100%;
  }
}
</style>
