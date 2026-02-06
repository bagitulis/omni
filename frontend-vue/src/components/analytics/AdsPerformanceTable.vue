<template>
  <div
    class="bg-white rounded-lg shadow-sm border border-gray-100 overflow-hidden"
  >
    <div
      class="p-4 border-b border-gray-100 flex justify-between items-center bg-gray-50"
    >
      <h3 class="font-semibold text-gray-700">Detailed Performance</h3>
      <div class="text-xs text-gray-500">
        Top {{ sortedData.length }} Products
      </div>
    </div>

    <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th
              scope="col"
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
              @click="sortBy('product_name')"
            >
              Product
            </th>
            <th
              scope="col"
              class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
              @click="sortBy('spend')"
            >
              Spend
              <span v-if="sortKey === 'spend'">{{ sortDesc ? "↓" : "↑" }}</span>
            </th>
            <th
              scope="col"
              class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
              @click="sortBy('gmv')"
            >
              GMV
              <span v-if="sortKey === 'gmv'">{{ sortDesc ? "↓" : "↑" }}</span>
            </th>
            <th
              scope="col"
              class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
              @click="sortBy('roas')"
            >
              ROAS
              <span v-if="sortKey === 'roas'">{{ sortDesc ? "↓" : "↑" }}</span>
            </th>
            <th
              scope="col"
              class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
              @click="sortBy('score.composite_score')"
            >
              Score
              <span v-if="sortKey === 'score.composite_score'">{{
                sortDesc ? "↓" : "↑"
              }}</span>
            </th>
            <th
              scope="col"
              class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider"
            >
              Action
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="data.length === 0">
            <td colspan="6" class="px-6 py-4 text-center text-sm text-gray-500">
              No data found.
            </td>
          </tr>
          <tr
            v-for="item in sortedData"
            :key="item.product_id"
            class="hover:bg-gray-50 transition-colors"
          >
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <div class="ml-0">
                  <div
                    class="text-sm font-medium text-gray-900 truncate max-w-xs"
                    :title="item.product_name"
                  >
                    {{ item.product_name }}
                  </div>
                  <div class="text-xs text-gray-500">{{ item.product_id }}</div>
                </div>
              </div>
            </td>
            <td
              class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500"
            >
              {{ formatCurrency(item.spend) }}
            </td>
            <td
              class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900 font-medium"
            >
              {{ formatCurrency(item.gmv) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
              <span :class="getRoasClass(item.roas)"
                >{{ item.roas.toFixed(2) }}x</span
              >
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-center">
              <div
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                :class="getScoreClass(item.score.composite_score)"
              >
                {{ item.score.composite_score }}
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-center text-sm">
              <span
                class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
                :class="getActionClass(item.score.category)"
              >
                {{ item.score.category }}
              </span>
              <!-- Tooltip logic could go here for item.score.action -->
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, defineProps } from "vue";
import type { ProductPerformance } from "../../stores/analytics";

const props = defineProps<{
  data: ProductPerformance[];
}>();

const sortKey = ref("score.composite_score"); // Default sort by score
const sortDesc = ref(true);

const sortedData = computed(() => {
  return [...props.data].sort((a: any, b: any) => {
    // Handle nested keys like 'score.compositeScore'
    const getVal = (obj: any, key: string) =>
      key.split(".").reduce((o, k) => (o || {})[k], obj);

    const valA = getVal(a, sortKey.value);
    const valB = getVal(b, sortKey.value);

    if (valA < valB) return sortDesc.value ? 1 : -1;
    if (valA > valB) return sortDesc.value ? -1 : 1;
    return 0;
  });
});

function sortBy(key: string) {
  if (sortKey.value === key) {
    sortDesc.value = !sortDesc.value;
  } else {
    sortKey.value = key;
    sortDesc.value = true; // Default desc for new columns (usually looking for high numbers)
  }
}

function formatCurrency(val: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(val);
}

function getRoasClass(roas: number) {
  if (roas >= 10) return "text-green-600 font-bold";
  if (roas >= 5) return "text-green-600";
  if (roas >= 2) return "text-gray-900";
  return "text-red-500 font-bold";
}

function getScoreClass(score: number) {
  if (score >= 80) return "bg-green-100 text-green-800";
  if (score >= 50) return "bg-yellow-100 text-yellow-800";
  return "bg-red-100 text-red-800";
}

function getActionClass(category: string) {
  if (category === "SCALE_UP" || category.includes("LANJUTKAN"))
    return "bg-green-100 text-green-800";
  if (category === "STOP" || category.includes("HENTIKAN"))
    return "bg-red-100 text-red-800";
  if (category === "MONITOR" || category.includes("PANTAU"))
    return "bg-yellow-100 text-yellow-800";
  return "bg-gray-100 text-gray-800";
}
</script>
