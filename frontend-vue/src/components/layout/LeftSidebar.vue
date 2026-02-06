<template>
  <div class="left-sidebar-wrapper">
    <!-- Mobile Overlay Backdrop -->
    <Transition name="fade">
      <div
        v-if="showOverlay"
        class="sidebar-overlay"
        @click="$emit('toggle')"
      />
    </Transition>

    <aside
      class="left-sidebar smooth-all"
      :class="{ collapsed, 'mobile-open': !collapsed }"
    >
      <!-- Toggle Button -->
      <button
        class="sidebar-toggle press-effect"
        @click="$emit('toggle')"
        :title="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        :aria-expanded="!collapsed"
      >
        <svg
          class="toggle-icon"
          :class="{ rotated: collapsed }"
          width="16"
          height="16"
          viewBox="0 0 16 16"
        >
          <path
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            d="M10 3L5 8l5 5"
          />
        </svg>
      </button>

      <!-- Logo / Header -->
      <div class="sidebar-header" v-if="!collapsed">
        <span class="header-icon">🛒</span>
        <span class="app-name">Ecommerce</span>
      </div>

      <!-- Navigation Menu -->
      <SidebarMenu
        :collapsed="collapsed"
        :platforms="platforms"
        :product-platforms="productPlatforms"
        :expanded-sections="expandedSections"
        :active-platform="activePlatform"
        @toggle-expand="toggleExpand"
      />

      <!-- Collapsed View -->
      <CollapsedView v-if="collapsed" />
    </aside>
  </div>
</template>

<script lang="ts">
import {
  defineComponent,
  computed,
  ref,
  onMounted,
  onUnmounted,
  watch,
} from "vue";
import { useRoute } from "vue-router";
import SidebarMenu from "./sidebar/SidebarMenu.vue";
import CollapsedView from "./sidebar/CollapsedView.vue";
import { useMenuExpansion } from "./sidebar/composables/useMenuExpansion";
import { usePlatformConfig } from "./sidebar/composables/usePlatformConfig";
import { useUIStore } from "@/store/ui";

export default defineComponent({
  name: "LeftSidebar",
  components: { SidebarMenu, CollapsedView },
  props: {
    collapsed: { type: Boolean, default: false },
    activeTab: { type: String, default: "logs" },
    activePlatform: { type: String, default: "shopee" },
  },
  emits: ["toggle", "tab-change", "platform-change", "product-platform-change"],
  setup(props) {
    const route = useRoute();
    const uiStore = useUIStore();
    const { expandedSections, toggleExpand } = useMenuExpansion();
    const { platforms, productPlatforms } = usePlatformConfig();
    const windowWidth = ref(
      typeof window !== "undefined" ? window.innerWidth : 1024
    );

    const handleResize = () => {
      windowWidth.value = window.innerWidth;
    };
    onMounted(() => {
      window.addEventListener("resize", handleResize);
      // Auto-expand section based on current route on mount
      uiStore.expandSectionForRoute(route.path);
    });
    onUnmounted(() => window.removeEventListener("resize", handleResize));

    // Watch route changes to close unrelated submenus
    watch(
      () => route.path,
      (newPath) => {
        uiStore.expandSectionForRoute(newPath);
      }
    );

    const showOverlay = computed(
      () => windowWidth.value < 768 && !props.collapsed
    );

    return {
      expandedSections,
      toggleExpand,
      platforms,
      productPlatforms,
      showOverlay,
    };
  },
});
</script>

<style scoped>
.left-sidebar-wrapper {
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
.left-sidebar {
  width: 200px;
  background: var(--color-bg-primary, #ffffff);
  color: var(--color-text-primary, #1f2937);
  display: flex;
  flex-direction: column;
  position: relative;
  border-right: 1px solid var(--color-border, #e5e7eb);
  z-index: var(--z-sidebar, 500);
  box-shadow: var(--shadow-sm);
  height: 100%;
}
.left-sidebar.collapsed {
  width: 56px;
}

/* Toggle Button */
.sidebar-toggle {
  position: absolute;
  top: 12px;
  right: -14px;
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

/* Header */
.sidebar-header {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--color-border, #e5e7eb);
  background: var(--color-bg-secondary, #f8fafc);
}
.header-icon {
  font-size: 1.35rem;
  flex-shrink: 0;
}
.app-name {
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--color-text-primary);
  letter-spacing: 0.5px;
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .left-sidebar {
    width: 180px;
  }
  .left-sidebar.collapsed {
    width: 56px;
  }
  .sidebar-toggle {
    width: 24px;
    height: 24px;
  }
}

/* Responsive - Mobile */
@media (max-width: 767px) {
  .left-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100vh;
    width: 260px;
    transform: translateX(-100%);
    box-shadow: var(--shadow-xl);
  }
  .left-sidebar.mobile-open {
    transform: translateX(0);
  }
  .left-sidebar.collapsed {
    transform: translateX(-100%);
  }
  .sidebar-toggle {
    right: -40px;
    top: 16px;
    width: 32px;
    height: 32px;
  }
}

@media (max-width: 480px) {
  .left-sidebar {
    width: 100%;
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
