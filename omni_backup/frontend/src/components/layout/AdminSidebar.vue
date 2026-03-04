<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <h1>Admin Panel</h1>
      <p class="user-info">{{ username }} ({{ role }})</p>
    </div>

    <nav class="sidebar-nav">
      <router-link
        to="/admin/dashboard"
        class="nav-item"
        :class="{ active: isActive('/admin/dashboard') }"
      >
        <span class="icon">📊</span>
        <span class="label">Dashboard</span>
      </router-link>

      <router-link
        v-if="role === 'admin' || role === 'owner'"
        to="/admin/users"
        class="nav-item"
        :class="{ active: isActive('/admin/users') }"
      >
        <span class="icon">👥</span>
        <span class="label">User Management</span>
      </router-link>

      <router-link
        v-if="role === 'owner'"
        to="/admin/roles"
        class="nav-item"
        :class="{ active: isActive('/admin/roles') }"
      >
        <span class="icon">🔑</span>
        <span class="label">Role Management</span>
      </router-link>

      <router-link
        v-if="role === 'admin' || role === 'owner'"
        to="/admin/audit"
        class="nav-item"
        :class="{ active: isActive('/admin/audit') }"
      >
        <span class="icon">📋</span>
        <span class="label">Audit Logs</span>
      </router-link>

      <router-link
        v-if="role?.toLowerCase() === 'developer'"
        to="/admin/shop-setup"
        class="nav-item"
        :class="{ active: isActive('/admin/shop-setup') }"
      >
        <span class="icon">🏪</span>
        <span class="label">Shop Setup</span>
      </router-link>

      <router-link
        to="/admin/settings"
        class="nav-item"
        :class="{ active: isActive('/admin/settings') }"
      >
        <span class="icon">⚙️</span>
        <span class="label">Settings</span>
      </router-link>

      <div class="nav-divider"></div>

      <button
        type="button"
        @click="$emit('show-password-modal')"
        class="nav-item nav-button"
      >
        <span class="icon">🔐</span>
        <span class="label">Change Password</span>
      </button>

      <button
        type="button"
        @click="$emit('logout')"
        class="nav-item nav-button danger"
      >
        <span class="icon">🚪</span>
        <span class="label">Logout</span>
      </button>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";

interface Props {
  username: string;
  role: string;
}

defineProps<Props>();
defineEmits(["show-password-modal", "logout"]);

const route = useRoute();

const isActive = (path: string): boolean => {
  return route.path === path || route.path.startsWith(path);
};
</script>

<style scoped>
.sidebar {
  width: 250px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  overflow-y: auto;
  box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.sidebar-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
}

.user-info {
  margin: 5px 0 0 0;
  font-size: 12px;
  opacity: 0.8;
}

.sidebar-nav {
  flex: 1;
  padding: 15px 0;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  color: rgba(255, 255, 255, 0.9);
  text-decoration: none;
  cursor: pointer;
  transition: all 0.3s ease;
  border-left: 3px solid transparent;
}

/* Style button nav items to match links */
.nav-button {
  background: none;
  border: none;
  border-left: 3px solid transparent;
  width: 100%;
  text-align: left;
  font-family: inherit;
}

.nav-item:hover,
.nav-item.active {
  background-color: rgba(0, 0, 0, 0.2);
  border-left-color: white;
  color: white;
}

.nav-item.danger:hover {
  background-color: rgba(220, 53, 69, 0.3);
  border-left-color: #dc3545;
}

.nav-item .icon {
  font-size: 18px;
  margin-right: 12px;
  width: 24px;
  text-align: center;
}

.nav-item .label {
  font-size: 14px;
  font-weight: 500;
}

.nav-divider {
  height: 1px;
  background-color: rgba(255, 255, 255, 0.2);
  margin: 10px 0;
}

.sidebar-nav::-webkit-scrollbar {
  width: 6px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 3px;
}

.sidebar-nav::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}

/* Nav button (styled like nav-item for semantic correctness) */
.nav-button {
  background: none;
  border: none;
  width: 100%;
  text-align: left;
  font-family: inherit;
}

@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    height: auto;
    max-height: 60px;
    flex-direction: row;
    justify-content: space-between;
  }

  .sidebar-header {
    padding: 10px 15px;
    border-bottom: none;
    border-right: 1px solid rgba(255, 255, 255, 0.2);
  }

  .sidebar-header h1 {
    font-size: 16px;
  }

  .user-info {
    display: none;
  }

  .sidebar-nav {
    display: flex;
    padding: 0;
    gap: 0;
    flex-wrap: wrap;
  }

  .nav-item {
    padding: 10px 15px;
    border-left: none;
    border-bottom: 3px solid transparent;
  }

  .nav-item.active {
    border-left: none;
    border-bottom-color: white;
  }

  .nav-item .label {
    display: none;
  }
}
</style>
