<template>
  <div class="users-page">
    <div class="page-header">
      <h1>User Management</h1>
      <button @click="showCreateForm = true" class="btn btn-primary">
        + Create User
      </button>
    </div>

    <div class="filters">
      <label for="user-search" class="sr-only">Search users</label>
      <input
        id="user-search"
        v-model="searchQuery"
        type="text"
        placeholder="Search by username or email..."
        class="search-input"
      />
    </div>

    <div v-if="loading" class="loading">Loading users...</div>

    <div v-else-if="filteredUsers.length === 0" class="empty">
      No users found
    </div>

    <table v-else class="users-table" aria-label="User accounts">
      <thead>
        <tr>
          <th scope="col">Username</th>
          <th scope="col">Email</th>
          <th scope="col">Role</th>
          <th scope="col">Created</th>
          <th scope="col">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in filteredUsers" :key="user.tenantId">
          <td>{{ user.username }}</td>
          <td>{{ user.email }}</td>
          <td>
            <span :class="['role-badge', user.role]">{{ user.role }}</span>
          </td>
          <td>{{ formatDate(user.createdAt) }}</td>
          <td>
            <button
              @click="editUser(user)"
              class="action-btn edit"
              title="Edit"
              aria-label="Edit user"
            >
              <span aria-hidden="true">✏️</span>
            </button>
            <button
              @click="deleteUser(user.tenantId)"
              class="action-btn delete"
              title="Delete"
              aria-label="Delete user"
            >
              <span aria-hidden="true">🗑️</span>
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <UserCreateForm
      v-if="showCreateForm"
      @created="handleUserCreated"
      @closed="showCreateForm = false"
    />

    <UserEditForm
      v-if="selectedUser && selectedUser.username && selectedUser.email"
      :user="selectedUser as Required<typeof selectedUser>"
      @saved="handleUserSaved"
      @closed="selectedUser = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import UserManagementService, {
  type User,
} from "../services/userManagementService";
import UserCreateForm from "../components/UserCreateForm.vue";
import UserEditForm from "../components/UserEditForm.vue";

const users = ref<User[]>([]);
const loading = ref(true);
const searchQuery = ref("");
const showCreateForm = ref(false);
const selectedUser = ref<User | null>(null);

const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value;
  const query = searchQuery.value.toLowerCase();
  return users.value.filter(
    (u) =>
      (u.username?.toLowerCase().includes(query) ?? false) ||
      (u.email?.toLowerCase().includes(query) ?? false),
  );
});

const loadUsers = async () => {
  loading.value = true;
  try {
    const data = await UserManagementService.listAllUsers();
    users.value = data || [];
  } catch (error) {
    console.error("Failed to load users:", error);
  } finally {
    loading.value = false;
  }
};

const editUser = (user: User) => {
  selectedUser.value = user;
};

const deleteUser = async (tenantId: string) => {
  if (!confirm("Are you sure you want to delete this user?")) return;

  try {
    await UserManagementService.deleteUser(tenantId);
    users.value = users.value.filter((u) => u.tenantId !== tenantId);
  } catch (error) {
    console.error("Failed to delete user:", error);
  }
};

const formatDate = (date: string) => {
  try {
    return new Date(date).toLocaleDateString();
  } catch {
    return date;
  }
};

const handleUserCreated = () => {
  showCreateForm.value = false;
  loadUsers();
};

const handleUserSaved = () => {
  selectedUser.value = null;
  loadUsers();
};

onMounted(() => {
  loadUsers();
});
</script>

<style scoped>
.users-page {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.btn-primary {
  padding: 10px 20px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-primary:hover {
  background: #5568d3;
}

.filters {
  margin-bottom: 20px;
}

.search-input {
  width: 100%;
  max-width: 300px;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.loading,
.empty {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
}

.users-table thead {
  background: #f5f5f5;
  border-bottom: 2px solid #ddd;
}

.users-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #333;
}

.users-table td {
  padding: 12px;
  border-bottom: 1px solid #eee;
}

.users-table tbody tr:hover {
  background: #f9f9f9;
}

.role-badge {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.role-badge.owner {
  background: #ffe0b2;
  color: #e65100;
}

.role-badge.admin {
  background: #c8e6c9;
  color: #1b5e20;
}

.role-badge.user {
  background: #bbdefb;
  color: #0d47a1;
}

.action-btn {
  padding: 6px 10px;
  margin: 0 4px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 16px;
}

.action-btn:hover {
  opacity: 0.7;
}
</style>
