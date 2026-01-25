import { Request, Response } from "express";
import crypto from "crypto";

/**
 * Generate CSRF token
 */
export function generateCSRFToken(): string {
  return crypto.randomBytes(32).toString("hex");
}

/**
 * Set CSRF cookie
 */
export function setCSRFCookie(res: Response, token: string): void {
  const isProduction = process.env.NODE_ENV === "production";
  res.cookie("csrf_token", token, {
    httpOnly: false, // Must be readable by JavaScript
    secure: isProduction,
    sameSite: isProduction ? "lax" : "strict", // Lax allows top-level navigations
    maxAge: 24 * 60 * 60 * 1000, // 24 hours
    path: "/",
  });
}

/**
 * Get client IP address from request
 */
export function getClientIP(req: Request): string {
  const forwarded = req.headers["x-forwarded-for"];
  if (forwarded) {
    const ips = Array.isArray(forwarded)
      ? forwarded[0]
      : forwarded.split(",")[0];
    return ips.trim();
  }
  return req.ip || req.socket.remoteAddress || "unknown";
}

/**
 * Special user roles/names that should be excluded from tenant switching
 * ⚠️ NOTE: These are USER ROLES, not tenant IDs!
 * 'system' here refers to system-level operations, NOT the system.db tenant
 *
 * Real tenants are: yumna_bertigamart, tika_nusseyba
 * system.db is for GlobalConfig only (partnerId, appKey, etc.)
 */
export const DEVELOPER_ROLES = ["system", "developer", "admin"];

/**
 * @deprecated Use DEVELOPER_ROLES instead
 */
export const DEVELOPER_TENANTS = DEVELOPER_ROLES;
