<template>
  <!-- eslint-disable vue/no-reserved-component-names -->
  <div
    class="navbar bg-base-100 shadow-md border-b border-base-300 px-3 sm:px-4 md:px-6 py-2 sm:py-3 gap-2 sm:gap-4 flex-wrap sm:flex-nowrap"
  >
    <!-- Left Section: Logo -->
    <div class="flex-1 min-w-0">
      <a
        href="/"
        class="text-base sm:text-lg md:text-xl font-bold text-primary flex items-center gap-1 sm:gap-2 hover:opacity-80 transition-opacity whitespace-nowrap"
      >
        <i class="pi pi-shopping-cart text-lg sm:text-xl"></i>
        <span class="hidden xs:inline sm:inline text-sm sm:text-base"
          >Platform</span
        >
        <span class="hidden md:inline text-sm sm:text-base"
          >Ecommerce Platform</span
        >
      </a>
    </div>

    <!-- Right Section: Actions -->
    <div
      class="flex-none flex items-center gap-1 sm:gap-2 md:gap-3 flex-wrap justify-end"
    >
      <!-- Connection Status Badge -->
      <div
        :class="[
          'badge gap-1 sm:gap-2 px-2 sm:px-3 py-1.5 sm:py-2 text-xs sm:text-sm whitespace-nowrap',
          connectionStatus === 'connected'
            ? 'badge-success'
            : connectionStatus === 'error'
              ? 'badge-error'
              : connectionStatus === 'connecting'
                ? 'badge-warning'
                : 'badge-neutral',
        ]"
      >
        <span
          :class="[
            'inline-block w-1.5 sm:w-2 h-1.5 sm:h-2 rounded-full flex-shrink-0',
            connectionStatus === 'connected'
              ? 'bg-success'
              : connectionStatus === 'error'
                ? 'bg-error'
                : connectionStatus === 'connecting'
                  ? 'bg-warning'
                  : 'bg-neutral',
          ]"
        ></span>
        <span class="hidden sm:inline text-xs md:text-sm">{{
          connectionStatusText
        }}</span>
      </div>

      <!-- Theme Toggle Button -->
      <button
        @click="toggleThemeHandler"
        class="btn btn-ghost btn-xs sm:btn-sm md:btn-md"
        :title="isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
        :aria-label="isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
      >
        <span class="text-lg sm:text-xl md:text-2xl leading-none">{{
          isDark ? "☀️" : "🌙"
        }}</span>
      </button>

      <!-- Refresh Button -->
      <button
        @click="$emit('refresh-all')"
        :disabled="loading"
        class="btn btn-primary btn-xs sm:btn-sm md:btn-md gap-0 sm:gap-1 md:gap-2"
        aria-label="Refresh all data"
      >
        <i :class="['pi', loading ? 'pi-spin' : '', 'pi-refresh']"></i>
        <span class="hidden xs:inline text-xs sm:text-sm md:text-base"
          >Refresh</span
        >
      </button>
    </div>
  </div>
  <!-- eslint-enable -->
</template>

<script lang="ts">
import { computed, PropType, ComputedRef } from "vue";
import { useTheme } from "@/composables/useTheme";

export default {
  name: "DashboardHeader",
  props: {
    connectionStatus: {
      type: String as PropType<
        "connecting" | "connected" | "error" | "disconnected"
      >,
      default: "connecting",
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["refresh-all"],
  setup(props, { emit }) {
    const theme = useTheme();
    const isDark = computed(() => theme.isDarkMode.value);

    const connectionStatusText: ComputedRef<string> = computed(() => {
      const statusMap: Record<string, string> = {
        connecting: "Connecting...",
        connected: "Connected",
        error: "Disconnected",
      };
      return statusMap[props.connectionStatus] || "Unknown";
    });

    const toggleThemeHandler = () => {
      theme.toggleTheme();
      // Force re-render
      emit("refresh-all");
    };

    return {
      connectionStatusText,
      theme,
      isDark,
      toggleThemeHandler,
    };
  },
};
</script>

<style scoped>
/* All styling handled by Daisy UI classes */
</style>
