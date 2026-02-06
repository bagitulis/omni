/**
 * ML Analytics Composable
 * Handle API calls for ML-powered analytics data
 * Supports cursor-based pagination and caching
 */

import { ref, computed } from "vue";
import { useApi } from "./useApi";

const api = useApi();

// Types matching backend DTOs
export interface PortfolioHealth {
  health_score: number;
  health_label: string;
  last_updated: string;
  total_products: number;
  star_count: number;
  growth_count: number;
  stable_count: number;
  watch_count: number;
  problem_count: number;
  scale_up_count: number;
  maintain_count: number;
  reduce_count: number;
  stop_count: number;
  total_cost: number;
  total_revenue: number;
  total_profit: number;
  overall_roas: number;
  active_alerts: number;
  fatigue_warnings: number;
  churn_risks: number;
}

export interface MLProductAnalysis {
  product_id: string;
  product_name: string;
  creative_type: string;
  total_cost: number;
  total_revenue: number;
  total_profit: number;
  total_orders: number;
  roas: number;
  period_count: number;
  unified_score: number;
  roas_score: number;
  trend_score: number;
  volatility_score: number;
  momentum_score: number;
  category: string;
  action: string;
  action_label: string;
  budget_change_pct: number;
  has_fatigue_warning: boolean;
  has_churn_risk: boolean;
  fatigue_status: string;
  churn_risk_score: number;
  confidence_level: string;
  success_probability: number;
  trend_direction: string;
  trend_strength: number;
}

export interface MLAlert {
  id: string;
  product_id: string;
  product_name: string;
  alert_type: string;
  severity: string;
  message: string;
  created_at: string;
  status: string;
}

export interface ScoreDistribution {
  category: string;
  count: number;
  percentage: number;
}

export interface BudgetSimResult {
  expected_revenue: number;
  expected_profit: number;
  expected_roas: number;
  confidence_level: string;
  revenue_change_amt: number;
  profit_change_amt: number;
}

export interface PaginationMeta {
  total: number;
  limit: number;
  has_more: boolean;
  next_cursor?: string;
}

export interface AlertsMeta {
  total_active: number;
  high_priority: number;
  medium_priority: number;
  low_priority: number;
}

// State
const loading = ref(false);
const error = ref<string | null>(null);

const portfolioHealth = ref<PortfolioHealth | null>(null);
const products = ref<MLProductAnalysis[]>([]);
const productsMeta = ref<PaginationMeta>({
  total: 0,
  limit: 20,
  has_more: false,
});
const alerts = ref<MLAlert[]>([]);
const alertsMeta = ref<AlertsMeta>({
  total_active: 0,
  high_priority: 0,
  medium_priority: 0,
  low_priority: 0,
});
const distribution = ref<ScoreDistribution[]>([]);
const budgetSimResult = ref<BudgetSimResult | null>(null);

// Cache for product details
const productCache = new Map<string, MLProductAnalysis>();

