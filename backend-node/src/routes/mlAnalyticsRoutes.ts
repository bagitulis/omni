/**
 * ML Analytics Routes
 * API endpoints for ML-powered analytics
 * Single Responsibility: Route definitions only
 */

import { Router } from "express";
import { getMLAnalyticsController } from "../controllers/mlAnalyticsController";

const router = Router();
const controller = getMLAnalyticsController();

/**
 * GET /api/analytics/ml/portfolio-health
 * Get portfolio health summary with aggregated metrics
 * Query params: platform (optional, default: tiktok)
 */
router.get("/portfolio-health", (req, res) =>
  controller.getPortfolioHealth(req, res),
);

/**
 * GET /api/analytics/ml/products
 * Get products with ML analysis scores
 * Query params: limit, cursor, sort_by, sort_dir, category, action, platform
 */
router.get("/products", (req, res) => controller.getProducts(req, res));

/**
 * GET /api/analytics/ml/product/:productId
 * Get detailed analysis for a single product
 */
router.get("/product/:productId", (req, res) =>
  controller.getProductDetail(req, res),
);

/**
 * GET /api/analytics/ml/alerts
 * Get active alerts for anomalies and issues
 */
router.get("/alerts", (req, res) => controller.getAlerts(req, res));

/**
 * GET /api/analytics/ml/distribution
 * Get score distribution across categories
 */
router.get("/distribution", (req, res) => controller.getDistribution(req, res));

/**
 * POST /api/analytics/ml/budget-sim
 * Simulate budget changes and predict outcomes
 * Body: { product_ids: string[], budget_change_pct: number }
 */
router.post("/budget-sim", (req, res) => controller.simulateBudget(req, res));

export const mlAnalyticsRouter = router;
export default mlAnalyticsRouter;
