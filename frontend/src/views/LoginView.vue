<template>
  <div class="login-container">
    <!-- Auto-login loading overlay -->
    <div v-if="autoLoginInProgress" class="auto-login-overlay">
      <div class="auto-login-spinner"></div>
      <p>Auto-logging in as tester...</p>
    </div>

    <div class="login-card" v-show="!autoLoginInProgress">
      <h1>Login</h1>

      <!-- DEV MODE: Localhost bypass login -->
      <div v-if="isLocalhost" class="dev-mode-section">
        <div class="dev-mode-header">
          <span aria-hidden="true">🔧</span> Dev Mode (Localhost Only)
        </div>
        <div class="dev-mode-content">
          <label for="devTenant">Select Tenant:</label>
          <select
            id="devTenant"
            v-model="selectedDevTenant"
            class="dev-tenant-select"
          >
            <option value="yumna_bertigamart">Yumna - Bertigamart</option>
            <option value="tika_nusseyba">Tika - Nusseyba</option>
          </select>
          <button
            type="button"
            @click="handleDevLogin"
            :disabled="isLoading"
            class="btn-dev-login"
          >
            {{ isLoading ? "Logging in..." : "Quick Dev Login (tester)" }}
          </button>
        </div>
      </div>

      <!-- Session Expired Notice -->
      <div v-if="sessionExpired" class="session-expired-message" role="alert">
        Session expired. Please login again.
      </div>

      <!-- Account Locked Warning -->
      <div v-if="isLocked" class="locked-message" role="alert">
        <span aria-hidden="true">🔒</span>
        Account locked. Try again in {{ lockMinutesRemaining }} minute(s).
      </div>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            placeholder="Enter username"
            required
            :disabled="isLocked"
          />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            placeholder="Enter password"
            required
            :disabled="isLocked"
          />
        </div>

        <!-- reCAPTCHA Notice (when required) -->
        <div v-if="requiresCaptcha" class="captcha-notice">
          <span aria-hidden="true">🛡️</span>
          Protected by reCAPTCHA
        </div>

        <button
          type="submit"
          :disabled="isLoading || isLocked"
          class="btn-submit"
        >
          {{ isLoading ? "Logging in..." : "Login" }}
        </button>
      </form>

      <p v-if="error && !isLocked" class="error-message">{{ error }}</p>

      <!-- reCAPTCHA Branding (required by Google) -->
      <p class="recaptcha-branding">
        This site is protected by reCAPTCHA and the Google
        <a href="https://policies.google.com/privacy" target="_blank"
          >Privacy Policy</a
        >
        and
        <a href="https://policies.google.com/terms" target="_blank"
          >Terms of Service</a
        >
        apply.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "../store/authStore";
import { authService } from "../services/authService";
import { captchaService } from "../services/captchaService";
import api from "../services/api";

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

// Check if running on localhost (for dev bypass)
const isLocalhost = computed(() => {
  const hostname = window.location.hostname;
  return hostname === "localhost" || hostname === "127.0.0.1";
});

// Dev mode state
const selectedDevTenant = ref("yumna_bertigamart");

// Get returnUrl from query params (set when redirected due to expired token)
const returnUrl = computed(() => {
  const url = route.query.returnUrl as string;
  // Validate returnUrl to prevent open redirect attacks
  if (url && url.startsWith("/") && !url.startsWith("//")) {
    return url;
  }
  return "/";
});

// Check if redirected due to session expiry
const sessionExpired = computed(() => route.query.returnUrl !== undefined);

const form = reactive({
  username: "",
  password: "",
});

const isLoading = ref(false);
const error = ref("");
const requiresCaptcha = ref(false);
const isLocked = ref(false);
const lockMinutesRemaining = ref(0);

// Auto dev-login state
const autoLoginAttempted = ref(false);
const autoLoginInProgress = ref(false);

// Initialize reCAPTCHA on mount and trigger auto dev-login if on localhost
onMounted(async () => {
  await captchaService.init();

  // AUTO DEV-LOGIN: Automatically login on localhost
  // SECURITY: This only works when:
  // 1. Frontend is on localhost (checked here)
  // 2. Backend GO_ENV != "production" (checked server-side in auth_dev.go)
  if (isLocalhost.value && !autoLoginAttempted.value) {
    autoLoginAttempted.value = true;
    autoLoginInProgress.value = true;

    try {
      // Use default tenant (yumna_bertigamart) for auto-login
      await handleDevLogin();
    } catch (err: any) {
      // If auto-login fails, just show the normal login form
      // User can still manually select tenant and click dev login button
      console.warn(
        "[Dev Auto-Login] Failed, falling back to manual:",
        err.message,
      );
      autoLoginInProgress.value = false;
    }
  }
});

