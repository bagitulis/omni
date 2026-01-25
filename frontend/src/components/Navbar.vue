<template>
  <nav class="navbar">
    <div class="navbar-container">
      <!-- Left: Logo + Connection Status -->
      <div class="navbar-left">
        <router-link to="/" class="brand-link">
          <span class="brand-icon">📦</span>
          <span class="brand-text hidden sm:inline">MultiTenant</span>
        </router-link>

        <!-- Connection Status Badge (next to logo) -->
        <div
          v-if="authStore.isAuthenticated"
          :class="['status-badge', connectionStatus]"
        >
          <span class="status-dot"></span>
          <span class="status-text hidden sm:inline">{{
            connectionStatusText
          }}</span>
        </div>
      </div>

      <!-- Right: Actions & User Info -->
      <div v-if="authStore.isAuthenticated" class="navbar-right">
        <!-- Token Status Dropdown -->
        <TokenStatusDropdown />

        <!-- Theme Toggle -->
        <button
          @click="toggleTheme"
          class="btn-icon"
          :title="isDark ? 'Light Mode' : 'Dark Mode'"
          aria-label="Toggle theme"
        >
          {{ isDark ? "☀️" : "🌙" }}
        </button>

        <!-- Tenant Switcher (Developer Only) -->
        <div v-if="isDeveloper" class="tenant-switcher hidden lg:flex">
          <select
            v-model="selectedTenant"
            @change="switchTenant"
            class="tenant-select"
            aria-label="Select tenant"
          >
            <option value="" disabled>Tenant</option>
            <option
              v-for="tenant in availableTenants"
              :key="tenant.id"
              :value="tenant.id"
            >
              {{ tenant.name }}
            </option>
          </select>
        </div>

        <span class="user-info hidden md:flex">
          👤 {{ authStore.user?.username }}
          <span class="role-badge" :class="userRole">{{ userRole }}</span>
        </span>

        <!-- Admin Panel Button -->
        <router-link
          v-if="canAccessAdmin"
          to="/admin/dashboard"
          class="btn-admin hidden sm:inline-flex"
          title="Admin Panel"
        >
          ⚙️
        </router-link>

        <button @click="handleLogout" class="btn-logout">
          <span class="hidden sm:inline">Logout</span>
          <span class="sm:hidden">🚪</span>
        </button>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../store/authStore";
import { authService } from "../services/authService";
import api from "../services/api";
import PermissionService, {
  type UserRole,
} from "../services/permissionService";
import {
  useUnifiedHeader,
  useConnectionStatusText,
} from "../composables/useUnifiedHeader";
import { useTheme } from "../composables/useTheme";
import TokenStatusDropdown from "./layout/TokenStatusDropdown.vue";

const router = useRouter();
const authStore = useAuthStore();
const { isDarkMode, toggleTheme: themeToggle } = useTheme();

// Unified header state
const { connectionStatus } = useUnifiedHeader();

const connectionStatusText = useConnectionStatusText(connectionStatus);
const isDark = computed(() => isDarkMode.value);

// Tenant state
const currentTenantId = localStorage.getItem("tenantId") || "";
const selectedTenant = ref(currentTenantId);
const availableTenants = ref<{ id: string; name: string }[]>([]);

const userRole = computed(() => (authStore.user?.role || "user") as UserRole);
const isDeveloper = computed(() => userRole.value === "developer");
const canAccessAdmin = computed(() =>
  PermissionService.canAccessAdmin(userRole.value),
);

const toggleTheme = () => {
  themeToggle();
};

const loadTenants = async () => {
  if (!isDeveloper.value) {
    availableTenants.value = [];
    return;
  }

  try {
    const data = await api.get("/auth/tenants");
    const tenants = data?.tenants || [];
    availableTenants.value = tenants.map(
      (t: { id: string; shopName?: string; name?: string }) => ({
        id: t.id,
        name: t.shopName || t.name || t.id,
      }),
    );

    if (!selectedTenant.value && tenants.length > 0) {
      selectedTenant.value = tenants[0].id;
    }
  } catch (error) {
    console.error("Failed to load tenants:", error);
    availableTenants.value = [];
  }
};

const switchTenant = async () => {
  if (!selectedTenant.value) return;

  try {
    const data = await api.post("/auth/switch-tenant", {
      tenantId: selectedTenant.value,
    });

    if (data?.token) {
      localStorage.setItem("authToken", data.token);
    }
    localStorage.setItem("tenantId", selectedTenant.value);
    window.location.reload();
  } catch (error) {
    console.error("Failed to switch tenant:", error);
    localStorage.setItem("tenantId", selectedTenant.value);
    window.location.reload();
  }
};

const handleLogout = async () => {
  await authService.logout();
  authStore.clearAuth();
  router.push("/login");
};

watch(
  isDeveloper,
  (newVal) => {
    if (newVal) loadTenants();
  },
  { immediate: true },
);

onMounted(() => {
  if (isDeveloper.value) loadTenants();
});
</script>

<style scoped>
@import "./styles/Navbar.styles.css";
</style>
