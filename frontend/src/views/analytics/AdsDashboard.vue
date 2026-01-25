<template>
  <div class="min-h-screen bg-gray-50 pb-12">
    <!-- Header -->
    <header class="bg-white shadow z-10 sticky top-0">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div
          class="flex flex-col md:flex-row md:items-center md:justify-between gap-4"
        >
          <div class="flex items-center space-x-4">
            <h1 class="text-2xl font-bold text-gray-900">Ads Analytics</h1>

            <!-- Platform Switcher -->
            <div class="bg-gray-100 p-1 rounded-lg inline-flex">
              <button
                @click="store.setPlatform('shopee')"
                class="px-3 py-1 text-sm font-medium rounded-md transition-colors"
                :class="
                  store.platform === 'shopee'
                    ? 'bg-white shadow text-orange-600'
                    : 'text-gray-500 hover:text-gray-700'
                "
              >
                Shopee
              </button>
              <button
                @click="store.setPlatform('tiktok')"
                class="px-3 py-1 text-sm font-medium rounded-md transition-colors"
                :class="
                  store.platform === 'tiktok'
                    ? 'bg-white shadow text-black'
                    : 'text-gray-500 hover:text-gray-700'
                "
              >
                TikTok
              </button>
            </div>
          </div>

          <div class="flex items-center space-x-3">
            <!-- Simplified Date Picker (Last 30 days fixed for MVP or could be added later) -->
            <div class="text-sm text-gray-500 bg-gray-100 px-3 py-2 rounded">
              {{ formatDate(store.startDate) }} -
              {{ formatDate(store.endDate) }}
            </div>

            <button
              @click="showUploadModal = true"
              class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none"
            >
              <svg
                class="h-4 w-4 mr-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
                />
              </svg>
              Upload Report
            </button>
          </div>
        </div>
      </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <!-- Loading State -->
      <div
        v-if="store.loading && !store.summary"
        class="flex justify-center py-20"
      >
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
        ></div>
      </div>

      <!-- Error State -->
      <div v-if="store.error" class="bg-red-50 border-l-4 border-red-500 p-4">
        <div class="flex">
          <div class="flex-shrink-0">
            <svg
              class="h-5 w-5 text-red-400"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fill-rule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                clip-rule="evenodd"
              />
            </svg>
          </div>
          <div class="ml-3">
            <p class="text-sm text-red-700">{{ store.error }}</p>
          </div>
        </div>
      </div>

      <template v-if="store.summary">
        <!-- KPI Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            title="Total Ad Spend"
            :value="store.summary.totalCost"
            type="currency"
            sub-label="Avg ACOS"
            :sub-value="store.summary.avgAcos.toFixed(2) + '%'"
          />
          <StatCard
            title="Total GMV (Omzet)"
            :value="store.summary.totalRevenue"
            type="currency"
            sub-label="Total Orders"
            :sub-value="store.summary.totalConversions"
            trend="10"
          />
          <StatCard
            title="ROAS"
            :value="store.summary.avgRoas"
            type="percentage"
            sub-label="CTR"
            :sub-value="store.summary.avgCtr.toFixed(2) + '%'"
            :trend="store.summary.avgRoas > 5 ? 5 : -2"
          />
          <!-- Placeholder Health Score (Calculate in frontend or fetch if backend provides it aggregated) -->
          <StatCard
            title="Overall Health Score"
            value="85"
            type="number"
            sub-label="Active Products"
            :sub-value="store.summary.productCount"
          />
        </div>

        <!-- Charts -->
        <AdsTrendChart :data="store.trends" :loading="store.loading" />

        <!-- Detailed Table -->
        <AdsPerformanceTable :data="store.products" />
      </template>

      <div
        v-if="!store.summary && !store.loading"
        class="text-center py-20 bg-white rounded-lg border border-dashed border-gray-300"
      >
        <svg
          class="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">
          No Analytics Data
        </h3>
        <p class="mt-1 text-sm text-gray-500">
          Get started by uploading your latest ads report.
        </p>
        <div class="mt-6">
          <button
            @click="showUploadModal = true"
            class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
          >
            Upload Report
          </button>
        </div>
      </div>
    </main>

    <AdsUploadModal
      :is-open="showUploadModal"
      :platform-name="store.platform === 'shopee' ? 'Shopee Ads' : 'TikTok Ads'"
      @close="showUploadModal = false"
      @uploaded="handleUploadSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useAnalyticsStore } from "../../stores/analytics";
import StatCard from "../../components/analytics/StatCard.vue";
import AdsTrendChart from "../../components/analytics/AdsTrendChart.vue";
import AdsPerformanceTable from "../../components/analytics/AdsPerformanceTable.vue";
import AdsUploadModal from "../../components/analytics/AdsUploadModal.vue";

const store = useAnalyticsStore();
const showUploadModal = ref(false);

onMounted(() => {
  store.fetchAll();
});

function handleUploadSuccess() {
  // Toast notification could go here
  console.log("Upload successful");
}

function formatDate(date: Date) {
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "short",
  }).format(date);
}
</script>
