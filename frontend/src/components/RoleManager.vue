<template>
  <div class="role-container">
    <div class="role-header">
      <h1>Role Management</h1>
      <p>View and manage user roles and permissions</p>
    </div>

    <div class="roles-grid">
      <div v-for="role in roles" :key="role" class="role-card">
        <div class="role-title">
          <h3>{{ formatRoleName(role) }}</h3>
          <span class="role-badge">{{ role }}</span>
        </div>

        <div class="role-description">
          {{ getRoleDescription(role) }}
        </div>

        <div class="permissions-list">
          <h4>Permissions:</h4>
          <div v-if="getPermissions(role).length > 0" class="permissions">
            <span
              v-for="permission in getPermissions(role)"
              :key="permission"
              class="permission-tag"
            >
              ✓ {{ formatPermission(permission) }}
            </span>
          </div>
          <div v-else class="no-permissions">No permissions</div>
        </div>

        <div class="role-users">
          <small>
            <strong>Users with this role:</strong>
            {{ getUserCountForRole(role) }}
          </small>
        </div>
      </div>
    </div>

    <div class="permission-matrix">
      <h3>Permission Matrix</h3>
      <table class="matrix-table" aria-label="Permission matrix">
        <thead>
          <tr>
            <th scope="col">Permission</th>
            <th scope="col" v-for="role in roles" :key="role">
              {{ formatRoleName(role) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="permission in allPermissions" :key="permission">
            <td class="permission-name">{{ formatPermission(permission) }}</td>
            <td v-for="role in roles" :key="`${role}-${permission}`">
              <span v-if="hasPermission(role, permission)" class="check"
                >✓</span
              >
              <span v-else class="no-check">✗</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import PermissionService, {
  type UserRole,
} from "../services/permissionService";
import UserManagementService, {
  type User,
} from "../services/userManagementService";

const roles: UserRole[] = ["developer", "owner", "admin", "user"];
const users = ref<User[]>([]);

const userCounts = computed(() => {
  return {
    owner: users.value.filter((u) => u.role === "owner").length,
    admin: users.value.filter((u) => u.role === "admin").length,
    user: users.value.filter((u) => u.role === "user").length,
  };
});

const allPermissions = computed(() => {
  const permissions = new Set<string>();
  roles.forEach((role: UserRole) => {
    PermissionService.getAllPermissions(role).forEach((p) =>
      permissions.add(p),
    );
  });
  return Array.from(permissions).sort();
});

const loadUsers = async () => {
  try {
    users.value = await UserManagementService.listAllUsers();
  } catch (error) {
    console.error("Failed to load users for role counts:", error);
  }
};

const formatRoleName = (role: UserRole) => {
  return role.charAt(0).toUpperCase() + role.slice(1);
};

const getRoleDescription = (role: UserRole) => {
  return PermissionService.getRoleDescription(role);
};

const getPermissions = (role: UserRole) => {
  return PermissionService.getAllPermissions(role);
};

const hasPermission = (role: UserRole, permission: string) => {
  return PermissionService.hasPermission(role, permission);
};

const formatPermission = (permission: string) => {
  return permission
    .split(".")
    .map((p) => p.charAt(0).toUpperCase() + p.slice(1))
    .join(" ");
};

const getUserCountForRole = (role: UserRole) => {
  return userCounts.value[role];
};

onMounted(() => {
  loadUsers();
});
</script>

<style scoped>
.role-container {
  padding: 20px;
  background: #f9f9f9;
}

.role-header {
  margin-bottom: 30px;
}

.role-header h1 {
  margin: 0 0 5px 0;
  color: #333;
}

.role-header p {
  margin: 0;
  color: #666;
}

.roles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.role-card {
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.role-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.role-title h3 {
  margin: 0;
  color: #333;
}

.role-badge {
  background: #667eea;
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.role-description {
  color: #666;
  margin-bottom: 15px;
  line-height: 1.5;
}

.permissions-list h4 {
  margin: 0 0 10px 0;
  color: #333;
}

.permissions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 15px;
}

.permission-tag {
  background: #e8f0fe;
  color: #1967d2;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.no-permissions {
  color: #6b7280;
  font-style: italic;
}

.role-users {
  border-top: 1px solid #eee;
  padding-top: 10px;
  color: #666;
}

.permission-matrix {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.permission-matrix h3 {
  margin: 0 0 15px 0;
  color: #333;
}

.matrix-table {
  width: 100%;
  border-collapse: collapse;
}

.matrix-table thead {
  background: #f5f5f5;
  border-bottom: 2px solid #ddd;
}

.matrix-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #333;
}

.matrix-table td {
  padding: 12px;
  border-bottom: 1px solid #eee;
  text-align: center;
}

.permission-name {
  text-align: left;
  font-weight: 500;
}

.check {
  color: #28a745;
  font-weight: bold;
  font-size: 18px;
}

.no-check {
  color: #dc3545;
  font-weight: bold;
  font-size: 18px;
}

.matrix-table tbody tr:hover {
  background: #f9f9f9;
}
</style>
