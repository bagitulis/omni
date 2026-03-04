<template>
  <div class="roles-page">
    <RoleManager />
  </div>
</template>

<script setup lang="ts">
import RoleManager from "../components/RoleManager.vue";
</script>

<style scoped>
.roles-page {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
</style>
