import { ref, computed } from "vue";
import axios from "axios";

export interface ProductFromAds {
  product_id: string;
  product_name: string;
  total_cost: number;
  total_revenue: number;
  avg_roas: number;
  source: "tiktok" | "shopee";
}

export interface SimulationRequest {
  product_id: string;
  target_roas: number;
  budget_per_day: number;
  period_days: number;
}

export interface SimulationResult {
  feasibility: "ACHIEVABLE" | "DIFFICULT" | "NOT_ACHIEVABLE";
  confidence_percent: number;
  current_roas: number;
  projected_roas: number;
  trend_prediction: "UP" | "DOWN" | "STAGNANT";
  optimal_budget: number;
  recommendation: string;
  alternatives: {
    target_roas?: number;
    required_budget?: number;
    budget?: number;
    expected_roas?: number;
  }[];
}

export interface CalendarEvent {
  date: string;
  type: string;
  description: string;
  multiplier: number;
}

export function useBudgetSimulation() {
  const loading = ref(false);
  const loadingProducts = ref(false);
  const error = ref<string | null>(null);

  const products = ref<ProductFromAds[]>([]);
  const selectedProduct = ref<ProductFromAds | null>(null);
  const targetRoas = ref<number>(5.0);
  const budgetPerDay = ref<number>(100000);
  const periodDays = ref<number>(7);

  const simulationResult = ref<SimulationResult | null>(null);
  const calendarEvents = ref<CalendarEvent[]>([]);

  const hasResult = computed(() => simulationResult.value !== null);

  const feasibilityColor = computed(() => {
    if (!simulationResult.value) return "gray";
    switch (simulationResult.value.feasibility) {
      case "ACHIEVABLE":
        return "green";
      case "DIFFICULT":
        return "orange";
      case "NOT_ACHIEVABLE":
        return "red";
      default:
        return "gray";
    }
  });

  const trendIcon = computed(() => {
    if (!simulationResult.value) return "";
    switch (simulationResult.value.trend_prediction) {
      case "UP":
        return "📈";
      case "DOWN":
        return "📉";
      case "STAGNANT":
        return "➡️";
      default:
        return "";
    }
  });

  async function fetchProducts(): Promise<void> {
    loadingProducts.value = true;
    error.value = null;
    try {
      const response = await axios.get("/api/analytics/products/from-ads");
      products.value = response.data.products || [];
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to fetch products";
      products.value = [];
    } finally {
      loadingProducts.value = false;
    }
  }

  async function fetchCalendarEvents(days: number = 30): Promise<void> {
    try {
      const response = await axios.get(
        `/api/analytics/intelligence/calendar?days=${days}`,
      );
      calendarEvents.value = response.data.events || [];
    } catch (err: any) {
      console.error("Failed to fetch calendar events:", err);
      calendarEvents.value = [];
    }
  }

  async function runSimulation(): Promise<SimulationResult | null> {
    if (!selectedProduct.value) {
      error.value = "Please select a product";
      return null;
    }

    if (targetRoas.value <= 0) {
      error.value = "Target ROAS must be greater than 0";
      return null;
    }

    if (budgetPerDay.value <= 0) {
      error.value = "Budget per day must be greater than 0";
      return null;
    }

    loading.value = true;
    error.value = null;
    simulationResult.value = null;

    try {
      const request: SimulationRequest = {
        product_id: selectedProduct.value.product_id,
        target_roas: targetRoas.value,
        budget_per_day: budgetPerDay.value,
        period_days: periodDays.value,
      };

      const response = await axios.post(
        "/api/analytics/simulation/calculate",
        request,
      );
      simulationResult.value = response.data;
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || "Simulation failed";
      return null;
    } finally {
      loading.value = false;
    }
  }

  function selectProduct(product: ProductFromAds): void {
    selectedProduct.value = product;
    simulationResult.value = null;
  }

  function resetSimulation(): void {
    simulationResult.value = null;
    error.value = null;
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  }

  function formatRoas(value: number): string {
    return `${value.toFixed(2)}x`;
  }

  function formatPercent(value: number): string {
    return `${value.toFixed(0)}%`;
  }

  return {
    // State
    loading,
    loadingProducts,
    error,
    products,
    selectedProduct,
    targetRoas,
    budgetPerDay,
    periodDays,
    simulationResult,
    calendarEvents,

    // Computed
    hasResult,
    feasibilityColor,
    trendIcon,

    // Actions
    fetchProducts,
    fetchCalendarEvents,
    runSimulation,
    selectProduct,
    resetSimulation,

    // Formatters
    formatCurrency,
    formatRoas,
    formatPercent,
  };
}
