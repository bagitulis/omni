<template>
  <div v-if="modelValue" class="modal-overlay">
    <div class="modal">
      <h2>Change Password</h2>
      <form @submit.prevent="$emit('change-password', passwordForm)">
        <div class="form-group">
          <label>Current Password</label>
          <input
            v-model="passwordForm.currentPassword"
            type="password"
            required
          />
        </div>
        <div class="form-group">
          <label>New Password</label>
          <input v-model="passwordForm.newPassword" type="password" required />
        </div>
        <div class="form-group">
          <label>Confirm Password</label>
          <input
            v-model="passwordForm.confirmPassword"
            type="password"
            required
          />
        </div>

        <div v-if="message" :class="['message', status]">
          {{ message }}
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary">Change Password</button>
          <button
            type="button"
            @click="$emit('close')"
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
import { ref } from "vue";

interface Props {
  modelValue: boolean;
  message?: string;
  status?: "success" | "error" | "";
}

defineProps<Props>();
defineEmits(["close", "change-password"]);

const passwordForm = ref({
  currentPassword: "",
  newPassword: "",
  confirmPassword: "",
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal {
  background: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 90%;
}

.modal h2 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #333;
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

.form-group input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: inherit;
  font-size: 14px;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.message {
  margin-bottom: 15px;
  padding: 12px;
  border-radius: 4px;
  font-weight: 500;
}

.message.success {
  background-color: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.message.error {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.2s;
  flex: 1;
}

.btn-primary {
  background-color: #667eea;
  color: white;
}

.btn-primary:hover {
  background-color: #5568d3;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background-color: #5a6268;
}
</style>
