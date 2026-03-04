/**
 * CSRF Protection Middleware
 * Prevents Cross-Site Request Forgery attacks
 *
 * Strategy: Double Submit Cookie Pattern + SameSite Cookie
 */

import { Request, Response, NextFunction } from "express";
import crypto from "crypto";
import { getLogger } from "../utils/logger";

const logger = getLogger("CSRFMiddleware");

// CSRF token config
const CSRF_TOKEN_LENGTH = 32;
const CSRF_COOKIE_NAME = "csrf_token";
const CSRF_HEADER_NAME = "x-csrf-token";

// Routes that don't need CSRF protection (webhooks, OAuth callbacks, public auth)
const CSRF_EXEMPT_PATHS = [
  "/api/webhooks",
  "/api/platform-auth",
  "/api/health",
  "/api/n8n",
  "/api/auth/login", // Login doesn't have CSRF token yet
  "/api/auth/register", // Register is also pre-auth
  "/api/csrf-token", // CSRF token endpoint itself
  "/api/captcha", // CAPTCHA endpoint (public, before login)
  "/api/status", // Public status check
  "/api/lazada", // All Lazada APIs (protected by auth token)
  "/api/tiktok", // All TikTok APIs (protected by auth token)
  "/api/shopee", // All Shopee APIs (protected by auth token)
  "/api/clone", // Clone API (protected by auth token)
];

// Methods that don't modify state
const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

/**
 * Generate a cryptographically secure CSRF token
 */
function generateCSRFToken(): string {
  return crypto.randomBytes(CSRF_TOKEN_LENGTH).toString("hex");
}

/**
 * Check if path is exempt from CSRF protection
 */
function isExemptPath(path: string): boolean {
  return CSRF_EXEMPT_PATHS.some((exempt) => path.startsWith(exempt));
}

/**
 * CSRF Protection Middleware
 *
 * For GET requests: Sets CSRF token in cookie if not present
 * For mutating requests: Validates CSRF token from header matches cookie
 */
export function csrfProtection(
  req: Request,
  res: Response,
  next: NextFunction
): void {
  // Skip for exempt paths (webhooks, OAuth callbacks)
  if (isExemptPath(req.path)) {
    return next();
  }

  // Skip for safe methods but ensure token is set
  if (SAFE_METHODS.has(req.method)) {
    ensureCSRFToken(req, res);
    return next();
  }

  // For mutating requests, validate CSRF token
  const cookieToken = req.cookies?.[CSRF_COOKIE_NAME];
  const headerToken = req.headers[CSRF_HEADER_NAME] as string;

  // Both tokens must be present and match
  if (!cookieToken || !headerToken) {
    logger.warn(`CSRF validation failed: Missing token. Path: ${req.path}`);
    res.status(403).json({
      success: false,
      error: "CSRF token missing. Please refresh the page and try again.",
      code: "CSRF_MISSING",
    });
    return;
  }

  // Constant-time comparison to prevent timing attacks
  // Ensure both tokens have the same length before comparison
  const cookieStr = String(cookieToken);
  const headerStr = String(headerToken);

  if (cookieStr.length !== headerStr.length) {
    logger.warn(`CSRF validation failed: Token mismatch. Path: ${req.path}`);
    res.status(403).json({
      success: false,
      error: "CSRF token invalid. Please refresh the page and try again.",
      code: "CSRF_INVALID",
    });
    return;
  }

  const cookieBuffer = new Uint8Array(Buffer.from(cookieStr, "utf8"));
  const headerBuffer = new Uint8Array(Buffer.from(headerStr, "utf8"));

  if (!crypto.timingSafeEqual(cookieBuffer, headerBuffer)) {
    logger.warn(`CSRF validation failed: Token mismatch. Path: ${req.path}`);
    res.status(403).json({
      success: false,
      error: "CSRF token invalid. Please refresh the page and try again.",
      code: "CSRF_INVALID",
    });
    return;
  }

  // Rotate token after successful validation for extra security
  const newToken = generateCSRFToken();
  setCSRFCookie(res, newToken);

  next();
}

/**
 * Ensure CSRF token cookie exists
 */
function ensureCSRFToken(req: Request, res: Response): void {
  if (!req.cookies?.[CSRF_COOKIE_NAME]) {
    const token = generateCSRFToken();
    setCSRFCookie(res, token);
  }
}

/**
 * Set CSRF token cookie with secure options
 */
function setCSRFCookie(res: Response, token: string): void {
  const isProduction = process.env.NODE_ENV === "production";
  res.cookie(CSRF_COOKIE_NAME, token, {
    httpOnly: false, // Must be readable by JavaScript to send in header
    secure: isProduction,
    sameSite: isProduction ? "lax" : "strict", // Lax allows top-level navigations
    maxAge: 24 * 60 * 60 * 1000, // 24 hours
    path: "/",
  });
}

/**
 * Endpoint to get a fresh CSRF token
 * Frontend can call this on app initialization
 */
export function csrfTokenEndpoint(_req: Request, res: Response): void {
  const token = generateCSRFToken();
  setCSRFCookie(res, token);

  res.json({
    success: true,
    message: "CSRF token set in cookie",
  });
}

export default csrfProtection;
