<template>
  <div class="admin-layout">
    <AdminSidebar
      :username="currentUser.username"
      :role="currentUser.role"
      @show-password-modal="showChangePasswordModal = true"
      @logout="logout"
    />

    <main class="main-content">
      <div class="top-bar">
        <div class="breadcrumb">{{ breadcrumb }}</div>
        <div class="top-bar-actions">
          <button
            @click="toggleFullscreen"
            class="icon-btn"
            title="Fullscreen"
            aria-label="Toggle fullscreen mode"
          >
            <span aria-hidden="true">🖥️</span>
          </button>
        </div>
      </div>
      <div class="page-content">
        <router-view></router-view>
      </div>
    </main>

    <ChangePasswordModal
      v-model="showChangePasswordModal"
      :message="passwordMessage"
      :status="passwordStatus"
      @close="showChangePasswordModal = false"
      @change-password="handleChangePassword"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import AdminSidebar from "../components/layout/AdminSidebar.vue";
import ChangePasswordModal from "../components/layout/ChangePasswordModal.vue";
import UserManagementService from "../services/userManagementService";
import { useAppStore } from "@/store/app";

const router = useRouter();
const route = useRoute();
const appStore = useAppStore();

onMounted(() => {
  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }
});

const showChangePasswordModal = ref(false);
const passwordMessage = ref("");
const passwordStatus = ref<"success" | "error" | "">("");

const currentUser = ref({
  username: localStorage.getItem("userName") || "User",
  role: localStorage.getItem("userRole") || "user",
  tenantId: localStorage.getItem("tenantId") || "",
});

const breadcrumb = computed(() => {
  const pathMap: Record<string, string> = {
    "/admin/dashboard": "Dashboard",
    "/admin/users": "User Management",
    "/admin/roles": "Role Management",
    "/admin/audit": "Audit Logs",
    "/admin/settings": "Settings",
    "/admin/shop-setup": "Shop Setup",
  };
  return pathMap[route.path] || "Admin Panel";
});

const toggleFullscreen = async () => {
  try {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen();
    } else {
      await document.exitFullscreen();
    }
  } catch (error) {
    console.error("Fullscreen error:", error);
  }
};

const handleChangePassword = async (form: {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}) => {
  if (form.newPassword !== form.confirmPassword) {
    passwordStatus.value = "error";
    passwordMessage.value = "Passwords do not match";
    return;
  }

  if (form.newPassword.length < 6) {
    passwordStatus.value = "error";
    passwordMessage.value = "Password must be at least 6 characters";
    return;
  }

  try {
    await UserManagementService.changePassword(
      currentUser.value.tenantId,
      form.currentPassword,
      form.newPassword,
    );
    passwordStatus.value = "success";
    passwordMessage.value = "Password changed successfully!";
    setTimeout(() => {
      showChangePasswordModal.value = false;
    }, 1500);
  } catch (error: any) {
    passwordStatus.value = "error";
    passwordMessage.value = error.message;
  }
};

const logout = () => {
  if (confirm("Are you sure you want to logout?")) {
    localStorage.removeItem("token");
    localStorage.removeItem("username");
    localStorage.removeItem("role");
    localStorage.removeItem("tenantId");
    router.push("/login");
  }
};
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  background-color: #f5f5f5;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 30px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.breadcrumb {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.top-bar-actions {
  display: flex;
  gap: 10px;
}

.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.icon-btn:hover {
  background-color: #f0f0f0;
}

.page-content {
  flex: 1;
  overflow-y: auto;
  padding: 30px;
}

@media (max-width: 768px) {
  .admin-layout {
    flex-direction: column;
  }

  .page-content {
    padding: 15px;
  }
}
</style>
