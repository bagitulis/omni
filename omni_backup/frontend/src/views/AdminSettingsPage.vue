<template>
  <div class="settings-page">
    <AdminSettings />
  </div>
</template>

<script setup lang="ts">
import AdminSettings from "../components/AdminSettings.vue";
</script>

<style scoped>
.settings-page {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
</style>
