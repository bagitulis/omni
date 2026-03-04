import { Router, Request, Response } from "express";
import { AuthController } from "../controllers/authController";
import { authMiddleware } from "../middleware/authMiddleware";
import { loginLimiter, registerLimiter } from "../middleware/rateLimiter";
import {
  validateLogin,
  validateRegistration,
  handleValidationErrors,
} from "../middleware/validation";

const router = Router();
const authController = new AuthController();

/**
 * Public routes (no authentication required)
 * Tenant context is set by global middleware in server.ts
 */

// POST /api/auth/register - Register new user
router.post(
  "/register",
  registerLimiter,
  validateRegistration,
  handleValidationErrors,
  async (req: Request, res: Response) => {
    await authController.register(req, res);
  }
);

// POST /api/auth/login - Login user
router.post(
  "/login",
  loginLimiter,
  validateLogin,
  handleValidationErrors,
  async (req: Request, res: Response) => {
    await authController.login(req, res);
  }
);

// POST /api/auth/verify - Verify token validity
router.post("/verify", async (req, res) => {
  await authController.verifyToken(req, res);
});

// GET /api/auth/tenants - Get available tenants for switching
router.get("/tenants", async (req, res) => {
  await authController.getAvailableTenants(req, res);
});

/**
 * Protected routes (authentication required)
 */

// GET /api/auth/me - Get current user info
router.get("/me", authMiddleware, async (req, res) => {
  await authController.getCurrentUser(req, res);
});

// POST /api/auth/switch-tenant - Switch to different tenant
router.post("/switch-tenant", authMiddleware, async (req, res) => {
  await authController.switchTenant(req, res);
});

// POST /api/auth/logout - Logout user
router.post("/logout", authMiddleware, async (req, res) => {
  await authController.logout(req, res);
});

export default router;
