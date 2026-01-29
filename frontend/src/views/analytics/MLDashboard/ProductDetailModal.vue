<template>
  <div v-if="product" class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>{{ product.product_name || product.product_id }}</h2>
        <button @click="$emit('close')" class="btn-close" type="button">
          ✕
        </button>
      </div>
      <div class="modal-body">
        <div class="detail-grid">
          <!-- Score Section -->
          <div class="detail-section">
            <h4>Unified Score</h4>
            <div class="score-large" :class="scoreClass(product.unified_score)">
              {{ product.unified_score.toFixed(0) }}
            </div>
            <div class="score-breakdown">
              <div class="breakdown-item">
                <span>ROAS Score</span>
                <span>{{ product.roas_score.toFixed(0) }}</span>
              </div>
              <div class="breakdown-item">
                <span>Trend Score</span>
                <span>{{ product.trend_score.toFixed(0) }}</span>
              </div>
              <div class="breakdown-item">
                <span>Momentum Score</span>
                <span>{{ product.momentum_score.toFixed(0) }}</span>
              </div>
              <div class="breakdown-item">
                <span>Volatility Score</span>
                <span>{{ product.volatility_score.toFixed(0) }}</span>
              </div>
            </div>
          </div>

          <!-- Financial Section -->
          <div class="detail-section">
            <h4>Financial Metrics</h4>
            <div class="metric-row">
              <span>Total Cost</span>
              <span class="cost">{{ formatCurrency(product.total_cost) }}</span>
            </div>
            <div class="metric-row">
              <span>Total Revenue</span>
              <span class="revenue">{{
                formatCurrency(product.total_revenue)
              }}</span>
            </div>
            <div class="metric-row">
              <span>Total Profit</span>
              <span :class="product.total_profit >= 0 ? 'revenue' : 'cost'">
                {{ formatCurrency(product.total_profit) }}
              </span>
            </div>
            <div class="metric-row">
              <span>ROAS</span>
              <span class="roas">{{ product.roas.toFixed(2) }}x</span>
            </div>
          </div>

          <!-- Recommendation Section -->
          <div class="detail-section full-width">
            <h4>Recommendation</h4>
            <div class="recommendation-box">
              <ActionBadge :action="product.action" />
              <p class="rec-label">{{ product.action_label }}</p>
              <p class="rec-budget" v-if="product.budget_change_pct !== 0">
                Budget Change:
                <strong
                  :class="
                    product.budget_change_pct > 0 ? 'positive' : 'negative'
                  "
                >
                  {{ product.budget_change_pct > 0 ? "+" : ""
                  }}{{ product.budget_change_pct }}%
                </strong>
              </p>
              <p class="rec-confidence">
                Confidence:
                <strong>{{ product.confidence_level }}</strong>
                ({{ (product.success_probability * 100).toFixed(0) }}% success
                probability)
              </p>
            </div>
          </div>

          <!-- Status Section -->
          <div class="detail-section full-width">
            <h4>Status Indicators</h4>
            <div class="status-row">
              <div class="status-item">
                <span class="status-label">Category</span>
                <span
                  class="category-badge"
                  :class="'cat-' + product.category.toLowerCase()"
                >
                  {{ product.category }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">Trend</span>
                <span
                  class="trend-badge"
                  :class="'trend-' + product.trend_direction.toLowerCase()"
                >
                  {{ product.trend_direction }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">Fatigue</span>
                <span
                  class="fatigue-badge"
                  :class="'fatigue-' + product.fatigue_status.toLowerCase()"
                >
                  {{ product.fatigue_status }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">Churn Risk</span>
                <span>{{ product.churn_risk_score.toFixed(0) }}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { MLProductAnalysis } from "@/composables/useMLAnalytics";
import { ActionBadge } from "@/components/analytics/ml";

defineProps<{
  product: MLProductAnalysis | null;
  formatCurrency: (value: number) => string;
}>();

defineEmits<{
  close: [];
}>();

function scoreClass(score: number): string {
  if (score >= 70) return "score-excellent";
  if (score >= 55) return "score-good";
  if (score >= 40) return "score-fair";
  return "score-poor";
}
</script>

<style scoped>
@import "./ProductDetailModal.styles.css";
</style>
