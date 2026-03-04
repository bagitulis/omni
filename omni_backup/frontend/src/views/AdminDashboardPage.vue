<template>
  <div class="dashboard-page">
    <h1 class="visually-hidden">Admin Dashboard</h1>
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" aria-hidden="true">👥</div>
        <div class="stat-content">
          <h2>Total Users</h2>
          <p class="stat-value">{{ totalUsers }}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" aria-hidden="true">🔐</div>
        <div class="stat-content">
          <h2>Admins</h2>
          <p class="stat-value">{{ adminCount }}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" aria-hidden="true">👤</div>
        <div class="stat-content">
          <h2>Your Role</h2>
          <p class="stat-value">{{ userRoleDisplay }}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" aria-hidden="true">📋</div>
        <div class="stat-content">
          <h2>Audit Logs</h2>
          <p class="stat-value">{{ auditCount }}</p>
        </div>
      </div>
    </div>

    <div class="quick-actions">
      <h2>Quick Actions</h2>
      <div class="actions-grid">
        <router-link to="/admin/users" class="action-card">
          <span class="icon" aria-hidden="true">👥</span>
          <span>Manage Users</span>
        </router-link>
        <router-link to="/admin/roles" class="action-card">
          <span class="icon" aria-hidden="true">🔐</span>
          <span>View Roles</span>
        </router-link>
        <router-link to="/admin/audit" class="action-card">
          <span class="icon" aria-hidden="true">📋</span>
          <span>View Logs</span>
        </router-link>
        <router-link to="/admin/settings" class="action-card">
          <span class="icon" aria-hidden="true">⚙️</span>
          <span>Settings</span>
        </router-link>
      </div>
    </div>

    <div class="recent-activity">
      <h2>Recent Activity</h2>
      <div v-if="recentLogs.length === 0" class="empty">No activity yet</div>
      <div v-else class="activity-list">
        <div
          v-for="log in recentLogs"
          :key="`${log.timestamp}`"
          class="activity-item"
        >
          <div class="activity-time">{{ formatTime(log.timestamp) }}</div>
          <div class="activity-action">{{ log.action }}</div>
          <div class="activity-user">{{ log.userId || "system" }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import UserManagementService, {
  type User,
} from "../services/userManagementService";
import AuditService, { type AuditLog } from "../services/auditService";
import PermissionService, {
  type UserRole,
} from "../services/permissionService";

const users = ref<User[]>([]);
const auditLogs = ref<AuditLog[]>([]);
const userRole = (localStorage.getItem("userRole") || "user") as UserRole;

const totalUsers = computed(() => users.value.length);
const adminCount = computed(
  () => users.value.filter((u: User) => u.role === "admin").length
);
const auditCount = computed(() => auditLogs.value.length);
const userRoleDisplay = computed(() =>
  PermissionService.getRoleDisplayName(userRole)
);
const recentLogs = computed(() => auditLogs.value.slice(0, 5));

const formatTime = (timestamp: string | undefined) => {
  if (!timestamp) return "-";
  try {
    const date = new Date(timestamp);
    return date.toLocaleTimeString();
  } catch {
    return timestamp;
  }
};

const loadData = async () => {
  try {
    const usersData = await UserManagementService.listAllUsers();
    users.value = usersData || [];

    const logsData = await AuditService.getTenantAuditLogs(20);
    auditLogs.value = logsData || [];
  } catch (error) {
    console.error("Failed to load dashboard data:", error);
  }
};

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.dashboard-page {
  padding: 20px 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.stat-card {
  background: white;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  gap: 15px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  font-size: 32px;
}

.stat-content h2 {
  margin: 0 0 5px 0;
  color: #6b7280;
  font-size: 14px;
  text-transform: uppercase;
  font-weight: 600;
}

.stat-value {
  margin: 0;
  color: #333;
  font-size: 24px;
  font-weight: bold;
}

.quick-actions {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 40px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.quick-actions h2 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 1.25rem;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 15px;
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 20px;
  background: #f5f5f5;
  border: 1px solid #eee;
  border-radius: 8px;
  text-decoration: none;
  color: #333;
  transition: all 0.2s;
}

.action-card:hover {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

.action-card .icon {
  font-size: 24px;
}

.action-card span:last-child {
  font-weight: 600;
  text-align: center;
  font-size: 14px;
}

.recent-activity {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.recent-activity h2 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 1.25rem;
}

/* Visually hidden but accessible to screen readers */
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.empty {
  text-align: center;
  color: #6b7280;
  padding: 20px;
}

.activity-list {
  max-height: 400px;
  overflow-y: auto;
}

.activity-item {
  display: grid;
  grid-template-columns: 80px 1fr 100px;
  gap: 15px;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid #eee;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-time {
  color: #6b7280;
  font-size: 12px;
}

.activity-action {
  color: #333;
  font-weight: 500;
}

.activity-user {
  text-align: right;
  color: #667eea;
  font-size: 12px;
}
</style>
