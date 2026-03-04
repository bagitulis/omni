<template>
  <div class="audit-page">
    <AuditLogsViewer />
  </div>
</template>

<script setup lang="ts">
import AuditLogsViewer from "../components/AuditLogsViewer.vue";
</script>

<style scoped>
.audit-page {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
</style>