// Dev login handler - calls secure backend endpoint
const handleDevLogin = async () => {
  try {
    isLoading.value = true;
    error.value = "";

    // Call backend dev-login endpoint (only works in dev mode on backend)
    const response = await api.post<{
      success: boolean;
      token: string;
      tenant_id: string;
      user: { id: string; username: string; email: string; role: string };
      dev_mode: boolean;
    }>("/auth/dev-login", {
      tenant_id: selectedDevTenant.value,
    });

    if (!response.success) {
      throw new Error("Dev login failed");
    }

    // Store user data
    localStorage.setItem("authToken", response.token);
    localStorage.setItem("userRole", response.user.role || "user");
    localStorage.setItem("userName", response.user.username || "User");
    localStorage.setItem("tenantId", response.tenant_id);

    authStore.setAuth({
      token: response.token,
      user: response.user,
      tenant_id: response.tenant_id,
    });

    // Redirect to home
    router.push(returnUrl.value);
  } catch (err: any) {
    error.value = err.message || "Dev login failed - is backend in dev mode?";
    // If auto-login was in progress, hide the overlay so user can see the form
    autoLoginInProgress.value = false;
  } finally {
    isLoading.value = false;
  }
};

const handleLogin = async () => {
  if (isLocked.value) return;

  try {
    isLoading.value = true;
    error.value = "";

    // Get reCAPTCHA token if required
    let recaptchaToken: string | undefined;
    if (requiresCaptcha.value) {
      const token = await captchaService.execute("login");
      if (!token) {
        error.value = "reCAPTCHA verification failed. Please try again.";
        return;
      }
      recaptchaToken = token;
    }

    const response = await authService.login(
      form.username,
      form.password,
      recaptchaToken,
    );

    // Store user data in localStorage for admin panel
    localStorage.setItem("authToken", response.token);
    localStorage.setItem("userRole", response.user.role || "user");
    localStorage.setItem("userName", response.user.username || "User");
    // Backend returns tenant_id (snake_case) per AGENTS.MD convention
    localStorage.setItem(
      "tenantId",
      response.tenant_id || response.user.username || "",
    );

    authStore.setAuth({
      token: response.token,
      user: response.user,
    });

    // Redirect to returnUrl or home page
    router.push(returnUrl.value);
  } catch (err: any) {
    error.value = err.message || "Login failed";

    // Update CAPTCHA/lock status from error response
    if (err.requiresCaptcha) {
      requiresCaptcha.value = true;
    }

    if (err.isLocked) {
      isLocked.value = true;
      lockMinutesRemaining.value = err.lockMinutesRemaining || 30;
    }
  } finally {
    isLoading.value = false;
  }
};
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  background: white;
  padding: 40px;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 100%;
  max-width: 400px;
}

h1 {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  color: #555;
  font-weight: 500;
}

input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.btn-submit {
  width: 100%;
  padding: 10px;
  background: #4338ca;
  color: #ffffff;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.3s;
}

.btn-submit:hover:not(:disabled) {
  background: #3730a3;
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  color: #dc3545;
  text-align: center;
  margin-top: 15px;
  font-size: 14px;
}

.locked-message {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 20px;
  text-align: center;
  font-size: 14px;
}

.session-expired-message {
  background: #fffbeb;
  border: 1px solid #fcd34d;
  color: #92400e;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 20px;
  text-align: center;
  font-size: 14px;
}

.captcha-notice {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  padding: 8px 12px;
  border-radius: 4px;
  margin-bottom: 15px;
  text-align: center;
  font-size: 13px;
}

.recaptcha-branding {
  margin-top: 20px;
  text-align: center;
  font-size: 11px;
  color: #6b7280;
}

.recaptcha-branding a {
  color: #4338ca;
  text-decoration: none;
}

.recaptcha-branding a:hover {
  text-decoration: underline;
}

input:disabled {
  background: #f3f4f6;
  cursor: not-allowed;
}

/* Dev Mode Styles */
.dev-mode-section {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  border: 2px solid #f59e0b;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.dev-mode-header {
  font-weight: 600;
  color: #92400e;
  margin-bottom: 12px;
  font-size: 14px;
}

.dev-mode-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.dev-mode-content label {
  color: #78350f;
  font-size: 13px;
  margin-bottom: 0;
}

.dev-tenant-select {
  width: 100%;
  padding: 10px;
  border: 1px solid #d97706;
  border-radius: 4px;
  font-size: 14px;
  background: white;
  color: #1f2937;
}

.dev-tenant-select:focus {
  outline: none;
  border-color: #b45309;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.2);
}

.btn-dev-login {
  width: 100%;
  padding: 10px;
  background: #d97706;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.3s;
}

.btn-dev-login:hover:not(:disabled) {
  background: #b45309;
}

.btn-dev-login:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Auto-login overlay styles */
.auto-login-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 20px;
  color: white;
  font-size: 18px;
  font-weight: 500;
}

.auto-login-spinner {
  width: 50px;
  height: 50px;
  border: 4px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
