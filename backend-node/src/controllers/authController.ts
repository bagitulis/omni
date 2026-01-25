import { Request, Response } from "express";
import { AuthService } from "../services/authService";
import { JWTService } from "../services/jwtService";
import { getUserManagementService } from "../services/userManagementService";
import { getDbManager } from "../services/dbConnectionManager";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { tenantContext } from "../utils/tenantContext";
import {
  generateCSRFToken,
  setCSRFCookie,
  getClientIP,
  DEVELOPER_TENANTS,
} from "../utils/authHelpers";
import {
  checkLoginStatusAcrossTenants,
  verifyCaptcha,
  findUserAcrossTenants,
  getPostFailedLoginStatus,
} from "../utils/loginHelpers";

export class AuthController {
  private jwtService = new JWTService();

  /**
   * Register new user
   * POST /api/auth/register
   *
   * Auto-provisions:
   *   ✅ Tenant database (Prisma)
   *   ✅ Job database
   *   ✅ Tenant configuration
   */
  async register(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { username, email, password } = req.body;
      const shopName = req.body.shopName || `Shop ${username}`;

      // Get user management service
      const userService = getUserManagementService();

      // Create user with automatic tenant provisioning
      const user = await userService.createUser({
        username,
        email,
        password,
        shopName,
      });

      res.status(201).json({
        message: "User registered successfully",
        user,
      });
    } catch (error: any) {
      res.status(400).json({ error: error.message });
    }
  }

  /**
   * Login user
   * POST /api/auth/login
   *
   * Security flow:
   * - 1-3 failed attempts: Normal login (no CAPTCHA)
   * - 4-10 failed attempts: Require reCAPTCHA
   * - 10+ failed attempts: Lock account for 30 minutes
   */
  async login(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { username, password, recaptchaToken } = req.body;
      const ipAddress = getClientIP(req);

      console.log(`[AuthController] 🔐 Login attempt for user: ${username}`);

      if (!username || !password) {
        res.status(400).json({ error: "Username and password are required" });
        return;
      }

      // Check login status across all tenants
      const loginStatus = await checkLoginStatusAcrossTenants(username);

      if (loginStatus.isLocked) {
        res.status(423).json({
          error: `Account locked. Try again in ${loginStatus.lockMinutesRemaining} minute(s).`,
          code: "ACCOUNT_LOCKED",
          isLocked: true,
          lockMinutesRemaining: loginStatus.lockMinutesRemaining,
        });
        return;
      }

      // Verify CAPTCHA if required
      if (loginStatus.requiresCaptcha) {
        if (!recaptchaToken) {
          res.status(400).json({
            error: "reCAPTCHA verification required",
            code: "CAPTCHA_REQUIRED",
            requiresCaptcha: true,
          });
          return;
        }

        const captchaResult = await verifyCaptcha(recaptchaToken, ipAddress);
        if (!captchaResult.valid) {
          res.status(400).json({
            error: captchaResult.error || "reCAPTCHA verification failed",
            code: "CAPTCHA_INVALID",
            requiresCaptcha: true,
          });
          return;
        }
      }

      // Find user across all tenants
      let foundResult, foundTenantId;
      try {
        const result = await findUserAcrossTenants(username, password);
        foundResult = result.result;
        foundTenantId = result.tenantId;
      } catch (e: any) {
        if (e.message.includes("locked")) {
          res.status(423).json({
            error: e.message,
            code: "ACCOUNT_LOCKED",
            isLocked: true,
          });
          return;
        }
        throw e;
      }

      if (!foundResult || !foundTenantId) {
        const postStatus = await getPostFailedLoginStatus(username);
        if (postStatus) {
          res.status(401).json({
            error: "Invalid username or password",
            requiresCaptcha: postStatus.requiresCaptcha,
            isLocked: postStatus.isLocked,
            lockMinutesRemaining: postStatus.lockMinutesRemaining,
          });
          return;
        }
        res.status(401).json({ error: "Invalid username or password" });
        return;
      }

      // Generate JWT token
      const token = this.jwtService.generateToken({
        userId: foundResult.user.id,
        username: foundResult.user.username,
        role: foundResult.user.role,
        tenantId: foundTenantId,
      });

      // Set CSRF token cookie
      const csrfToken = generateCSRFToken();
      setCSRFCookie(res, csrfToken);

      res.status(200).json({
        message: "Login successful",
        user: foundResult.user,
        token,
        tenantId: foundTenantId,
      });
    } catch (error: any) {
      res.status(401).json({ error: error.message });
    }
  }

  /**
   * Get current user info
   * GET /api/auth/me
   */
  async getCurrentUser(req: AuthRequest, res: Response): Promise<void> {
    try {
      const userId = (req as any).userId;
      const tenantId = req.tenantId || tenantContext.getTenantId();

      if (!userId) {
        res.status(401).json({ error: "Unauthorized" });
        return;
      }

      const prisma = getDbManager().getConnection(tenantId);
      const authService = new AuthService(prisma);

      const user = await authService.getUserById(userId);

      if (!user) {
        res.status(404).json({ error: "User not found" });
        return;
      }

      res.status(200).json(user);
    } catch (error: any) {
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * Verify token
   * POST /api/auth/verify
   */
  async verifyToken(req: Request, res: Response): Promise<void> {
    try {
      const authHeader = req.headers.authorization;

      if (!authHeader) {
        res.status(401).json({ error: "No token provided" });
        return;
      }

      const token = this.jwtService.extractToken(authHeader);

      if (!token) {
        res.status(401).json({ error: "Invalid token format" });
        return;
      }

      const payload = this.jwtService.verifyToken(token);

      res.status(200).json({
        valid: true,
        payload,
      });
    } catch (error: any) {
      res.status(401).json({
        valid: false,
        error: error.message,
      });
    }
  }

  /**
   * Get available tenants for switching
   * GET /api/auth/tenants
   *
   * Returns only tenants with actual platform data (excludes developer accounts)
   * Developer accounts (tester, developer, admin) don't have platform credentials
   */
  async getAvailableTenants(_req: Request, res: Response): Promise<void> {
    try {
      const allTenants = getDbManager().getAvailableTenants();

      // Filter out developer accounts - they don't have platform data
      const tenants = allTenants.filter(
        (t) => !DEVELOPER_TENANTS.includes(t.id.toLowerCase()),
      );

      res.status(200).json({ tenants });
    } catch (error: any) {
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * Switch tenant for current user (developer only)
   * POST /api/auth/switch-tenant
   */
  async switchTenant(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { tenantId } = req.body;
      // authMiddleware sets req.role, tenantMiddleware sets req.userRole
      const userRole = (req as any).role || (req as any).userRole;

      if (!tenantId) {
        res.status(400).json({ error: "tenantId is required" });
        return;
      }

      // Only developer can switch tenants
      if (userRole !== "developer") {
        res.status(403).json({ error: "Only developers can switch tenants" });
        return;
      }

      // Validate tenant exists
      const tenants = getDbManager().getAvailableTenants();
      if (!tenants.find((t) => t.id === tenantId)) {
        res.status(400).json({ error: `Tenant '${tenantId}' not found` });
        return;
      }

      // Generate new token dengan tenant baru
      // IMPORTANT: Use req.role (set by authMiddleware), NOT req.userRole
      const token = this.jwtService.generateToken({
        userId: (req as any).userId,
        username: (req as any).username,
        role: userRole, // Use the verified userRole from check above
        tenantId,
      });

      res.status(200).json({
        message: `Switched to tenant '${tenantId}'`,
        token,
        tenantId,
      });
    } catch (error: any) {
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * Logout user
   * POST /api/auth/logout
   */
  async logout(_req: Request, res: Response): Promise<void> {
    try {
      // Logout is a client-side operation (clear token from localStorage)
      // Server can invalidate tokens if needed, but for now just confirm
      res.status(200).json({
        message: "Logout successful",
      });
    } catch (error: any) {
      res.status(500).json({ error: error.message });
    }
  }
}
