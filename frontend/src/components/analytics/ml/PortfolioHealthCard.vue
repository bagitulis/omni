<template>
  <div class="health-card">
    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="skeleton-gauge"></div>
      <div class="skeleton-text"></div>
    </div>

    <!-- No Data State -->
    <div v-else-if="!health || health.total_products === 0" class="empty-state">
      <Icon name="chart-bar" size="xl" class="empty-icon" />
      <p class="empty-title">No portfolio data available</p>
      <p class="empty-hint">
        Upload TikTok Ads data first via Analytics > TikTok Ads
      </p>
    </div>

    <!-- Health Content -->
    <template v-else>
      <!-- Health Score Gauge -->
      <div class="score-section">
        <div class="gauge-container">
          <svg viewBox="0 0 100 60" class="gauge-svg">
            <!-- Background arc -->
            <path
              d="M 10 50 A 40 40 0 0 1 90 50"
              fill="none"
              stroke="#e5e7eb"
              stroke-width="8"
              stroke-linecap="round"
            />
            <!-- Value arc -->
            <path
              :d="gaugeArc"
              fill="none"
              :stroke="gaugeColor"
              stroke-width="8"
              stroke-linecap="round"
            />
          </svg>
          <div class="gauge-value">
            <span class="score-number">{{
              Math.round(health.health_score)
            }}</span>
            <span class="score-label">{{ health.health_label }}</span>
          </div>
        </div>
      </div>

      <!-- Category Distribution -->
      <div class="category-section">
        <h4>Product Categories</h4>
        <div class="category-bars">
          <div class="category-bar" v-for="cat in categories" :key="cat.name">
            <div class="bar-label">
              <span class="cat-icon" :style="{ color: cat.color }">{{
                cat.icon
              }}</span>
              <span>{{ cat.name }}</span>
            </div>
            <div class="bar-track">
              <div
                class="bar-fill"
                :style="{
                  width: catPercent(cat.count) + '%',
                  backgroundColor: cat.color,
                }"
              ></div>
            </div>
            <span class="bar-count">{{ cat.count }}</span>
          </div>
        </div>
      </div>

      <!-- Financial Summary -->
      <div class="financial-section">
        <div class="fin-item">
          <span class="fin-label">Total Cost</span>
          <span class="fin-value cost">{{
            formatCurrency(health.total_cost)
          }}</span>
        </div>
        <div class="fin-item">
          <span class="fin-label">Total Revenue</span>
          <span class="fin-value revenue">{{
            formatCurrency(health.total_revenue)
          }}</span>
        </div>
        <div class="fin-item">
          <span class="fin-label">ROAS</span>
          <span class="fin-value" :class="roasClass"
            >{{ health.overall_roas.toFixed(2) }}x</span
          >
        </div>
      </div>

      <!-- Alerts Badge -->
      <div v-if="health.active_alerts > 0" class="alerts-badge">
        <Icon name="warning" size="sm" />
        <span
          >{{ health.active_alerts }} Active Alert{{
            health.active_alerts > 1 ? "s" : ""
          }}</span
        >
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { PortfolioHealth } from "@/composables/useMLAnalytics";
import Icon from "@/components/ui/Icon.vue";

const props = defineProps<{
  health: PortfolioHealth | null;
  loading: boolean;
  formatCurrency: (val: number) => string;
}>();

const categories = computed(() => {
  if (!props.health) return [];
  return [
    {
      name: "Star",
      count: props.health.star_count,
      color: "#10b981",
      icon: "★",
    },
    {
      name: "Growth",
      count: props.health.growth_count,
      color: "#3b82f6",
      icon: "↑",
    },
    {
      name: "Stable",
      count: props.health.stable_count,
      color: "#6b7280",
      icon: "―",
    },
    {
      name: "Watch",
      count: props.health.watch_count,
      color: "#f59e0b",
      icon: "!",
    },
    {
      name: "Problem",
      count: props.health.problem_count,
      color: "#ef4444",
      icon: "✕",
    },
  ];
});

