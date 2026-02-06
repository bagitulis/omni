<template>
  <div class="collapsed-menu">
    <router-link
      to="/"
      class="collapsed-item"
      :class="{ active: isActive('dashboard') }"
      title="Dashboard"
      aria-label="Dashboard"
    >
      <span aria-hidden="true">🏠</span>
    </router-link>

    <router-link
      to="/product-manager"
      class="collapsed-item"
      :class="{ active: isActive('product-manager') }"
      title="Product Manager"
      aria-label="Product Manager"
    >
      <span aria-hidden="true">🛒</span>
    </router-link>

    <router-link
      to="/operation/shopee"
      class="collapsed-item"
      :class="{ active: isActive('operation') }"
      title="Operation"
      aria-label="Operation"
    >
      <span aria-hidden="true">▶️</span>
    </router-link>

    <router-link
      to="/order-manager"
      class="collapsed-item"
      :class="{ active: isActive('order-manager') }"
      title="Order Manager"
      aria-label="Order Manager"
    >
      <span aria-hidden="true">📋</span>
    </router-link>

    <router-link
      to="/route-mapping"
      class="collapsed-item"
      :class="{ active: isActive('route-mapping') }"
      title="Route Mapper"
      aria-label="Route Mapper"
    >
      <span aria-hidden="true">🗺️</span>
    </router-link>

    <router-link
      to="/inventory"
      class="collapsed-item"
      :class="{ active: isActive('inventory') }"
      title="Inventory"
      aria-label="Inventory"
    >
      <span aria-hidden="true">📦</span>
    </router-link>

    <router-link
      to="/settings/google-sheets"
      class="collapsed-item"
      :class="{ active: isActive('settings') }"
      title="Settings"
      aria-label="Settings"
    >
      <span aria-hidden="true">⚙️</span>
    </router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useSidebarNavigation } from "./composables/useSidebarNavigation";

export default defineComponent({
  name: "CollapsedView",
  setup() {
    const { isActive } = useSidebarNavigation();
    return { isActive };
  },
});
</script>

<style scoped>
.collapsed-menu {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 8px;
  gap: 12px;
  flex: 1;
  overflow-y: auto;
}

.collapsed-item {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 8px;
  color: hsl(var(--bc));
  transition: all 0.2s ease;
  position: relative;
  text-decoration: none;
  border: none;
  padding: 0;
  background: linear-gradient(
    135deg,
    hsl(var(--b2)) 0%,
    hsl(var(--b2) / 0.8) 100%
  );
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.08),
    inset 0 1px 2px rgba(255, 255, 255, 0.1);
  font-size: 1.4rem;
}

.collapsed-item:hover {
  background: linear-gradient(
    135deg,
    hsl(var(--pf) / 0.15) 0%,
    hsl(var(--p) / 0.1) 100%
  );
  color: hsl(var(--p));
  transform: translateY(-2px);
  box-shadow:
    0 4px 12px rgba(59, 130, 246, 0.15),
    inset 0 1px 2px rgba(255, 255, 255, 0.1);
}

.collapsed-item.router-link-active,
.collapsed-item.active {
  background: linear-gradient(
    135deg,
    hsl(var(--pf) / 0.25) 0%,
    hsl(var(--p) / 0.15) 100%
  );
  color: hsl(var(--p));
  border: none;
  box-shadow:
    0 4px 12px rgba(59, 130, 246, 0.2),
    inset 0 0 0 2px hsl(var(--p) / 0.3);
}

.collapsed-menu::-webkit-scrollbar {
  width: 5px;
}

.collapsed-menu::-webkit-scrollbar-track {
  background: hsl(var(--b2));
  border-radius: 2px;
}

.collapsed-menu::-webkit-scrollbar-thumb {
  background: hsl(var(--b3));
  border-radius: 2px;
}

.collapsed-menu::-webkit-scrollbar-thumb:hover {
  background: hsl(var(--p) / 0.5);
}

@media (max-width: 1024px) {
  .collapsed-item {
    width: 40px;
    height: 40px;
    font-size: 1.3rem;
  }
}

@media (max-width: 480px) {
  .collapsed-item {
    width: 44px;
    height: 44px;
    font-size: 1.3rem;
  }
}
</style>
