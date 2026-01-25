<template>
  <div
    class="bg-white p-4 rounded-lg shadow-sm border border-gray-100 flex flex-col h-full"
  >
    <div class="text-sm text-gray-500 mb-1">{{ title }}</div>
    <div class="flex items-end justify-between mt-auto">
      <div>
        <div class="text-2xl font-bold text-gray-800">{{ formattedValue }}</div>
        <div v-if="subValue" class="text-xs text-gray-400 mt-1">
          {{ subLabel }}: {{ subValue }}
        </div>
      </div>
      <div v-if="trend" class="flex items-center text-xs" :class="trendClass">
        <span class="mr-1">{{ trendIcon }}</span>
        <span>{{ trend }}%</span>
      </div>
    </div>

    <!-- Optional Visual Indicator (like a progress bar for Health Score) -->
    <div v-if="isHealth" class="w-full bg-gray-100 rounded-full h-1.5 mt-3">
      <div
        class="h-1.5 rounded-full"
        :class="healthColorClass"
        :style="{ width: `${value}%` }"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps({
  title: String,
  value: [String, Number],
  type: {
    type: String,
    default: "text", // currency, number, percentage, score
  },
  subLabel: String,
  subValue: [String, Number],
  trend: Number,
  trendInverse: {
    type: Boolean,
    default: false,
  },
});

const isHealth = computed(() => props.title?.toLowerCase().includes("score"));

const formattedValue = computed(() => {
  if (props.value === undefined || props.value === null) return "-";

  if (props.type === "currency") {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0,
    }).format(Number(props.value));
  }
  if (props.type === "percentage") {
    return Number(props.value).toFixed(2) + "x"; // e.g. ROAS 5.0x
  }
  if (props.type === "percent_sign") {
    return Number(props.value).toFixed(2) + "%";
  }
  if (props.type === "number") {
    return new Intl.NumberFormat("id-ID").format(Number(props.value));
  }
  return props.value;
});

const trendClass = computed(() => {
  if (!props.trend) return "text-gray-400";
  if (props.trend > 0)
    return props.trendInverse ? "text-red-500" : "text-green-500";
  return props.trendInverse ? "text-green-500" : "text-red-500";
});

const trendIcon = computed(() => {
  if (!props.trend) return "-";
  return props.trend > 0 ? "▲" : "▼";
});

const healthColorClass = computed(() => {
  const val = Number(props.value);
  if (val >= 80) return "bg-green-500";
  if (val >= 50) return "bg-yellow-400";
  return "bg-red-500";
});
</script>