const catPercent = (count: number) => {
  if (!props.health || props.health.total_products === 0) return 0;
  return (count / props.health.total_products) * 100;
};

const gaugeArc = computed(() => {
  if (!props.health) return "";
  const score = Math.min(Math.max(props.health.health_score, 0), 100);
  const angle = (score / 100) * 180;
  const rad = (angle * Math.PI) / 180;
  const x = 50 + 40 * Math.cos(Math.PI - rad);
  const y = 50 - 40 * Math.sin(Math.PI - rad);
  const largeArc = angle > 90 ? 1 : 0;
  return `M 10 50 A 40 40 0 ${largeArc} 1 ${x} ${y}`;
});

const gaugeColor = computed(() => {
  if (!props.health) return "#e5e7eb";
  const score = props.health.health_score;
  if (score >= 70) return "#10b981";
  if (score >= 55) return "#3b82f6";
  if (score >= 40) return "#f59e0b";
  return "#ef4444";
});

const roasClass = computed(() => {
  if (!props.health) return "";
  const roas = props.health.overall_roas;
  if (roas >= 3) return "roas-excellent";
  if (roas >= 2) return "roas-good";
  if (roas >= 1) return "roas-ok";
  return "roas-poor";
});
</script>

<style scoped>
.health-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 40px;
}

.skeleton-gauge {
  width: 120px;
  height: 60px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
  border-radius: 60px 60px 0 0;
}

.skeleton-text {
  width: 80px;
  height: 20px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
  border-radius: 4px;
}

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.empty-icon {
  font-size: 48px;
  display: block;
  margin-bottom: 8px;
}

.empty-icon :deep(.icon) {
  width: 48px;
  height: 48px;
  color: #6b7280;
}

.empty-title {
  font-size: 16px;
  font-weight: 500;
  color: #374151;
  margin: 0 0 8px;
}

.empty-hint {
  font-size: 13px;
  color: #9ca3af;
  margin: 0;
}

.score-section {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.gauge-container {
  position: relative;
  width: 160px;
  height: 90px;
}

.gauge-svg {
  width: 100%;
  height: 100%;
}

.gauge-value {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  text-align: center;
}

.score-number {
  font-size: 32px;
  font-weight: 700;
  color: #1f2937;
  display: block;
}

.score-label {
  font-size: 14px;
  color: #6b7280;
}

.category-section {
  margin-bottom: 20px;
}

.category-section h4 {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin: 0 0 12px 0;
}

.category-bars {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.category-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bar-label {
  width: 80px;
  font-size: 13px;
  color: #4b5563;
  display: flex;
  align-items: center;
  gap: 4px;
}

.cat-icon {
  font-weight: bold;
}

.bar-track {
  flex: 1;
  height: 8px;
  background: #f3f4f6;
  border-radius: 4px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease;
}

.bar-count {
  width: 30px;
  text-align: right;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}

.financial-section {
  display: flex;
  justify-content: space-between;
  padding: 16px 0;
  border-top: 1px solid #e5e7eb;
}

.fin-item {
  text-align: center;
}

.fin-label {
  display: block;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 4px;
}

.fin-value {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.fin-value.cost {
  color: #dc2626;
}

.fin-value.revenue {
  color: #059669;
}

.roas-excellent {
  color: #059669;
}
.roas-good {
  color: #3b82f6;
}
.roas-ok {
  color: #f59e0b;
}
.roas-poor {
  color: #dc2626;
}

.alerts-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 16px;
  background: #fef3c7;
  border-radius: 8px;
  color: #92400e;
  font-size: 14px;
  font-weight: 500;
  margin-top: 16px;
}

.alerts-badge :deep(.icon) {
  width: 16px;
  height: 16px;
  color: #d97706;
}
</style>
