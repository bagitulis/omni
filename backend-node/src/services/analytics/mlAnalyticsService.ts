/**
 * ML Analytics Service
 * Business logic for ML-powered product analytics
 * Single Responsibility: ML analytics calculations and data aggregation
 */

import { PrismaClient } from "@prisma/client";
import { getLogger } from "../../utils/logger";

const logger = getLogger("MLAnalyticsService");

// Response types matching frontend expectations (snake_case)
export interface PortfolioHealthResponse {
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

export class MLAnalyticsService {
  constructor(
    private prisma: PrismaClient,
    private tenantId: string,
  ) {}

  /**
   * Get portfolio health summary
   */
  async getPortfolioHealth(): Promise<PortfolioHealthResponse> {
    try {
      // Get all product summaries for analysis
      const summaries = await this.prisma.tiktokAdsProductSummary.findMany({
        where: {
          tenantId: this.tenantId,
          periodType: "monthly",
        },
        orderBy: {
          periodDate: "desc",
        },
      });

      if (summaries.length === 0) {
        return this.getEmptyPortfolioHealth();
      }

      // Aggregate metrics
      const totalCost = summaries.reduce((sum, s) => sum + s.totalCost, 0);
      const totalRevenue = summaries.reduce(
        (sum, s) => sum + s.totalRevenue,
        0,
      );
      const totalProfit = totalRevenue - totalCost;
      const overallRoas = totalCost > 0 ? totalRevenue / totalCost : 0;

      // Get predictions for categorization
      const predictions = await this.prisma.tiktokAdsMLPrediction.findMany({
        where: {
          tenantId: this.tenantId,
        },
      });

      // Categorize products
      const categories = this.categorizeProducts(summaries);
      const actions = this.getActionCounts(summaries);

      // Calculate health score (0-100)
      const healthScore = this.calculateHealthScore({
        totalProducts: summaries.length,
        avgRoas: overallRoas,
        profitMargin: totalRevenue > 0 ? totalProfit / totalRevenue : 0,
        categories,
      });

      const healthLabel = this.getHealthLabel(healthScore);

      return {
        health_score: Math.round(healthScore),
        health_label: healthLabel,
        last_updated: new Date().toISOString(),
        total_products: summaries.length,
        star_count: categories.star,
        growth_count: categories.growth,
        stable_count: categories.stable,
        watch_count: categories.watch,
        problem_count: categories.problem,
        scale_up_count: actions.scaleUp,
        maintain_count: actions.maintain,
        reduce_count: actions.reduce,
        stop_count: actions.stop,
        total_cost: totalCost,
        total_revenue: totalRevenue,
        total_profit: totalProfit,
        overall_roas: Math.round(overallRoas * 100) / 100,
        active_alerts: predictions.filter((p) => p.hasAnomaly).length,
        fatigue_warnings: 0,
        churn_risks: predictions.filter((p) => p.hasAnomaly).length,
      };
    } catch (error) {
      logger.error(`Portfolio health calculation failed: ${error}`);
      throw error;
    }
  }

