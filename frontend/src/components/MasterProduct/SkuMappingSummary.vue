<template>
  <div class="mapping-summary">
    <div
      class="summary-item"
      v-for="(stats, platform) in platforms"
      :key="platform"
    >
      <span class="platform-icon">{{
        getPlatformIcon(platform as string)
      }}</span>
      <span class="platform-name">{{ platform }}</span>
      <span class="linked-count">{{ stats.linked }}/{{ stats.total }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { MappingStatus } from "@/services/masterProductService";

interface Props {
  platforms: MappingStatus["platforms"];
}
defineProps<Props>();

const getPlatformIcon = (platform: string): string => {
  const icons: Record<string, string> = {
    shopee: "🟠",
    tiktok: "⬛",
    lazada: "🔵",
  };
  return icons[platform] || "⚪";
};
</script>

<style scoped>
.mapping-summary {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.5rem;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.platform-icon {
  font-size: 1.25rem;
}

.platform-name {
  font-weight: 600;
  text-transform: capitalize;
}

.linked-count {
  color: #6b7280;
  font-size: 0.875rem;
}
</style>