export function useMLAnalytics() {
  // Fetch portfolio health summary
  async function fetchPortfolioHealth(platform: string = "tiktok") {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.get<{
        success: boolean;
        data: PortfolioHealth;
      }>(`/analytics/ml/portfolio-health?platform=${platform}`);
      if (response.success) {
        portfolioHealth.value = response.data;
      }
      return response;
    } catch (err) {
      error.value = "Failed to fetch portfolio health";
      console.error("Portfolio health error:", err);
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Fetch products with pagination
  async function fetchProducts(
    options: {
      limit?: number;
      cursor?: string;
      sortBy?: string;
      sortDir?: string;
      category?: string;
      action?: string;
      platform?: string;
      append?: boolean;
    } = {},
  ) {
    loading.value = true;
    error.value = null;
    try {
      const params = new URLSearchParams();
      if (options.limit) params.append("limit", options.limit.toString());
      if (options.cursor) params.append("cursor", options.cursor);
      if (options.sortBy) params.append("sort_by", options.sortBy);
      if (options.sortDir) params.append("sort_dir", options.sortDir);
      if (options.category) params.append("category", options.category);
      if (options.action) params.append("action", options.action);
      if (options.platform) params.append("platform", options.platform);

      const response = await api.get<{
        success: boolean;
        data: MLProductAnalysis[];
        meta: PaginationMeta;
      }>(`/analytics/ml/products?${params.toString()}`);

      if (response.success) {
        if (options.append && options.cursor) {
          products.value = [...products.value, ...response.data];
        } else {
          products.value = response.data;
        }
        productsMeta.value = response.meta;
      }
      return response;
    } catch (err) {
      error.value = "Failed to fetch products";
      console.error("Products error:", err);
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Fetch single product detail
  async function fetchProductDetail(productId: string) {
    // Check cache first
    if (productCache.has(productId)) {
      return { success: true, data: productCache.get(productId)! };
    }

    loading.value = true;
    try {
      const response = await api.get<{
        success: boolean;
        data: MLProductAnalysis;
      }>(`/analytics/ml/product/${productId}`);
      if (response.success) {
        productCache.set(productId, response.data);
        // Limit cache size
        if (productCache.size > 20) {
          const firstKey = productCache.keys().next().value;
          if (firstKey) productCache.delete(firstKey);
        }
      }
      return response;
    } catch (err) {
      console.error("Product detail error:", err);
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Fetch alerts
  async function fetchAlerts() {
    loading.value = true;
    try {
      const response = await api.get<{
        success: boolean;
        data: MLAlert[];
        meta: AlertsMeta;
      }>("/analytics/ml/alerts");
      if (response.success) {
        alerts.value = response.data;
        alertsMeta.value = response.meta;
      }
      return response;
    } catch (err) {
      console.error("Alerts error:", err);
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Fetch score distribution
  async function fetchDistribution() {
    try {
      const response = await api.get<{
        success: boolean;
        data: ScoreDistribution[];
      }>("/analytics/ml/distribution");
      if (response.success) {
        distribution.value = response.data;
      }
      return response;
    } catch (err) {
      console.error("Distribution error:", err);
      return null;
    }
  }

  // Simulate budget change
  async function simulateBudget(productIds: string[], budgetChangePct: number) {
    loading.value = true;
    try {
      const response = await api.post<{
        success: boolean;
        data: BudgetSimResult;
      }>("/analytics/ml/budget-sim", {
        product_ids: productIds,
        budget_change_pct: budgetChangePct,
      });
      if (response.success) {
        budgetSimResult.value = response.data;
      }
      return response;
    } catch (err) {
      console.error("Budget simulation error:", err);
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Load more products (infinite scroll)
  async function loadMoreProducts() {
    if (!productsMeta.value.has_more || !productsMeta.value.next_cursor) {
      return;
    }
    await fetchProducts({
      cursor: productsMeta.value.next_cursor,
      append: true,
    });
  }

  // Formatters
  function formatCurrency(value: number): string {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0,
    }).format(value);
  }

  function formatNumber(value: number): string {
    return new Intl.NumberFormat("id-ID").format(value);
  }

  function formatPercent(value: number): string {
    return `${(value * 100).toFixed(1)}%`;
  }

  function formatScore(value: number): string {
    return value.toFixed(1);
  }

  // Computed
  const hasAlerts = computed(() => alerts.value.length > 0);
  const highPriorityAlerts = computed(() =>
    alerts.value.filter((a) => a.severity === "HIGH"),
  );
  const topProducts = computed(() =>
    [...products.value]
      .sort((a, b) => b.unified_score - a.unified_score)
      .slice(0, 10),
  );
  const problemProducts = computed(() =>
    products.value.filter((p) => p.category === "PROBLEM"),
  );

  // Clear cache
  function clearCache() {
    productCache.clear();
  }

  return {
    // State
    loading,
    error,
    portfolioHealth,
    products,
    productsMeta,
    alerts,
    alertsMeta,
    distribution,
    budgetSimResult,

    // Actions
    fetchPortfolioHealth,
    fetchProducts,
    fetchProductDetail,
    fetchAlerts,
    fetchDistribution,
    simulateBudget,
    loadMoreProducts,
    clearCache,

    // Formatters
    formatCurrency,
    formatNumber,
    formatPercent,
    formatScore,

    // Computed
    hasAlerts,
    highPriorityAlerts,
    topProducts,
    problemProducts,
  };
}
