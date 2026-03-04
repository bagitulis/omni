<template>
  <div class="order-tabs">
    <button
      v-for="tab in orderTabs"
      :key="tab.value"
      :class="['tab-button', { active: activeTab === tab.value }]"
      @click="emit('tab-changed', tab.value)"
    >
      <span class="tab-label">{{ tab.label }}</span>
      <span class="tab-count">{{ tabCounts[tab.value] || 0 }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
interface TabCount {
  [key: string]: number;
}

interface OrderTab {
  label: string;
  value: string;
}

defineProps<{
  activeTab: string;
  tabCounts: TabCount;
  orderTabs: OrderTab[];
}>();

const emit = defineEmits<{
  'tab-changed': [tabValue: string];
}>();
</script>

<style scoped>
@import './OrderManager.styles.css';

.order-tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 0;
  flex-wrap: wrap;
}

.tab-button {
  padding: 11px 20px;
  border: 2px solid #e2e8f0;
  background: white;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  color: #4a5568;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}

.tab-button:hover {
  border-color: #3498db;
  color: #3498db;
  background: #f0f6ff;
}

.tab-button.active {
  background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
  color: white;
  border-color: #2980b9;
  box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
}

.tab-label {
  font-weight: 600;
}

.tab-count {
  background: rgba(255, 255, 255, 0.2);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
  min-width: 28px;
  text-align: center;
}

.tab-button.active .tab-count {
  background: rgba(255, 255, 255, 0.3);
}
</style>
