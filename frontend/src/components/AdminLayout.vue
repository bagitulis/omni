<template>
  <div class="admin-layout">
    <AdminSidebar />
    <div class="admin-main">
      <AdminTopBar @change-password="showChangePassword = true" />
      <main class="admin-content">
        <router-view />
      </main>
    </div>

    <ChangePasswordModal
      v-if="showChangePassword"
      @closed="showChangePassword = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import AdminSidebar from "../components/AdminSidebar.vue";
import AdminTopBar from "../components/AdminTopBar.vue";
import ChangePasswordModal from "../components/ChangePasswordModal.vue";

const showChangePassword = ref(false);
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  background: #f5f5f5;
}

.admin-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.admin-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

@media (max-width: 768px) {
  .admin-layout {
    flex-direction: column;
  }

  .admin-main {
    height: calc(100vh - 48px);
  }
}
</style>