  /**
   * Get paginated list of products with ML analysis
   */
  async getProducts(options: {
    limit?: number;
    cursor?: string;
    sortBy?: string;
    sortDir?: string;
    category?: string;
    action?: string;
    platform?: string;
  }): Promise<{
    products: MLProductAnalysis[];
    meta: PaginationMeta;
  }> {
    const limit = options.limit || 20;

    try {
      // Get product summaries
      const summaries = await this.prisma.tiktokAdsProductSummary.findMany({
        where: {
          tenantId: this.tenantId,
          periodType: "monthly",
        },
        take: limit + 1,
        orderBy: {
          totalRevenue: "desc",
        },
      });

      const hasMore = summaries.length > limit;
      const products = summaries.slice(0, limit);

      // Get predictions for each product
      const productIds = products.map((p) => p.productId);
      const predictions = await this.prisma.tiktokAdsMLPrediction.findMany({
        where: {
          tenantId: this.tenantId,
          productId: { in: productIds },
        },
      });

      const predictionMap = new Map(predictions.map((p) => [p.productId, p]));

      // Try to get product names from TiktokProduct table
      let productNames: Map<string, string> = new Map();
      try {
        const tiktokProducts = await this.prisma.tiktokProduct.findMany({
          where: {
            tenantId: this.tenantId,
            productId: { in: productIds },
          },
          select: {
            productId: true,
            name: true,
          },
        });
        productNames = new Map(
          tiktokProducts.map((p) => [p.productId, p.name]),
        );
      } catch {
        // Table might not exist or be empty
      }

      // Transform to MLProductAnalysis
      const analysisResults: MLProductAnalysis[] = products.map((summary) => {
        const prediction = predictionMap.get(summary.productId);
        const profit = summary.totalRevenue - summary.totalCost;
        const roas =
          summary.totalCost > 0 ? summary.totalRevenue / summary.totalCost : 0;

        // Calculate scores
        const roasScore = Math.min(roas * 20, 100);
        const trendScore = this.calculateTrendScore(
          summary.roiTrend || "stable",
        );
        const volatilityScore = 75;
        const momentumScore = this.calculateMomentumScore(
          summary.ordersTrend || "stable",
        );
        const unifiedScore =
          (roasScore + trendScore + volatilityScore + momentumScore) / 4;

        // Categorize
        const category = this.getProductCategory(roas, profit);
        const action = this.getProductAction(category, roas, profit);

        return {
          product_id: summary.productId,
          product_name:
            productNames.get(summary.productId) ||
            summary.topVideoTitle ||
            `Product ${summary.productId.slice(-6)}`,
          creative_type: summary.bestCreativeType || "Unknown",
          total_cost: Math.round(summary.totalCost),
          total_revenue: Math.round(summary.totalRevenue),
          total_profit: Math.round(profit),
          total_orders: summary.totalOrders,
          roas: Math.round(roas * 100) / 100,
          period_count: 1,
          unified_score: Math.round(unifiedScore * 10) / 10,
          roas_score: Math.round(roasScore * 10) / 10,
          trend_score: Math.round(trendScore * 10) / 10,
          volatility_score: Math.round(volatilityScore * 10) / 10,
          momentum_score: Math.round(momentumScore * 10) / 10,
          category,
          action,
          action_label: this.getActionLabel(action),
          budget_change_pct: this.getBudgetChange(action),
          has_fatigue_warning: false,
          has_churn_risk: prediction?.hasAnomaly || false,
          fatigue_status: "none",
          churn_risk_score: prediction?.hasAnomaly ? 0.7 : 0,
          confidence_level: prediction?.roiConfidence
            ? this.getConfidenceLevel(prediction.roiConfidence)
            : "medium",
          success_probability: prediction?.roiConfidence || 0.5,
          trend_direction: summary.roiTrend || "stable",
          trend_strength: 0.5,
        };
      });

      // Filter by category/action if specified
      let filteredResults = analysisResults;
      if (options.category) {
        filteredResults = filteredResults.filter(
          (p) => p.category === options.category!.toUpperCase(),
        );
      }
      if (options.action) {
        filteredResults = filteredResults.filter(
          (p) => p.action === options.action!.toUpperCase(),
        );
      }

      return {
        products: filteredResults,
        meta: {
          total: filteredResults.length,
          limit,
          has_more: hasMore,
          next_cursor: hasMore ? products[products.length - 1].id : undefined,
        },
      };
    } catch (error) {
      logger.error(`Get products failed: ${error}`);
      throw error;
    }
  }

  /**
   * Get single product detail
   */
  async getProductDetail(productId: string): Promise<MLProductAnalysis | null> {
    const result = await this.getProducts({ limit: 100 });
    return result.products.find((p) => p.product_id === productId) || null;
  }

