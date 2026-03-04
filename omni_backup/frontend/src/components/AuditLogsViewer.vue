<template>
  <div class="audit-container">
    <div class="audit-header">
      <h1>Audit Logs</h1>
      <div class="filters">
        <label for="filter-action" class="sr-only">Filter by action</label>
        <input
          id="filter-action"
          v-model="filterAction"
          type="text"
          placeholder="Filter by action..."
          class="filter-input"
          aria-label="Filter by action"
        />
        <label for="filter-user" class="sr-only">Filter by user</label>
        <select
          id="filter-user"
          v-model="filterUserId"
          class="filter-select"
          aria-label="Filter by user"
        >
          <option value="">All Users</option>
          <option value="system">System</option>
        </select>
        <button @click="loadLogs" class="btn btn-sm">Refresh</button>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading audit logs...</div>

    <div v-else-if="filteredLogs.length === 0" class="empty">
      No audit logs found
    </div>

    <table v-else class="audit-table" aria-label="Audit logs">
      <thead>
        <tr>
          <th scope="col">Timestamp</th>
          <th scope="col">Action</th>
          <th scope="col">User ID</th>
          <th scope="col">Target</th>
          <th scope="col">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="log in filteredLogs" :key="`${log.timestamp}-${log.action}`">
          <td>{{ formatDate(log.timestamp) }}</td>
          <td>{{ log.action }}</td>
          <td>{{ log.userId || "system" }}</td>
          <td>{{ log.targetUserId || log.targetTenantId || "-" }}</td>
          <td>
            <span :class="['status', log.status]">{{ log.status }}</span>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="logs.length > 0" class="pagination">
      <small>Showing {{ filteredLogs.length }} of {{ logs.length }} logs</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import AuditService, { type AuditLog } from "../services/auditService";

const logs = ref<AuditLog[]>([]);
const loading = ref(true);
const filterAction = ref("");
const filterUserId = ref("");

const filteredLogs = computed(() => {
  return logs.value.filter((log) => {
    const matchAction =
      !filterAction.value ||
      log.action.toLowerCase().includes(filterAction.value.toLowerCase());
    const matchUser =
      !filterUserId.value ||
      log.userId === filterUserId.value ||
      filterUserId.value === "system";
    return matchAction && matchUser;
  });
});

const loadLogs = async () => {
  loading.value = true;
  try {
    const data = await AuditService.getTenantAuditLogs(100);
    logs.value = data || [];
  } catch (error) {
    console.error("Failed to load audit logs:", error);
  } finally {
    loading.value = false;
  }
};

const formatDate = (timestamp: string | undefined) => {
  if (!timestamp) return "-";
  try {
    const date = new Date(timestamp);
    return date.toLocaleString();
  } catch {
    return timestamp;
  }
};

onMounted(() => {
  loadLogs();
});
</script>

<style scoped>
.audit-container {
  padding: 20px;
  background: #f9f9f9;
  border-radius: 8px;
}

.audit-header {
  margin-bottom: 20px;
}

.audit-header h1 {
  margin: 0 0 15px 0;
  color: #333;
}

.filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.filter-input,
.filter-select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: inherit;
}

.filter-input:focus,
.filter-select:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.btn-sm {
  padding: 8px 16px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-sm:hover {
  background: #5568d3;
}

.loading {
  padding: 30px;
  text-align: center;
  color: #4b5563;
}

.empty {
  padding: 30px;
  text-align: center;
  color: #4b5563;
}

.audit-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.audit-table thead {
  background: #f5f5f5;
  border-bottom: 2px solid #ddd;
}

.audit-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #333;
}

.audit-table td {
  padding: 12px;
  border-bottom: 1px solid #eee;
}

.audit-table tbody tr:hover {
  background: #f9f9f9;
}

.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.status.success {
  background: #d4edda;
  color: #155724;
}

.status.error {
  background: #f8d7da;
  color: #721c24;
}

.pagination {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #eee;
  text-align: center;
  color: #666;
}
</style>
