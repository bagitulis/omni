<template>
  <div class="submenu">
    <router-link
      v-for="platform in productPlatforms"
      :key="platform.value"
      :to="`/product-manager/${platform.value}`"
      class="submenu-item"
      :class="{ active: isActive && activePlatform === platform.value }"
    >
      <PlatformBadge
        :platform="platform.value"
        size="xs"
        class="submenu-platform-badge"
      />
      <span>{{ platform.label }}</span>
    </router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from "vue";
import PlatformBadge from "@/components/PlatformBadge.vue";

interface Platform {
  value: string;
  label: string;
  icon: string;
}

export default defineComponent({
  name: "MenuProductManagerSection",
  components: { PlatformBadge },
  props: {
    productPlatforms: {
      type: Array as PropType<Platform[]>,
      required: true,
    },
    isActive: {
      type: Boolean,
      default: false,
    },
    activePlatform: {
      type: String,
      default: "shopee",
    },
  },
});
</script>

<style scoped>
.submenu {
  background: linear-gradient(135deg, hsl(var(--b2) / 0.8) 0%, hsl(var(--b2) / 0.5) 100%);
  border-left: 3px solid hsl(var(--p) / 0.4);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.submenu-item {
  padding: 9px 16px 9px 50px;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: hsl(var(--bc) / 0.8);
  text-decoration: none;
  border-radius: 0.375rem;
  margin: 0 8px;
  font-size: 0.9rem;
}

.submenu-item:hover {
  background: linear-gradient(135deg, hsl(var(--p) / 0.1) 0%, hsl(var(--pf) / 0.08) 100%);
  color: hsl(var(--p));
  padding-left: 52px;
  box-shadow: inset 0 1px 2px rgba(59, 130, 246, 0.1);
}

.submenu-item.router-link-active,
.submenu-item.active {
  background: hsl(var(--b1));
  color: hsl(var(--p));
  border-left: 3px solid hsl(var(--p));
  font-weight: 600;
  border-radius: 0.375rem;
}

.submenu-platform-badge {
  flex-shrink: 0;
  min-width: 24px;
}
</style>
