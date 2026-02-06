<template>
  <nav
    v-if="shouldShow"
    class="mobile-bottom-nav"
    role="navigation"
    aria-label="Mobile navigation"
  >
    <div class="nav-container">
      <button
        v-for="item in navItems"
        :key="item.id"
        :class="['nav-item', { active: isActive(item.id) }]"
        @click="handleNavClick(item)"
        :aria-current="isActive(item.id) ? 'page' : undefined"
        :aria-label="item.label"
      >
        <span class="nav-icon">{{ item.icon }}</span>
        <span class="nav-label">{{ item.label }}</span>
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAuthStore } from "@/store/authStore";

interface NavItem {
  id: string;
  label: string;
  icon: string;
  route: string;
}

const route = useRoute();
const router = useRouter();
const uiStore = useUIStore();
const authStore = useAuthStore();

// Only show on mobile when authenticated
const shouldShow = computed(() => {
  const isMobile = window.innerWidth <= 768;
  const isAuthenticated = authStore.isAuthenticated;
  const isAuthPage = route.path === "/login" || route.path === "/register";
  return isMobile && isAuthenticated && !isAuthPage;
});

const navItems: NavItem[] = [
  {
    id: "product-management",
    label: "Products",
    icon: "📦",
    route: "/product-manager",
  },
  { id: "order-management", label: "Orders", icon: "🛒", route: "/orders" },
  { id: "inventory", label: "Inventory", icon: "📊", route: "/inventory" },
  { id: "settings", label: "Settings", icon: "⚙️", route: "/settings" },
];

const isActive = (id: string): boolean => {
  return uiStore.activeTab === id;
};

const handleNavClick = (item: NavItem) => {
  if (item.route) {
    uiStore.setActiveTab(item.id as any);
    router.push(item.route);
  }
};
</script>

<style scoped>
.mobile-bottom-nav {
  display: none;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--color-bg-primary, #ffffff);
  border-top: 1px solid var(--color-border, #e5e7eb);
  z-index: var(--z-fixed, 200);
  padding-bottom: env(safe-area-inset-bottom, 0);
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
}

.dark .mobile-bottom-nav {
  background: var(--color-bg-secondary, #1f2937);
  border-top-color: var(--color-border, #374151);
}

.nav-container {
  display: flex;
  justify-content: space-around;
  align-items: stretch;
  height: 56px;
  max-width: 100%;
  padding: 0 4px;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-width: 0;
  padding: 6px 4px;
  background: transparent;
  border: none;
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  transition: all var(--transition-fast, 150ms ease);
  border-radius: var(--radius-md, 0.375rem);
  gap: 2px;
}

.nav-item:hover {
  background: var(--color-bg-tertiary, #f3f4f6);
}

.dark .nav-item:hover {
  background: var(--color-bg-tertiary, #374151);
}

.nav-item.active {
  color: var(--color-primary-600, #2563eb); /* Changed for WCAG AA compliance */
}

.nav-item.active .nav-icon {
  transform: scale(1.1);
}

.nav-icon {
  font-size: 1.25rem;
  line-height: 1;
  transition: transform var(--transition-fast);
}

.nav-label {
  font-size: 0.65rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

/* Show only on mobile */
@media (max-width: 768px) {
  .mobile-bottom-nav {
    display: block;
  }
}

/* Landscape phone adjustments */
@media (max-width: 768px) and (orientation: landscape) {
  .nav-container {
    height: 48px;
  }

  .nav-label {
    display: none;
  }

  .nav-icon {
    font-size: 1.5rem;
  }
}

/* Very small screens */
@media (max-width: 360px) {
  .nav-label {
    font-size: 0.6rem;
  }

  .nav-icon {
    font-size: 1.1rem;
  }
}
</style>
