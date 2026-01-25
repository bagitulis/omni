<template>
  <div>
    <!-- Settings Tab dengan Submenu -->
    <div class="menu-section" :class="{ active: isActive }">
      <div class="menu-item parent" @click="$emit('toggle')">
        <div class="menu-label-flex">
          <span>⚙️</span>
          <span>Settings</span>
        </div>
        <span class="expand-icon" :class="{ rotated: expanded }">▼</span>
      </div>

      <!-- Submenu untuk Settings -->
      <div class="submenu" v-if="expanded">
        <router-link
          to="/settings/google-sheets"
          class="submenu-item"
          :class="{ active: isSettingsGoogleSheetsActive }"
        >
          <span>🔗</span>
          <span>Google Sheets</span>
        </router-link>
        <router-link
          to="/settings/webhook"
          class="submenu-item"
          :class="{ active: isSettingsWebhookActive }"
        >
          <span>🔗</span>
          <span>Platform Integration</span>
        </router-link>
      </div>
    </div>

    <!-- Script Monitor moved to top-level menu -->
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "MenuSettingsSection",
  props: {
    expanded: {
      type: Boolean,
      default: false,
    },
    isActive: {
      type: Boolean,
      default: false,
    },
    isSettingsGoogleSheetsActive: {
      type: Boolean,
      default: false,
    },
    isSettingsWebhookActive: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["toggle"],
});
</script>

<style scoped>
.menu-section {
  display: flex;
  flex-direction: column;
}

.menu-item {
  padding: 11px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: hsl(var(--bc));
  position: relative;
  text-decoration: none;
  border-left: 3px solid transparent;
  margin: 0 8px;
  border-radius: 0.375rem;
  font-size: 0.95rem;
}

.menu-item:hover {
  background: linear-gradient(
    135deg,
    hsl(var(--pf) / 0.1) 0%,
    hsl(var(--p) / 0.08) 100%
  );
  color: hsl(var(--p));
  border-left-color: hsl(var(--p) / 0.3);
  box-shadow: inset 0 1px 3px rgba(59, 130, 246, 0.1);
}

.menu-item.active {
  background: linear-gradient(
    135deg,
    hsl(var(--pf) / 0.2) 0%,
    hsl(var(--p) / 0.1) 100%
  );
  color: hsl(var(--p));
  border-left-color: hsl(var(--p));
  font-weight: 600;
  box-shadow:
    inset 0 1px 4px rgba(59, 130, 246, 0.15),
    0 1px 3px rgba(59, 130, 246, 0.1);
}

.menu-label-flex {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.menu-label-flex span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.expand-icon {
  font-size: 0.9rem;
  transition: transform 0.3s ease;
  flex-shrink: 0;
  color: #00d9ff;
  font-weight: 700;
}

.expand-icon.rotated {
  transform: rotate(180deg);
}

.submenu {
  background: linear-gradient(
    135deg,
    hsl(var(--b2) / 0.8) 0%,
    hsl(var(--b2) / 0.5) 100%
  );
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
  background: linear-gradient(
    135deg,
    hsl(var(--p) / 0.1) 0%,
    hsl(var(--pf) / 0.08) 100%
  );
  color: hsl(var(--p));
  padding-left: 52px;
  box-shadow: inset 0 1px 2px rgba(59, 130, 246, 0.1);
}

.submenu-item.active,
.submenu-item.router-link-active {
  background: hsl(var(--b1));
  color: hsl(var(--p));
  border-left: 3px solid hsl(var(--p));
  font-weight: 600;
  border-radius: 0.375rem;
}
</style>
