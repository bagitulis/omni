<template>
  <div class="section">
    <AuthSection
      :authStatus="authStatus"
      :loading="loading"
      :authLoading="authLoading"
      @initiate-auth="$emit('initiate-auth')"
      @logout="$emit('logout')"
      @refresh-auth="$emit('refresh-auth')"
    />
  </div>
</template>

<script setup lang="ts">
import AuthSection from "./AuthSection.vue";

interface AuthStatus {
  authenticated: boolean;
  settings: Record<string, string>;
  has_credentials: boolean;
}

interface Props {
  authStatus: AuthStatus;
  loading: boolean;
  authLoading: boolean;
}

defineProps<Props>();

defineEmits<{
  "initiate-auth": [];
  logout: [];
  "refresh-auth": [];
}>();
</script>

<style scoped>
.section {
  margin-bottom: 30px;
  border: 1px solid #e1e5e9;
  border-radius: 12px;
  overflow: hidden;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}
</style>
