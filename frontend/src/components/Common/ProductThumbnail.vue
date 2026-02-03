<template>
  <div class="product-thumbnail" :class="size">
    <!-- Actual image -->
    <img
      v-if="!hasError && imageUrl"
      :src="imageUrl"
      :alt="alt"
      class="product-img"
      loading="lazy"
      referrerpolicy="no-referrer"
      @error="handleImageError"
      @load="handleImageLoad"
    />

    <!-- Loading skeleton -->
    <div v-if="isLoading" class="skeleton-loader">
      <div class="skeleton-pulse"></div>
    </div>

    <!-- Fallback placeholder -->
    <div v-else-if="hasError || !imageUrl" class="image-placeholder">
      <Icon name="image" size="lg" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import Icon from "@/components/ui/Icon.vue";

interface Props {
  src?: string;
  alt: string;
  platform?: string;
  size?: "small" | "medium" | "large";
}

const props = withDefaults(defineProps<Props>(), {
  platform: "",
  size: "medium",
});

const isLoading = ref(true);
const hasError = ref(false);

const imageUrl = computed(() => {
  if (!props.src) return null;

  let url = props.src;

  // Shopee CDN optimization
  if (props.platform === "shopee" && url.includes("cf.shopee")) {
    // If it already has _tn, leave it
    if (url.includes("_tn")) return url;

    // If it has an extension like .jpg, insert _tn before it
    if (/\.(jpg|jpeg|png|webp)$/i.test(url)) {
      return url.replace(/(\.[^.]+)$/, "_tn$1");
    }

    // Otherwise append _tn
    return `${url}_tn`;
  }

  // Lazada CDN optimization
  if (props.platform === "lazada" && url.includes("lazada")) {
    if (!url.includes("_200x200")) {
      // Try to insert before extension if present
      if (/\.(jpg|jpeg|png|webp)$/i.test(url)) {
        return url.replace(/(\.[^.]+)$/, "_200x200$1");
      }
      return `${url}_200x200`;
    }
  }

  return url;
});

const handleImageError = () => {
  isLoading.value = false;
  hasError.value = true;
};

const handleImageLoad = () => {
  isLoading.value = false;
  hasError.value = false;
};

watch(
  () => props.src,
  (value) => {
    isLoading.value = Boolean(value);
    hasError.value = false;
  },
  { immediate: true },
);
</script>

<style scoped>
.product-thumbnail {
  position: relative;
  overflow: hidden;
  border-radius: 6px;
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-thumbnail.small {
  width: 40px;
  height: 40px;
}
.product-thumbnail.medium {
  width: 60px;
  height: 60px;
}
.product-thumbnail.large {
  width: 80px;
  height: 80px;
}

.product-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: opacity 0.2s ease;
  position: relative;
  z-index: 0;
}

.skeleton-loader {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: loading 1.5s ease-in-out infinite;
  z-index: 1;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #bdbdbd;
  font-size: 1.25rem;
  position: absolute;
  inset: 0;
}

@keyframes loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>
