<template>
  <div class="category-view">
    <div class="category-grid">
      <div
        v-for="category in categoryList"
        :key="category.key"
        class="category-section"
      >
        <div class="category-header">
          <div class="category-title">
            <h3>{{ category.label?.title || category.display }}</h3>
            <p>{{ category.label?.detail || category.display }}</p>
          </div>
          <span class="count-badge">{{ category.items.length }}</span>
        </div>

        <div v-if="category.items.length" class="category-items">
          <div
            v-for="item in category.items"
            :key="itemKey(category.key, item)"
            class="category-item"
          >
            <span
              v-if="item.method"
              class="method-badge"
              :class="`method-${item.method}`"
            >
              {{ item.method }}
            </span>
            <code class="endpoint-path">{{ item.endpoint }}</code>
            <span v-if="category.key === 'frontend_only'" class="components-chip">
              {{ (item as any).components?.length || 0 }} caller(s)
            </span>
            <span v-if="item.is_dynamic" class="dynamic-chip">Dynamic</span>
          </div>
        </div>

        <div v-else class="empty-category">No items for this category.</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface RouteSummary {
  endpoint: string;
  method?: string;
  is_dynamic?: boolean;
  category?: string;
  components?: string[];
}

interface RouteCategoriesUi {
  connected: RouteSummary[];
  frontendOnly: RouteSummary[];
  backendOnly: RouteSummary[];
  unused: RouteSummary[];
}

interface RouteCategoryLabels {
  [key: string]: { title: string; detail: string };
}

interface Props {
  categories: RouteCategoriesUi;
  labels: RouteCategoryLabels;
  searchQuery: string;
}

const props = defineProps<Props>();

const filterItems = (items: RouteSummary[]): RouteSummary[] => {
  const query = props.searchQuery.toLowerCase();
  if (!query) return items;

  return items.filter((item) => {
    const endpointMatch = item.endpoint.toLowerCase().includes(query);
    const componentMatch = (item.components || []).some((c) =>
      c.toLowerCase().includes(query)
    );
    return endpointMatch || componentMatch;
  });
};

const categoryList = computed(() => {
  const { categories, labels } = props;

  return [
    {
      key: 'connected',
      display: 'Connected',
      label: labels?.connected,
      items: filterItems(categories.connected || []),
    },
    {
      key: 'frontend_only',
      display: 'Frontend Missing Backend',
      label: labels?.frontend_only,
      items: filterItems(categories.frontendOnly || []),
    },
    {
      key: 'backend_only',
      display: 'Backend-only',
      label: labels?.backend_only,
      items: filterItems(categories.backendOnly || []),
    },
    {
      key: 'unused',
      display: 'Unused',
      label: labels?.unused,
      items: filterItems(categories.unused || []),
    },
  ];
});

const itemKey = (category: string, item: RouteSummary): string => {
  return `${category}-${item.method || 'GET'}-${item.endpoint}`;
};
</script>

<style scoped>
.category-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.category-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
}

.category-section {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
}

.category-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid #f0f0f0;
  padding-bottom: 10px;
}

.category-title h3 {
  margin: 0 0 4px 0;
  color: #1f2937;
  font-size: 1.05em;
}

.category-title p {
  margin: 0;
  color: #6b7280;
  font-size: 0.9em;
}

.count-badge {
  background: #111827;
  color: white;
  padding: 6px 10px;
  border-radius: 14px;
  font-size: 0.85em;
  min-width: 32px;
  text-align: center;
}

.category-items {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.category-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #eef2f7;
}

.method-badge {
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.85em;
  color: white;
  font-weight: 600;
}

.method-GET {
  background: #2563eb;
}

.method-POST {
  background: #16a34a;
}

.method-PUT {
  background: #f59e0b;
}

.method-DELETE {
  background: #dc2626;
}

.endpoint-path {
  flex: 1;
  font-family: monospace;
  background: white;
  padding: 6px 8px;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  font-size: 0.9em;
}

.components-chip,
.dynamic-chip {
  background: #e0f2fe;
  color: #0b4f6c;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.85em;
  font-weight: 600;
}

.dynamic-chip {
  background: #ede9fe;
  color: #5b21b6;
}

.empty-category {
  text-align: center;
  color: #9ca3af;
  padding: 20px 10px;
  border: 1px dashed #e5e7eb;
  border-radius: 8px;
  background: #f9fafb;
}
</style>
