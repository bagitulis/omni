<template>
  <header class="admin-topbar">
    <div class="topbar-left">
      <h3>{{ pageTitle }}</h3>
    </div>

    <div class="topbar-right">
      <div class="user-info">
        <span class="user-name">{{ userName }}</span>
        <span class="user-role">{{ userRoleDisplay }}</span>
      </div>

      <button @click="$emit('change-password')" class="action-btn">
        <span aria-hidden="true">🔑</span> Change Password
      </button>

      <button
        @click="toggleFullscreen"
        class="action-btn"
        aria-label="Toggle fullscreen"
      >
        <span aria-hidden="true">{{ isFullscreen ? "🗗" : "⛶" }}</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import PermissionService, {
  type UserRole,
} from "../services/permissionService";

defineEmits<{
  "change-password": [];
}>();

const route = useRoute();
const isFullscreen = ref(false);

const userRole = (localStorage.getItem("userRole") || "user") as UserRole;
const userName = localStorage.getItem("userName") || "User";

const pageTitle = computed(() => {
  const routeName = route.name || "Dashboard";
  return String(routeName)
    .replace(/([A-Z])/g, " $1")
    .trim();
});

const userRoleDisplay = computed(() => {
  return PermissionService.getRoleDisplayName(userRole);
});

const toggleFullscreen = async () => {
  try {
    if (!isFullscreen.value) {
      await document.documentElement.requestFullscreen();
      isFullscreen.value = true;
    } else {
      await document.exitFullscreen();
      isFullscreen.value = false;
    }
  } catch (error) {
    console.error("Fullscreen toggle failed:", error);
  }
};
</script>

<style scoped>
.admin-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background: white;
  border-bottom: 1px solid #eee;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.topbar-left h3 {
  margin: 0;
  color: #333;
  font-size: 18px;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.user-name {
  font-weight: 600;
  color: #333;
}

.user-role {
  font-size: 12px;
  color: #6b7280;
}

.action-btn {
  padding: 8px 12px;
  background: #f0f0f0;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
  color: #333;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #e8e8e8;
  border-color: #ccc;
}

@media (max-width: 768px) {
  .admin-topbar {
    padding: 10px 15px;
  }

  .topbar-left h3 {
    font-size: 16px;
  }

  .topbar-right {
    gap: 10px;
  }

  .action-btn {
    padding: 6px 10px;
    font-size: 12px;
  }
}
</style>
