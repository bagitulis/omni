/**
 * Shopee Ads Routes
 * API endpoints for Shopee Ads analytics
 * Single Responsibility: Route definitions only
 */

import { Router } from "express";
import multer from "multer";
import { getShopeeAdsController } from "../controllers/shopeeAdsController";

const router = Router();

// Configure multer for file uploads (memory storage)
const upload = multer({
  storage: multer.memoryStorage(),
  limits: {
    fileSize: 50 * 1024 * 1024, // 50MB max
  },
  fileFilter: (_req, file, cb) => {
    // Accept CSV files
    const allowedMimes = ["text/csv", "application/vnd.ms-excel"];
    const allowedExts = [".csv"];

    const ext = file.originalname.toLowerCase().slice(-4);
    const isAllowedExt = allowedExts.some((e) => ext.endsWith(e));
    const isAllowedMime =
      allowedMimes.includes(file.mimetype) ||
      file.mimetype.includes("csv") ||
      file.mimetype === "application/octet-stream";

    if (isAllowedExt || isAllowedMime) {
      cb(null, true);
    } else {
      cb(new Error("Only CSV files (.csv) are allowed"));
    }
  },
});

const controller = getShopeeAdsController();

/**
 * POST /api/analytics/shopee-ads/upload
 * Upload CSV file with Shopee Ads data
 * Body: multipart/form-data with 'file' field
 * Optional body params: mode (skip|update), uploadedBy
 */
router.post("/upload", upload.single("file"), (req, res) =>
  controller.uploadFile(req, res),
);

/**
 * GET /api/analytics/shopee-ads/dashboard
 * Get dashboard summary with aggregated metrics
 * Query params: periodStart, periodEnd (optional, ISO date strings)
 */
router.get("/dashboard", (req, res) => controller.getDashboard(req, res));

/**
 * GET /api/analytics/shopee-ads/data
 * Get product data with filters
 * Query params: periodStart, periodEnd, productId, biddingMode,
 *   minRoas, maxRoas, limit, offset, orderBy, orderDir
 */
router.get("/data", (req, res) => controller.getProductData(req, res));

/**
 * GET /api/analytics/shopee-ads/uploads
 * Get upload history
 * Query params: limit (default 20)
 */
router.get("/uploads", (req, res) => controller.getUploadHistory(req, res));

export const shopeeAdsRouter = router;
export default shopeeAdsRouter;