  /**
   * Get active alerts
   */
  async getAlerts(): Promise<{
    alerts: MLAlert[];
    meta: {
      total_active: number;
      high_priority: number;
      medium_priority: number;
      low_priority: number;
    };
  }> {
    try {
      const predictions = await this.prisma.tiktokAdsMLPrediction.findMany({
        where: {
          tenantId: this.tenantId,
          hasAnomaly: true,
        },
        take: 20,
        orderBy: {
          createdAt: "desc",
        },
      });

      const alerts: MLAlert[] = predictions.map((p) => ({
        id: p.id,
        product_id: p.productId,
        product_name: `Product ${p.productId.slice(-6)}`,
        alert_type: p.anomalyType || "anomaly",
        severity: "high",
        message: p.anomalyDescription || "Performance anomaly detected",
        created_at: p.createdAt.toISOString(),
        status: "active",
      }));

      return {
        alerts,
        meta: {
          total_active: alerts.length,
          high_priority: alerts.filter((a) => a.severity === "high").length,
          medium_priority: alerts.filter((a) => a.severity === "medium").length,
          low_priority: alerts.filter((a) => a.severity === "low").length,
        },
      };
    } catch (error) {
      logger.error(`Get alerts failed: ${error}`);
      return {
        alerts: [],
        meta: {
          total_active: 0,
          high_priority: 0,
          medium_priority: 0,
          low_priority: 0,
        },
      };
    }
  }

  /**
   * Get score distribution
   */
  async getDistribution(): Promise<ScoreDistribution[]> {
    try {
      const products = await this.getProducts({ limit: 1000 });

      const categories = products.products.reduce(
        (acc, p) => {
          acc[p.category] = (acc[p.category] || 0) + 1;
          return acc;
        },
        {} as Record<string, number>,
      );

      const total = products.products.length;

      return Object.entries(categories).map(([category, count]) => ({
        category,
        count,
        percentage: total > 0 ? Math.round((count / total) * 100) / 100 : 0,
      }));
    } catch (error) {
      logger.error(`Get distribution failed: ${error}`);
      return [];
    }
  }

  /**
   * Simulate budget change
   */
  async simulateBudget(
    productIds: string[],
    budgetChangePct: number,
  ): Promise<BudgetSimResult> {
    try {
      const products = await this.getProducts({ limit: 1000 });
      const selectedProducts = products.products.filter((p) =>
        productIds.includes(p.product_id),
      );

      if (selectedProducts.length === 0) {
        return {
          expected_revenue: 0,
          expected_profit: 0,
          expected_roas: 0,
          confidence_level: "low",
          revenue_change_amt: 0,
          profit_change_amt: 0,
        };
      }

      const currentRevenue = selectedProducts.reduce(
        (sum, p) => sum + p.total_revenue,
        0,
      );
      const currentCost = selectedProducts.reduce(
        (sum, p) => sum + p.total_cost,
        0,
      );
      const currentProfit = currentRevenue - currentCost;
      const avgRoas = currentCost > 0 ? currentRevenue / currentCost : 0;

      // Simulate with diminishing returns
      const newCost = currentCost * (1 + budgetChangePct / 100);
      const expectedRoas = avgRoas * (budgetChangePct > 0 ? 0.95 : 1.05);
      const expectedRevenue = newCost * expectedRoas;
      const expectedProfit = expectedRevenue - newCost;

      return {
        expected_revenue: Math.round(expectedRevenue),
        expected_profit: Math.round(expectedProfit),
        expected_roas: Math.round(expectedRoas * 100) / 100,
        confidence_level: "medium",
        revenue_change_amt: Math.round(expectedRevenue - currentRevenue),
        profit_change_amt: Math.round(expectedProfit - currentProfit),
      };
    } catch (error) {
      logger.error(`Budget simulation failed: ${error}`);
      throw error;
    }
  }

  // ==================== PRIVATE HELPERS ====================

  private getEmptyPortfolioHealth(): PortfolioHealthResponse {
    return {
      health_score: 0,
      health_label: "No Data",
      last_updated: new Date().toISOString(),
      total_products: 0,
      star_count: 0,
      growth_count: 0,
      stable_count: 0,
      watch_count: 0,
      problem_count: 0,
      scale_up_count: 0,
      maintain_count: 0,
      reduce_count: 0,
      stop_count: 0,
      total_cost: 0,
      total_revenue: 0,
      total_profit: 0,
      overall_roas: 0,
      active_alerts: 0,
      fatigue_warnings: 0,
      churn_risks: 0,
    };
  }

