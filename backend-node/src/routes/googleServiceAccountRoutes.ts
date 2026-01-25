import { Router } from "express";
import { GoogleServiceAccountController } from "../controllers/googleServiceAccountController";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";

const router = Router();
const controller = new GoogleServiceAccountController();

/**
 * Service Account Management Routes
 * All routes require authentication
 */

// GET /api/google/service-accounts - Get service accounts list
router.get("/", authMiddleware, requireAuth, async (req, res) => {
  await controller.getServiceAccounts(req, res);
});

// GET /api/google/service-accounts/stats - Get detailed statistics
router.get("/stats", authMiddleware, requireAuth, async (req, res) => {
  await controller.getServiceAccountStats(req, res);
});

// POST /api/google/service-accounts/switch - Switch service account
router.post("/switch", authMiddleware, requireAuth, async (req, res) => {
  await controller.switchServiceAccount(req, res);
});

// POST /api/google/service-accounts/detect-sheet - Detect sheet from URL
router.post("/detect-sheet", authMiddleware, requireAuth, async (req, res) => {
  await controller.detectSheet(req, res);
});

export default router;
