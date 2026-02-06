<template>
  <div class="right-sidebar" :class="{ collapsed: collapsed }">
    <!-- Toggle Button -->
    <button
      class="sidebar-toggle"
      @click="$emit('toggle')"
      :title="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      :aria-expanded="!collapsed"
      type="button"
    >
      <span v-if="collapsed" aria-hidden="true">◀</span>
      <span v-else aria-hidden="true">▶</span>
    </button>

    <!-- Expanded View -->
    <div class="sidebar-content" v-if="!collapsed">
      <TokenStatusSection
        :status="status"
        :loading="loading"
        @load-status="$emit('load-status')"
      />
    </div>

    <!-- Collapsed View -->
    <CollapsedSidebar
      v-else
      @toggle="$emit('toggle')"
      @load-status="$emit('load-status')"
      @show-token-modal="(p) => $emit('show-token-modal', p)"
    />
  </div>
</template>

<script lang="ts">
import TokenStatusSection from "./TokenStatusSection.vue";
import CollapsedSidebar from "./CollapsedSidebar.vue";

export default {
  name: "RightSidebar",
  components: { TokenStatusSection, CollapsedSidebar },
  props: {
    collapsed: { type: Boolean, default: false },
    status: { type: Object, default: null },
    loading: { type: Boolean, default: false },
  },
  emits: ["toggle", "show-token-modal", "load-status"],
};
</script>

<style scoped>
.right-sidebar {
  width: 220px;
  background: #ffffff;
  color: #1f2937;
  display: flex;
  flex-direction: column;
  position: relative;
  border-left: 1px solid var(--color-border, #e5e7eb);
  transition: all var(--transition-smooth);
  z-index: var(--z-sidebar, 500);
  height: calc(100vh - 48px);
  box-shadow: var(--shadow-sm);
  min-height: 0;
}

.dark .right-sidebar {
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
  border-left-color: var(--color-border);
}

.right-sidebar.collapsed {
  width: 56px;
}

.sidebar-toggle {
  position: absolute;
  top: 12px;
  left: -14px;
  width: 32px;
  height: 32px;
  background: #ffffff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #00d9ff;
  font-size: 1rem;
  font-weight: 700;
  box-shadow:
    0 4px 12px rgba(0, 217, 255, 0.4),
    inset 0 0 0 2px #00d9ff;
  transition: all 0.3s ease;
  z-index: 101;
  border: 2px solid #00d9ff;
}

.sidebar-toggle:hover {
  background: #f0f9ff;
  transform: scale(1.1);
  box-shadow:
    0 6px 20px rgba(0, 217, 255, 0.8),
    inset 0 0 0 2px #00d9ff;
}

.sidebar-toggle:active {
  transform: scale(0.95);
}

.sidebar-toggle span {
  color: #00d9ff;
  font-weight: 700;
  font-size: 1.1rem;
}

.sidebar-content {
  padding: 12px;
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
}

.sidebar-content::-webkit-scrollbar {
  width: 5px;
}

.sidebar-content::-webkit-scrollbar-track {
  background: hsl(var(--b2));
}

.sidebar-content::-webkit-scrollbar-thumb {
  background: hsl(var(--b3));
  border-radius: 2px;
}

.sidebar-content::-webkit-scrollbar-thumb:hover {
  background: hsl(var(--p) / 0.5);
}

@media (max-width: 1440px) {
  .right-sidebar {
    width: 220px;
  }

  .right-sidebar.collapsed {
    width: 56px;
  }
}

@media (max-width: 1024px) {
  .right-sidebar {
    width: 200px;
  }

  .right-sidebar.collapsed {
    width: 56px;
  }
}

@media (max-width: 768px) {
  .right-sidebar:not(.collapsed) {
    position: fixed;
    right: 0;
    top: 48px;
    height: calc(100vh - 48px);
    z-index: 999;
    width: 280px;
    box-shadow: var(--shadow-xl);
  }

  .right-sidebar.collapsed {
    position: relative;
    width: 56px;
    height: auto;
  }
}

@media (max-width: 480px) {
  .right-sidebar:not(.collapsed) {
    width: 100%;
    right: 0;
    top: 48px;
    height: calc(100vh - 48px);
    z-index: 999;
    border-left: none;
    border-top: 1px solid var(--color-border);
  }

  .right-sidebar.collapsed {
    display: none;
  }
}
</style>
