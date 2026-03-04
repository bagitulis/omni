/**
 * Google reCAPTCHA Service
 * Server-side verification for Google reCAPTCHA v2
 *
 * Security measures:
 * 1. Token verified server-side with Google's API
 * 2. Secret key never exposed to client
 * 3. Each token is one-time use (verified by Google)
 * 4. IP verification for additional security
 */

import { getLogger } from "../utils/logger";

const logger = getLogger("RecaptchaService");

// Google reCAPTCHA verification endpoint
const RECAPTCHA_VERIFY_URL = "https://www.google.com/recaptcha/api/siteverify";

// Get keys from environment
const RECAPTCHA_SECRET_KEY = process.env.RECAPTCHA_SECRET_KEY || "";
const RECAPTCHA_SITE_KEY = process.env.RECAPTCHA_SITE_KEY || "";

// Minimum score for reCAPTCHA v3 (0.0 - 1.0, higher = more likely human)
const MIN_SCORE_V3 = 0.5;

interface RecaptchaVerifyResponse {
  success: boolean;
  challenge_ts?: string;
  hostname?: string;
  score?: number; // v3 only
  action?: string; // v3 only
  "error-codes"?: string[];
}

export class CaptchaService {
  private static instance: CaptchaService;

  static getInstance(): CaptchaService {
    if (!CaptchaService.instance) {
      CaptchaService.instance = new CaptchaService();
    }
    return CaptchaService.instance;
  }

  /**
   * Get the reCAPTCHA site key for frontend
   */
  getSiteKey(): string {
    return RECAPTCHA_SITE_KEY;
  }

  /**
   * Check if reCAPTCHA is configured
   */
  isConfigured(): boolean {
    return !!RECAPTCHA_SECRET_KEY && !!RECAPTCHA_SITE_KEY;
  }

  /**
   * Verify reCAPTCHA token with Google
   *
   * @param token - reCAPTCHA response token from frontend
   * @param ipAddress - Client IP address (optional, for additional verification)
   * @param expectedAction - Expected action for v3 (optional)
   */
  async verifyCaptcha(
    token: string,
    ipAddress?: string,
    expectedAction?: string
  ): Promise<{ valid: boolean; error?: string; score?: number }> {
    // Check if reCAPTCHA is configured
    if (!RECAPTCHA_SECRET_KEY) {
      logger.warn(
        "reCAPTCHA secret key not configured - skipping verification"
      );
      // In development without config, allow through with warning
      if (process.env.NODE_ENV !== "production") {
        return { valid: true };
      }
      return { valid: false, error: "reCAPTCHA not configured" };
    }

    // Validate input
    if (!token) {
      return { valid: false, error: "reCAPTCHA token is required" };
    }

    try {
      // Build verification request
      const params = new URLSearchParams({
        secret: RECAPTCHA_SECRET_KEY,
        response: token,
      });

      if (ipAddress) {
        params.append("remoteip", ipAddress);
      }

      // Call Google's verification API
      const response = await fetch(RECAPTCHA_VERIFY_URL, {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: params.toString(),
      });

      if (!response.ok) {
        logger.error(`reCAPTCHA API error: ${response.status}`);
        return { valid: false, error: "reCAPTCHA verification failed" };
      }

      const result = (await response.json()) as RecaptchaVerifyResponse;

      // Check basic success
      if (!result.success) {
        const errorCodes = result["error-codes"]?.join(", ") || "unknown";
        logger.warn(`reCAPTCHA verification failed: ${errorCodes}`);
        return { valid: false, error: "reCAPTCHA verification failed" };
      }

      // For v3, check score
      if (result.score !== undefined) {
        if (result.score < MIN_SCORE_V3) {
          logger.warn(`reCAPTCHA score too low: ${result.score}`);
          return {
            valid: false,
            error: "Suspicious activity detected",
            score: result.score,
          };
        }

        // Check action matches (v3)
        if (expectedAction && result.action !== expectedAction) {
          logger.warn(
            `reCAPTCHA action mismatch: expected ${expectedAction}, got ${result.action}`
          );
          return { valid: false, error: "reCAPTCHA action mismatch" };
        }
      }

      logger.info(
        `reCAPTCHA verified successfully${
          result.score !== undefined ? ` (score: ${result.score})` : ""
        }`
      );

      return { valid: true, score: result.score };
    } catch (error) {
      logger.error("reCAPTCHA verification error:", error);
      return { valid: false, error: "reCAPTCHA verification error" };
    }
  }

  /**
   * Get CAPTCHA stats (for monitoring)
   */
  getStats(): { configured: boolean; siteKey: string } {
    return {
      configured: this.isConfigured(),
      siteKey: RECAPTCHA_SITE_KEY ? "***configured***" : "not set",
    };
  }
}

export default CaptchaService.getInstance();
