<template>
  <div class="modal-overlay">
    <div class="modal">
      <div class="modal-header">
        <h2>Edit User</h2>
        <button
          @click="$emit('closed')"
          class="close-btn"
          type="button"
          aria-label="Close modal"
        >
          ✕
        </button>
      </div>

      <form @submit.prevent="submitForm" class="form">
        <div class="form-group">
          <label for="edit-username">Username</label>
          <input
            id="edit-username"
            v-model="form.username"
            type="text"
            disabled
          />
        </div>

        <div class="form-group">
          <label for="edit-email">Email</label>
          <input id="edit-email" v-model="form.email" type="email" required />
        </div>

        <div class="form-group">
          <label for="edit-role">Role</label>
          <select id="edit-role" v-model="form.role">
            <option value="user">User</option>
            <option value="admin">Admin</option>
            <option value="owner">Owner</option>
            <option value="developer">Developer</option>
          </select>
        </div>

        <div class="form-group">
          <label for="edit-password">New Password (optional)</label>
          <input id="edit-password" v-model="form.password" type="password" />
          <small>Leave empty to keep current password</small>
        </div>

        <div v-if="message" :class="['message', status]">
          {{ message }}
        </div>

        <div class="form-actions">
          <button type="submit" class="btn btn-primary">Save Changes</button>
          <button
            type="button"
            @click="$emit('closed')"
            class="btn btn-secondary"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import UserManagementService from "../services/userManagementService";

interface User {
  tenantId: string;
  username?: string;
  email?: string;
  role?: string;
}

const props = defineProps<{
  user: User;
}>();

const emit = defineEmits<{
  saved: [];
  closed: [];
}>();

const form = ref({
  username: props.user.username,
  email: props.user.email,
  role: props.user.role,
  password: "",
});

const message = ref("");
const status = ref<"success" | "error" | "">("");

watch(
  () => props.user,
  (newUser) => {
    form.value = {
      username: newUser.username,
      email: newUser.email,
      role: newUser.role,
      password: "",
    };
  }
);

const submitForm = async () => {
  if (!form.value.email) {
    status.value = "error";
    message.value = "Email is required";
    return;
  }

  if (form.value.password && form.value.password.length < 6) {
    status.value = "error";
    message.value = "Password must be at least 6 characters";
    return;
  }

  try {
    const updates: any = {
      email: form.value.email,
      role: form.value.role,
    };

    if (form.value.password) {
      updates.password = form.value.password;
    }

    await UserManagementService.updateUser(props.user.tenantId, updates);
    status.value = "success";
    message.value = "User updated successfully!";
    setTimeout(() => {
      emit("saved");
    }, 1000);
  } catch (error: any) {
    status.value = "error";
    message.value = error.message;
  }
};
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 8px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 90%;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #eee;
}

.modal-header h2 {
  margin: 0;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: #6b7280;
}

.form {
  padding: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 600;
  color: #333;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: inherit;
}

.form-group input:disabled {
  background: #f5f5f5;
  cursor: not-allowed;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-group small {
  display: block;
  margin-top: 4px;
  color: #6b7280;
}

.message {
  margin-bottom: 15px;
  padding: 12px;
  border-radius: 4px;
  font-weight: 500;
}

.message.success {
  background: #d4edda;
  color: #155724;
}

.message.error {
  background: #f8d7da;
  color: #721c24;
}

.form-actions {
  display: flex;
  gap: 10px;
}

.btn {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5568d3;
}

.btn-secondary {
  background: #e0e0e0;
  color: #333;
}

.btn-secondary:hover {
  background: #d0d0d0;
}
</style>
