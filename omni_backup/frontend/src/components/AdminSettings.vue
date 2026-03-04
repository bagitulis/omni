<template>
  <div class="settings-container">
    <div class="settings-header">
      <h1>Admin Settings</h1>
    </div>

    <div class="settings-content">
      <div class="settings-section">
        <h2>Audit Settings</h2>
        <div class="setting-item">
          <div>
            <label for="audit-cleanup-days">Auto Cleanup Old Logs</label>
            <small>Automatically delete audit logs older than</small>
          </div>
          <div class="input-group">
            <input
              id="audit-cleanup-days"
              v-model="auditCleanupDays"
              type="number"
              min="1"
            />
            <span>days</span>
          </div>
        </div>

        <div class="setting-item">
          <div>
            <label for="enable-audit-logging">Enable Audit Logging</label>
            <small>Log all user actions and changes</small>
          </div>
          <input
            id="enable-audit-logging"
            v-model="enableAuditLogging"
            type="checkbox"
          />
        </div>

        <button @click="saveAuditSettings" class="btn btn-primary">
          Save Audit Settings
        </button>
      </div>

      <div class="settings-section">
        <h2>User Management</h2>
        <div class="setting-item">
          <div>
            <label for="min-password-length">Password Min Length</label>
            <small>Minimum password length for all users</small>
          </div>
          <input
            id="min-password-length"
            v-model="minPasswordLength"
            type="number"
            min="6"
          />
        </div>

        <div class="setting-item">
          <div>
            <label for="require-password-change"
              >Require Password Change on First Login</label
            >
            <small>Force users to change password on first access</small>
          </div>
          <input
            id="require-password-change"
            v-model="requirePasswordChange"
            type="checkbox"
          />
        </div>

        <button @click="saveUserSettings" class="btn btn-primary">
          Save User Settings
        </button>
      </div>

      <div class="settings-section danger-zone">
        <h2>Danger Zone</h2>
        <div class="setting-item">
          <div>
            <label>Delete All Audit Logs</label>
            <small>Permanently remove all audit log history</small>
          </div>
          <button @click="deleteAllLogs" class="btn btn-danger">
            Delete All Logs
          </button>
        </div>
      </div>

      <div v-if="message" :class="['message', messageType]">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import AuditService from "../services/auditService";

const auditCleanupDays = ref(30);
const enableAuditLogging = ref(true);
const minPasswordLength = ref(6);
const requirePasswordChange = ref(false);
const message = ref("");
const messageType = ref<"success" | "error" | "">("");

const saveAuditSettings = async () => {
  try {
    await AuditService.cleanupOldLogs(auditCleanupDays.value);
    message.value = "Audit settings saved successfully!";
    messageType.value = "success";
    setTimeout(() => {
      message.value = "";
    }, 3000);
  } catch (error: any) {
    message.value = error.message;
    messageType.value = "error";
  }
};

const saveUserSettings = () => {
  try {
    // Save to localStorage or backend
    localStorage.setItem(
      "userSettings",
      JSON.stringify({
        minPasswordLength: minPasswordLength.value,
        requirePasswordChange: requirePasswordChange.value,
      })
    );
    message.value = "User settings saved successfully!";
    messageType.value = "success";
    setTimeout(() => {
      message.value = "";
    }, 3000);
  } catch (error: any) {
    message.value = error.message;
    messageType.value = "error";
  }
};

const deleteAllLogs = async () => {
  if (!confirm("Are you sure? This action cannot be undone!")) {
    return;
  }

  try {
    await AuditService.cleanupOldLogs(0);
    message.value = "All audit logs deleted!";
    messageType.value = "success";
    setTimeout(() => {
      message.value = "";
    }, 3000);
  } catch (error: any) {
    message.value = error.message;
    messageType.value = "error";
  }
};
</script>

<style scoped>
.settings-container {
  padding: 20px;
  background: #f9f9f9;
  border-radius: 8px;
}

.settings-header h1 {
  margin: 0 0 30px 0;
  color: #333;
}

.settings-content {
  max-width: 600px;
}

.settings-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.settings-section h2 {
  margin: 0 0 20px 0;
  color: #333;
  border-bottom: 2px solid #667eea;
  padding-bottom: 10px;
  font-size: 1.25rem;
}

.settings-section.danger-zone h2 {
  border-bottom-color: #dc3545;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #eee;
}

.setting-item:last-of-type {
  border-bottom: none;
  padding-bottom: 15px;
}

.setting-item label {
  font-weight: 600;
  color: #333;
}

.setting-item small {
  display: block;
  color: #6b7280;
  margin-top: 5px;
}

.setting-item input[type="number"],
.setting-item input[type="text"] {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: inherit;
}

.setting-item input[type="checkbox"] {
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.input-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-group input {
  width: 80px;
}

.input-group span {
  color: #666;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.2s;
  margin-top: 15px;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5568d3;
}

.btn-danger {
  background: #dc3545;
  color: white;
}

.btn-danger:hover {
  background: #c82333;
}

.message {
  margin-top: 20px;
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
</style>
