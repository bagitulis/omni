<template>
  <div class="right-sidebar-wrapper">
    <!-- Mobile Overlay Backdrop -->
    <Transition name="fade">
      <div
        v-if="showOverlay"
        class="sidebar-overlay"
        @click="$emit('toggle')"
      />
    </Transition>

    <aside
      class="right-sidebar smooth-all"
      :class="{ collapsed, 'mobile-open': !collapsed }"
    >
      <!-- Toggle Button -->
      <button
        class="sidebar-toggle press-effect"
        @click="$emit('toggle')"
        :title="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      >
        <svg
          class="toggle-icon"
          :class="{ rotated: !collapsed }"
          width="16"
          height="16"
          viewBox="0 0 16 16"
        >
          <path
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            d="M6 3l5 5-5 5"
          />
        </svg>
      </button>

      <!-- Content -->
      <div class="sidebar-content" v-if="!collapsed">
        <TokenStatusSection
          :status="status"
          :loading="loading"
          @load-status="$emit('load-status')"
        />
      </div>

      <!-- Collapsed View -->
      <RightSidebarCollapsedView
        v-else
        @toggle="$emit('toggle')"
        @load-status="$emit('load-status')"
        @show-token-modal="$emit('show-token-modal', $event)"
      />
    </aside>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, ref, onMounted, onUnmounted } from "vue";
import TokenStatusSection from "./TokenStatusSection.vue";
import RightSidebarCollapsedView from "./RightSidebarCollapsedView.vue";

export default defineComponent({
  name: "RightSidebar",
  components: { TokenStatusSection, RightSidebarCollapsedView },
  props: {
    collapsed: { type: Boolean, default: false },
    status: { type: Object, default: null },
    loading: { type: Boolean, default: false },
  },
  emits: ["toggle", "show-token-modal", "load-status"],
  setup(props) {
    const windowWidth = ref(
      typeof window !== "undefined" ? window.innerWidth : 1024
    );
    const handleResize = () => {
      windowWidth.value = window.innerWidth;
    };

    onMounted(() => window.addEventListener("resize", handleResize));
    onUnmounted(() => window.removeEventListener("resize", handleResize));

    const showOverlay = computed(
      () => windowWidth.value < 768 && !props.collapsed
    );

    return { showOverlay };
  },
});
</script>

<style scoped>
.right-sidebar-wrapper {
  position: relative;
  z-index: var(--z-sidebar, 500);
}

/* Mobile Overlay */
.sidebar-overlay {
  display: none;
}
@media (max-width: 767px) {
  .sidebar-overlay {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(2px);
    z-index: calc(var(--z-sidebar, 500) - 1);
  }
}

/* Sidebar Container */
.right-sidebar {
  width: 220px;
  background: var(--color-bg-primary, #ffffff);
  color: var(--color-text-primary, #1f2937);
  display: flex;
  flex-direction: column;
  position: relative;
  border-left: 1px solid var(--color-border, #e5e7eb);
  z-index: var(--z-sidebar, 500);
  height: calc(100vh - 48px);
  box-shadow: var(--shadow-sm);
  min-height: 0;
}
.right-sidebar.collapsed {
  width: 56px;
}

/* Toggle Button */
.sidebar-toggle {
  position: absolute;
  top: 12px;
  left: -14px;
  width: 28px;
  height: 28px;
  background: var(--color-bg-primary, #ffffff);
  border: 2px solid var(--color-accent, #00d9ff);
  border-radius: var(--radius-full, 9999px);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--color-accent, #00d9ff);
  box-shadow: 0 2px 8px rgba(0, 217, 255, 0.3);
  z-index: 10;
}
.sidebar-toggle:hover {
  background: var(--color-primary-50, #eff6ff);
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(0, 217, 255, 0.5);
}
.toggle-icon {
  transition: transform var(--transition-smooth);
}
.toggle-icon.rotated {
  transform: rotate(180deg);
}

/* Content Area */
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
  background: var(--color-bg-secondary);
}
.sidebar-content::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: var(--radius-sm);
}
.sidebar-content::-webkit-scrollbar-thumb:hover {
  background: var(--color-text-muted);
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .right-sidebar {
    width: 200px;
  }
  .right-sidebar.collapsed {
    width: 56px;
  }
}

/* Responsive - Mobile */
@media (max-width: 767px) {
  .right-sidebar {
    position: fixed;
    right: 0;
    top: 48px;
    height: calc(100vh - 48px - 56px - env(safe-area-inset-bottom, 0px));
    width: 300px;
    transform: translateX(100%);
    box-shadow: var(--shadow-xl);
    transition: transform var(--transition-smooth);
  }
  .right-sidebar.mobile-open {
    transform: translateX(0);
  }
  .right-sidebar.collapsed {
    transform: translateX(100%);
    width: 0;
    border: none;
    overflow: hidden;
  }
  .sidebar-toggle {
    left: -44px;
    top: 16px;
    width: 36px;
    height: 36px;
    background: var(--color-primary, #667eea);
    border-color: var(--color-primary, #667eea);
    color: white;
  }
  .sidebar-toggle:hover {
    background: var(--color-primary-dark, #5a67d8);
  }
}

@media (max-width: 480px) {
  .right-sidebar {
    width: 100%;
    max-width: 100vw;
  }
  .sidebar-toggle {
    left: -48px;
  }
}

/* Transition classes */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-base);
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
