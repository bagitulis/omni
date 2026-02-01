<template>
  <div class="order-tabs">
    <button
      v-for="tab in orderTabs"
      :key="tab.value"
      :class="['tab-item', { active: activeTab === tab.value }]"
      @click="emit('tab-changed', tab.value)"
    >
      <span class="tab-label">{{ tab.label }}</span>
      <span class="tab-count">({{ tabCounts[tab.value] || 0 }})</span>
    </button>
    <div class="tab-indicator" :style="indicatorStyle"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from "vue";

interface TabCount {
  [key: string]: number;
}

interface OrderTab {
  label: string;
  value: string;
}

const props = defineProps<{
  activeTab: string;
  tabCounts: TabCount;
  orderTabs: OrderTab[];
}>();

const emit = defineEmits<{
  "tab-changed": [tabValue: string];
}>();

const indicatorStyle = ref({ left: "0px", width: "0px" });

const updateIndicator = () => {
  nextTick(() => {
    const activeIndex = props.orderTabs.findIndex(
      (t) => t.value === props.activeTab,
    );
    const buttons = document.querySelectorAll(".order-tabs .tab-item");
    if (buttons[activeIndex]) {
      const btn = buttons[activeIndex] as HTMLElement;
      indicatorStyle.value = {
        left: `${btn.offsetLeft}px`,
        width: `${btn.offsetWidth}px`,
      };
    }
  });
};

watch(() => props.activeTab, updateIndicator);
onMounted(updateIndicator);
</script>

<style scoped>
@import "./OrderManager.theme.css";

.order-tabs {
  display: flex;
  gap: 0;
  position: relative;
  border-bottom: 2px solid var(--om-border);
  background: var(--om-bg-primary);
  overflow-x: auto;
  scrollbar-width: none;
}

.order-tabs::-webkit-scrollbar {
  display: none;
}

.tab-item {
  position: relative;
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-secondary);
  transition: all var(--om-transition-fast);
  display: flex;
  align-items: center;
  gap: var(--om-spacing-xs);
  white-space: nowrap;
  flex-shrink: 0;
}

.tab-item:hover {
  color: #ee4d2d;
  background: rgba(238, 77, 45, 0.08);
}

.tab-item.active {
  color: #ee4d2d;
  font-weight: 600;
}

.tab-label {
  font-weight: inherit;
}

.tab-count {
  font-size: 0.75rem;
  margin-left: 4px;
}

.tab-indicator {
  position: absolute;
  bottom: -2px;
  height: 2px;
  background: #ee4d2d;
  border-radius: 2px 2px 0 0;
  transition: all var(--om-transition-normal);
}

/* Responsive */
@media (max-width: 768px) {
  .tab-item {
    padding: var(--om-spacing-sm) var(--om-spacing-md);
    font-size: var(--om-font-xs);
  }
}
</style>
