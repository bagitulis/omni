<template>
  <div :class="['platform-badge', sizeClass, statusClass]">
    <img
      :src="iconSrc"
      :alt="platformAltText"
      class="badge-icon"
      :style="{ width: sizePixels, height: sizePixels }"
    />
    <div v-if="showLabel" class="badge-label" aria-hidden="true">
      {{ platform.charAt(0).toUpperCase() + platform.slice(1) }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { getPlatformIconSvg } from "@/utils/platformIcons";

const props = defineProps({
  platform: {
    type: String,
    required: true,
  },
  size: {
    type: String,
    default: "lg", // xs, sm, base, md, lg, xl, 2xl, 3xl
  },
  status: {
    type: String,
    default: null, // valid, expired, warning
  },
  showLabel: {
    type: Boolean,
    default: false,
  },
});

const sizeMap: Record<string, number> = {
  "2xs": 12,
  xs: 16,
  sm: 20,
  base: 24,
  md: 32,
  lg: 48,
  xl: 64,
  "2xl": 80,
  "3xl": 96,
};

const iconSrc = computed(() => getPlatformIconSvg(props.platform));

const platformAltText = computed(() => {
  // Use descriptive alt text to avoid redundant-alt with adjacent label
  return props.showLabel
    ? ""
    : `${props.platform.charAt(0).toUpperCase() + props.platform.slice(1)} platform icon`;
});

const sizeClass = computed(() => `size-${props.size}`);

const statusClass = computed(() =>
  props.status ? `status-${props.status}` : ""
);

const sizePixels = computed(() => `${sizeMap[props.size]}px`);
</script>

<style scoped>
.platform-badge {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px;
  border-radius: 8px;
  background-color: #f0f4f8;
  transition: all 0.3s ease;
}

.badge-icon {
  object-fit: contain;
  flex-shrink: 0;
}

.badge-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #2c3e50;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Size Classes */
.size-2xs {
  padding: 4px;
}

.size-xs {
  padding: 6px;
}

.size-sm {
  padding: 8px;
}

.size-base {
  padding: 10px;
}

.size-md {
  padding: 12px;
}

.size-lg {
  padding: 16px;
}

.size-xl {
  padding: 20px;
}

.size-2xl {
  padding: 24px;
}

.size-3xl {
  padding: 28px;
}

/* Status Classes */
.status-valid {
  border: 2px solid #27ae60;
  background-color: #f0fdf4;
  box-shadow: 0 2px 8px rgba(39, 174, 96, 0.15);
}

.status-valid:hover {
  box-shadow: 0 4px 16px rgba(39, 174, 96, 0.25);
}

.status-expired {
  border: 2px solid #e74c3c;
  background-color: #fef2f2;
  box-shadow: 0 2px 8px rgba(231, 76, 60, 0.15);
}

.status-expired:hover {
  box-shadow: 0 4px 16px rgba(231, 76, 60, 0.25);
}

.status-warning {
  border: 2px solid #f39c12;
  background-color: #fffbf0;
  box-shadow: 0 2px 8px rgba(243, 156, 18, 0.15);
}

.status-warning:hover {
  box-shadow: 0 4px 16px rgba(243, 156, 18, 0.25);
}

/* Hover Effects */
.platform-badge:hover {
  transform: translateY(-2px);
}

/* Responsive */
@media (max-width: 768px) {
  .size-lg {
    padding: 12px;
  }

  .size-xl {
    padding: 16px;
  }

  .size-2xl {
    padding: 20px;
  }
}
</style>
