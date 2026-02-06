<template>
  <div
    id="app"
    class="min-h-screen w-screen overflow-x-hidden bg-white text-gray-900 transition-colors duration-200 dark:bg-gray-950 dark:text-gray-50"
  >
    <!-- Skip to main content link for keyboard users -->
    <a href="#main-content" class="skip-link">Skip to main content</a>

    <Navbar />
    <Toast ref="toastRef" />

    <!-- Page transitions -->
    <router-view v-slot="{ Component, route }">
      <transition name="page-fade" mode="out-in">
        <main id="main-content" :key="route.path" class="main-with-bottom-nav">
          <component :is="Component" />
        </main>
      </transition>
    </router-view>

    <!-- Mobile Bottom Navigation -->
    <MobileBottomNav />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeMount, onBeforeUnmount, watch } from "vue";
import Navbar from "@components/Navbar.vue";
import Toast from "@components/Toast.vue";
import MobileBottomNav from "@components/layout/MobileBottomNav.vue";
import { useToast } from "@/composables/useToast";
import { useTheme } from "@/composables/useTheme";
import { getApiBaseUrl } from "@/utils/apiHeaders";

const toastRef = ref();
const { setInstance } = useToast();
const { isDarkMode } = useTheme();

let darkModeQueryListener: ((e: MediaQueryListEvent) => void) | null = null;
let darkModeQuery: MediaQueryList | null = null;

// Initialize theme BEFORE component mounts to prevent CLS
const initializeTheme = () => {
  const html = document.documentElement;
  const storedTheme = localStorage.getItem("theme");

  if (storedTheme) {
    isDarkMode.value = storedTheme === "dark";
  } else {
    const prefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)"
    ).matches;
    isDarkMode.value = prefersDark;
  }

  if (isDarkMode.value) {
    html.classList.add("dark");
  } else {
    html.classList.remove("dark");
  }
};

onBeforeMount(() => {
  initializeTheme();
});

watch(
  isDarkMode,
  (newVal) => {
    const html = document.documentElement;
    if (newVal) {
      html.classList.add("dark");
    } else {
      html.classList.remove("dark");
    }
  },
  { immediate: true }
);

onMounted(() => {
  if (toastRef.value) {
    setInstance(toastRef.value);
  }

  // Fetch CSRF token on app initialization
  fetchCSRFToken();

  darkModeQuery = window.matchMedia("(prefers-color-scheme: dark)");
  darkModeQueryListener = (e) => {
    const storedTheme = localStorage.getItem("theme");
    if (!storedTheme) {
      isDarkMode.value = e.matches;
    }
  };

  darkModeQuery.addEventListener("change", darkModeQueryListener);
});

/**
 * Fetch CSRF token from backend
 * This sets the csrf_token cookie for subsequent mutating requests
 */
async function fetchCSRFToken() {
  try {
    const url = getApiBaseUrl("/csrf-token");
    await fetch(url, {
      method: "GET",
      credentials: "include", // Important: include cookies
    });
  } catch (error) {
    console.warn("Failed to fetch CSRF token:", error);
  }
}

onBeforeUnmount(() => {
  if (darkModeQuery && darkModeQueryListener) {
    darkModeQuery.removeEventListener("change", darkModeQueryListener);
  }
  darkModeQueryListener = null;
  darkModeQuery = null;
});
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  scroll-behavior: smooth;
  font-size: 16px;
}

body {
  font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
  overflow-x: hidden;
}

#app {
  min-height: 100vh;
}

/* Add padding for mobile bottom nav */
@media (max-width: 768px) {
  .main-with-bottom-nav {
    padding-bottom: calc(56px + env(safe-area-inset-bottom, 0px));
  }
}

/* Skip Link for Accessibility */
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  padding: 0.5rem 1rem;
  background: var(--color-primary-600, #2563eb);
  color: white;
  font-weight: 600;
  z-index: 9999;
  transition: top 0.2s ease;
  border-radius: 0 0 0.5rem 0;
}
.skip-link:focus {
  top: 0;
  outline: 2px solid white;
  outline-offset: 2px;
}

/* Page Transition - Fade */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 200ms ease;
}
.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}

/* Reduced Motion - Disable page transitions */
@media (prefers-reduced-motion: reduce) {
  .page-fade-enter-active,
  .page-fade-leave-active {
    transition: none;
  }
}

/* Optimize transitions for composited elements */
:deep(.btn),
:deep([role="button"]),
:deep(.badge),
:deep(.modal) {
  /* Use only opacity and transform for smooth animations */
  transition-property: opacity, transform;
  transition-duration: 0.2s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  will-change: auto;
}

/* Disable will-change by default to save memory */
:deep(.transition-all) {
  transition-duration: 0.3s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
