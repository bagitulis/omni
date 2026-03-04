<template>
  <aside class="admin-sidebar">
    <div class="sidebar-header">
      <h2>Admin Panel</h2>
    </div>

    <nav class="sidebar-nav">
      <router-link to="/admin/dashboard" class="nav-item" active-class="active">
        <span class="icon">📊</span>
        <span>Dashboard</span>
      </router-link>

      <router-link
        v-if="canManageUsers"
        to="/admin/users"
        class="nav-item"
        active-class="active"
      >
        <span class="icon">👥</span>
        <span>Users</span>
      </router-link>

      <router-link
        v-if="canManageRoles"
        to="/admin/roles"
        class="nav-item"
        active-class="active"
      >
        <span class="icon">🔐</span>
        <span>Roles</span>
      </router-link>

      <router-link
        v-if="canViewAudit"
        to="/admin/audit"
        class="nav-item"
        active-class="active"
      >
        <span class="icon">📋</span>
        <span>Audit Logs</span>
      </router-link>

      <router-link to="/admin/settings" class="nav-item" active-class="active">
        <span class="icon">⚙️</span>
        <span>Settings</span>
      </router-link>
    </nav>

    <div class="sidebar-footer">
      <button @click="logout" class="logout-btn">
        <span class="icon">🚪</span>
        <span>Logout</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import PermissionService, {
  type UserRole,
} from "../services/permissionService";

const router = useRouter();
const userRole = (localStorage.getItem("userRole") || "user") as UserRole;

const canManageUsers = computed(() =>
  PermissionService.canManageUsers(userRole)
);
const canManageRoles = computed(() => userRole === "owner");
const canViewAudit = computed(() =>
  PermissionService.hasPermission(userRole, "audit.view")
);

const logout = () => {
  localStorage.removeItem("authToken");
  localStorage.removeItem("userRole");
  router.push("/login");
};
</script>

<style scoped>
.admin-sidebar {
  width: 260px;
  background: #2c3e50;
  color: white;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow-y: auto;
  box-shadow: 2px 0 4px rgba(0, 0, 0, 0.1);
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-header h2 {
  margin: 0;
  font-size: 20px;
}

.sidebar-nav {
  flex: 1;
  padding: 10px 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  transition: all 0.2s;
  border-left: 3px solid transparent;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.nav-item.active {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
  border-left-color: #667eea;
}

.nav-item .icon {
  font-size: 18px;
}

.sidebar-footer {
  padding: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.logout-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  background: rgba(255, 59, 48, 0.2);
  color: #ff3b30;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.2s;
}

.logout-btn:hover {
  background: rgba(255, 59, 48, 0.3);
}

.logout-btn .icon {
  font-size: 18px;
}

@media (max-width: 768px) {
  .admin-sidebar {
    width: 200px;
  }

  .nav-item {
    padding: 10px 15px;
  }

  .sidebar-header h2 {
    font-size: 16px;
  }
}
</style>
