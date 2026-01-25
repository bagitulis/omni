<template>
  <div
    class="footer footer-center bg-base-200 text-base-content border-t border-base-300 py-3 px-4 md:py-4 md:px-6"
  >
    <div
      class="flex flex-col md:flex-row justify-between items-center w-full gap-3 md:gap-4"
    >
      <!-- Left Section: Last Updated & Status -->
      <div class="flex flex-col sm:flex-row gap-4 sm:gap-6 flex-1">
        <!-- Last Updated -->
        <div
          class="flex items-center gap-2 text-xs sm:text-sm font-medium whitespace-nowrap"
        >
          <i
            class="pi pi-clock text-info text-base sm:text-lg flex-shrink-0"
          ></i>
          <span class="text-base-content hidden sm:inline">Last updated:</span>
          <span class="font-semibold text-info text-xs sm:text-sm">{{
            lastUpdated
          }}</span>
        </div>

        <!-- Status -->
        <div
          :class="[
            'flex items-center gap-2 text-xs sm:text-sm font-medium whitespace-nowrap',
            loading ? 'text-warning' : 'text-success',
          ]"
        >
          <i
            :class="loading ? 'pi pi-spin pi-spinner' : 'pi pi-check-circle'"
            class="flex-shrink-0 text-base sm:text-lg"
          ></i>
          <span>{{ loading ? "Running..." : "Ready" }}</span>
        </div>
      </div>

      <!-- Divider (hidden on mobile) -->
      <div
        class="divider divider-horizontal m-0 hidden sm:flex flex-shrink-0"
      ></div>

      <!-- Right Section: Stats (No badges, just numbers) -->
      <div class="flex gap-4 sm:gap-6 justify-end">
        <!-- Requests Counter -->
        <div class="flex items-center gap-1.5 sm:gap-2 text-xs sm:text-sm">
          <i
            class="pi pi-arrow-right text-primary text-base sm:text-lg flex-shrink-0"
          ></i>
          <span class="text-base-content text-xs hidden sm:inline"
            >Requests:</span
          >
          <span class="font-bold text-primary text-sm sm:text-base">{{
            requestCount
          }}</span>
        </div>

        <!-- Errors Counter -->
        <div class="flex items-center gap-1.5 sm:gap-2 text-xs sm:text-sm">
          <i
            class="pi pi-exclamation-circle text-error text-base sm:text-lg flex-shrink-0"
          ></i>
          <span class="text-base-content text-xs hidden sm:inline"
            >Errors:</span
          >
          <span class="font-bold text-error text-sm sm:text-base">{{
            errorCount
          }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
export default {
  name: "DashboardStatusBar",
  props: {
    lastUpdated: {
      type: String,
      default: "",
    },
    loading: {
      type: Boolean,
      default: false,
    },
    requestCount: {
      type: Number,
      default: 0,
    },
    errorCount: {
      type: Number,
      default: 0,
    },
  },
};
</script>

<style scoped>
/* Responsive adjustments for small screens */
@media (max-width: 768px) {
  .footer {
    padding: 1rem !important;
  }

  .divider {
    display: none !important;
  }
}
</style>
