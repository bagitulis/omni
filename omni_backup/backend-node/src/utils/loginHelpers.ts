import { AuthService } from "../services/authService";
import { getDbManager } from "../services/dbConnectionManager";
import captchaService from "../services/captchaService";

export interface LoginStatus {
  requiresCaptcha: boolean;
  isLocked: boolean;
  lockMinutesRemaining: number;
}

export interface LoginResult {
  user: any;
  [key: string]: any;
}

/**
 * Check login status across all tenants for CAPTCHA/lock requirements
 */
export async function checkLoginStatusAcrossTenants(
  username: string,
): Promise<LoginStatus> {
  const tenantsToCheck = getDbManager()
    .getAvailableTenants()
    .map((t) => t.id);

  const loginStatus: LoginStatus = {
    requiresCaptcha: false,
    isLocked: false,
    lockMinutesRemaining: 0,
  };

  for (const tenantId of tenantsToCheck) {
    try {
      const prisma = getDbManager().getConnection(tenantId);
      const authService = new AuthService(prisma);
      const status = await authService.getLoginStatus(username);

      if (status.isLocked || status.requiresCaptcha) {
        return {
          requiresCaptcha: status.requiresCaptcha,
          isLocked: status.isLocked,
          lockMinutesRemaining: status.lockMinutesRemaining || 0,
        };
      }
    } catch {
      continue;
    }
  }

  return loginStatus;
}

/**
 * Verify reCAPTCHA token
 */
export async function verifyCaptcha(
  recaptchaToken: string,
  ipAddress: string,
): Promise<{ valid: boolean; error?: string }> {
  return captchaService.verifyCaptcha(recaptchaToken, ipAddress);
}

/**
 * Find user across all tenants
 */
export async function findUserAcrossTenants(
  username: string,
  password: string,
): Promise<{ result: LoginResult | null; tenantId: string | null }> {
  const tenantsToCheck = getDbManager()
    .getAvailableTenants()
    .map((t) => t.id);

  console.log(
    `[AuthController] 🔍 Scanning tenants: ${tenantsToCheck.join(", ")}`,
  );

  for (const tenantId of tenantsToCheck) {
    try {
      console.log(`[AuthController] 🔍 Checking tenant: ${tenantId}`);
      const prisma = getDbManager().getConnection(tenantId);
      const authService = new AuthService(prisma);
      const result = await authService.login({ username, password });

      if (result && result.user) {
        console.log(`[AuthController] ✅ User found in tenant: ${tenantId}`);
        return { result, tenantId };
      }
    } catch (e: any) {
      console.log(`[AuthController] ❌ Not found in ${tenantId}: ${e.message}`);

      // If account is locked, throw to handle in controller
      if (e.message.includes("locked")) {
        throw e;
      }
      continue;
    }
  }

  return { result: null, tenantId: null };
}

/**
 * Get post-failed-login status to determine if CAPTCHA should be shown
 */
export async function getPostFailedLoginStatus(
  username: string,
): Promise<LoginStatus | null> {
  const tenantsToCheck = getDbManager()
    .getAvailableTenants()
    .map((t) => t.id);

  for (const tenantId of tenantsToCheck) {
    try {
      const prisma = getDbManager().getConnection(tenantId);
      const authService = new AuthService(prisma);
      const status = await authService.getLoginStatus(username);

      if (status.requiresCaptcha || status.isLocked) {
        return {
          requiresCaptcha: status.requiresCaptcha,
          isLocked: status.isLocked,
          lockMinutesRemaining: status.lockMinutesRemaining || 0,
        };
      }
    } catch {
      continue;
    }
  }

  return null;
}
