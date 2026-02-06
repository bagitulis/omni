<template>
  <div class="order-tabs" role="tablist">
    <button
      v-for="tab in orderTabs"
      :key="tab.value"
      :class="['tab-item', { active: activeTab === tab.value }]"
      role="tab"
      :aria-selected="activeTab === tab.value"
      :tabindex="activeTab === tab.value ? 0 : -1"
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
  gap: var(--om-spacing-xs);
  position: relative;
  border-bottom: none;
  background: transparent;
  overflow-x: auto;
  scrollbar-width: none;
  padding: 2px;
}

.order-tabs::-webkit-scrollbar {
  display: none;
}

.tab-item {
  position: relative;
  padding: 0.55rem 1.05rem;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: var(--om-font-sm);
  font-weight: 600;
  color: var(--om-text-secondary);
  transition: all var(--om-transition-fast);
  display: flex;
  align-items: center;
  gap: var(--om-spacing-xs);
  white-space: nowrap;
  flex-shrink: 0;
  border-radius: var(--om-radius-full);
}

.tab-item:hover {
  color: var(--om-primary);
  background: #fff2ee;
}

.tab-item.active {
  color: var(--om-primary);
  background: #fff2ee;
  box-shadow: inset 0 0 0 1px rgba(238, 77, 45, 0.2);
}

.tab-label {
  font-weight: inherit;
}

.tab-count {
  font-size: 0.75rem;
  margin-left: 4px;
  color: inherit;
}

.tab-indicator {
  display: none;
}

/* Responsive */
@media (max-width: 768px) {
  .tab-item {
    padding: var(--om-spacing-sm) var(--om-spacing-md);
    font-size: var(--om-font-xs);
  }
}
</style>