  private categorizeProducts(summaries: any[]): Record<string, number> {
    const categories = {
      star: 0,
      growth: 0,
      stable: 0,
      watch: 0,
      problem: 0,
    };

    summaries.forEach((s) => {
      const roas = s.totalCost > 0 ? s.totalRevenue / s.totalCost : 0;
      const profit = s.totalRevenue - s.totalCost;

      if (roas > 3 && profit > 1000000) categories.star++;
      else if (roas > 2 && s.roiTrend === "up") categories.growth++;
      else if (roas > 1.5 && roas <= 2.5) categories.stable++;
      else if (roas > 1 && roas <= 1.5) categories.watch++;
      else categories.problem++;
    });

    return categories;
  }

  private getActionCounts(summaries: any[]): Record<string, number> {
    const actions = {
      scaleUp: 0,
      maintain: 0,
      reduce: 0,
      stop: 0,
    };

    summaries.forEach((s) => {
      const roas = s.totalCost > 0 ? s.totalRevenue / s.totalCost : 0;

      if (roas > 2.5) actions.scaleUp++;
      else if (roas > 1.5) actions.maintain++;
      else if (roas > 1) actions.reduce++;
      else actions.stop++;
    });

    return actions;
  }

  private calculateHealthScore(data: {
    totalProducts: number;
    avgRoas: number;
    profitMargin: number;
    categories: Record<string, number>;
  }): number {
    const { totalProducts, avgRoas, profitMargin, categories } = data;

    if (totalProducts === 0) return 0;

    const roasScore = Math.min(avgRoas * 25, 50);
    const profitScore = Math.min(profitMargin * 100, 30);
    const categoryScore =
      ((categories.star * 1 +
        categories.growth * 0.8 +
        categories.stable * 0.5) /
        totalProducts) *
      20;

    return roasScore + profitScore + categoryScore;
  }

  private getHealthLabel(score: number): string {
    if (score >= 80) return "Excellent";
    if (score >= 60) return "Good";
    if (score >= 40) return "Fair";
    if (score >= 20) return "Poor";
    return "Critical";
  }

  private calculateTrendScore(trend: string): number {
    if (trend === "up") return 85;
    if (trend === "stable") return 70;
    return 40;
  }

  private calculateMomentumScore(trend: string): number {
    if (trend === "up") return 80;
    if (trend === "stable") return 60;
    return 30;
  }

  private getProductCategory(roas: number, profit: number): string {
    if (roas > 3 && profit > 1000000) return "STAR";
    if (roas > 2 && profit > 500000) return "GROWTH";
    if (roas > 1.5 && roas <= 2.5) return "STABLE";
    if (roas > 1 && roas <= 1.5) return "WATCH";
    return "PROBLEM";
  }

  private getProductAction(
    category: string,
    roas: number,
    profit: number,
  ): string {
    if (category === "STAR" || (roas > 2.5 && profit > 0)) return "SCALE_UP";
    if (category === "GROWTH" || (roas > 1.5 && profit > 0)) return "MAINTAIN";
    if (category === "WATCH" || (roas > 1 && profit > 0)) return "REDUCE";
    return "STOP";
  }

  private getActionLabel(action: string): string {
    const labels: Record<string, string> = {
      SCALE_UP: "Scale Up Budget",
      MAINTAIN: "Maintain Current",
      REDUCE: "Reduce Budget",
      STOP: "Stop Campaign",
    };
    return labels[action] || "No Action";
  }

  private getBudgetChange(action: string): number {
    const changes: Record<string, number> = {
      SCALE_UP: 25,
      MAINTAIN: 0,
      REDUCE: -25,
      STOP: -100,
    };
    return changes[action] || 0;
  }

  private getConfidenceLevel(confidence: number): string {
    if (confidence >= 0.8) return "high";
    if (confidence >= 0.5) return "medium";
    return "low";
  }
}
