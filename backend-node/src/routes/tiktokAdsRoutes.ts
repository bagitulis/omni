/**
 * TikTok Ads Routes
 * API endpoints for TikTok Ads analytics
 * Single Responsibility: Route definitions only
 */

import { Router } from "express";
import multer from "multer";
import { getTiktokAdsController } from "../controllers/tiktokAdsController";

const router = Router();

// Configure multer for file uploads (memory storage)
const upload = multer({
  storage: multer.memoryStorage(),
  limits: {
    fileSize: 50 * 1024 * 1024, // 50MB max
  },
  fileFilter: (_req, file, cb) => {
    // Accept only Excel files
    const allowedMimes = [
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      "application/vnd.ms-excel",
    ];
    const allowedExts = [".xlsx", ".xls"];

    const ext = file.originalname.toLowerCase().slice(-5);
    const isAllowedExt = allowedExts.some((e) => ext.endsWith(e));
    const isAllowedMime = allowedMimes.includes(file.mimetype);

    if (isAllowedExt || isAllowedMime) {
      cb(null, true);
    } else {
      cb(new Error("Only Excel files (.xlsx, .xls) are allowed"));
    }
  },
});

const controller = getTiktokAdsController();

/**
 * POST /api/analytics/tiktok-ads/upload
 * Upload Excel file with TikTok Ads data
 * Body: multipart/form-data with 'file' field
 * Optional body params: mode (skip|update), uploadedBy
 */
router.post("/upload", upload.single("file"), (req, res) =>
  controller.uploadFile(req, res)
);

/**
 * GET /api/analytics/tiktok-ads/dashboard
 * Get dashboard summary with aggregated metrics
 * Query params: periodStart, periodEnd (optional, ISO date strings)
 */
router.get("/dashboard", (req, res) => controller.getDashboard(req, res));

/**
 * GET /api/analytics/tiktok-ads/data
 * Get creative data with filters
 * Query params: periodStart, periodEnd, campaignId, productId,
 *   creativeType, minRoi, maxRoi, limit, offset, orderBy, orderDir
 */
router.get("/data", (req, res) => controller.getCreativeData(req, res));

/**
 * GET /api/analytics/tiktok-ads/uploads
 * Get upload history
 * Query params: limit (default 20)
 */
router.get("/uploads", (req, res) => controller.getUploadHistory(req, res));

export const tiktokAdsRouter = router;
export default tiktokAdsRouter;
