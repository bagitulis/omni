/**
 * Google reCAPTCHA v3 Service (Frontend)
 * Invisible reCAPTCHA with scoring - no user interaction needed
 */

import api from "./api";

interface SiteKeyResponse {
  success: boolean;
  siteKey: string | null;
  configured: boolean;
}

class CaptchaService {
  private siteKey: string | null = null;
  private configured: boolean = false;
  private scriptLoaded: boolean = false;
  private loadingPromise: Promise<boolean> | null = null;

  /**
   * Get reCAPTCHA site key from backend
   */
  async getSiteKey(): Promise<string | null> {
    if (this.siteKey) return this.siteKey;

    try {
      const response = await api.get<SiteKeyResponse>("/captcha/site-key");
      this.siteKey = response.siteKey;
      this.configured = response.configured;
      return this.siteKey;
    } catch (error) {
      console.error("Failed to get reCAPTCHA site key:", error);
      return null;
    }
  }

  /**
   * Check if reCAPTCHA is configured
   */
  isConfigured(): boolean {
    return this.configured;
  }

  /**
   * Load reCAPTCHA v3 script dynamically
   */
  async loadScript(): Promise<boolean> {
    // Return existing promise if loading
    if (this.loadingPromise) return this.loadingPromise;
    if (this.scriptLoaded) return true;

    this.loadingPromise = (async () => {
      const siteKey = await this.getSiteKey();
      if (!siteKey) {
        console.warn("reCAPTCHA site key not available");
        return false;
      }

      return new Promise<boolean>((resolve) => {
        // Check if already loaded (grecaptcha exists and has execute method)
        if (
          typeof window.grecaptcha !== "undefined" &&
          typeof window.grecaptcha.execute === "function"
        ) {
          this.scriptLoaded = true;
          resolve(true);
          return;
        }

        const script = document.createElement("script");
        // v3 uses render parameter with site key
        script.src = `https://www.google.com/recaptcha/api.js?render=${siteKey}`;
        script.async = true;
        script.defer = true;

        script.onload = () => {
          this.scriptLoaded = true;
          resolve(true);
        };

        script.onerror = () => {
          console.error("Failed to load reCAPTCHA script");
          resolve(false);
        };

        document.head.appendChild(script);
      });
    })();

    return this.loadingPromise;
  }

  /**
   * Execute reCAPTCHA v3 and get token
   * This is invisible - no user interaction required
   *
   * @param action - Action name for analytics (e.g., 'login', 'register')
   * @returns Token string or null if failed
   */
  async execute(action: string = "login"): Promise<string | null> {
    try {
      await this.loadScript();

      const siteKey = this.siteKey;
      if (!siteKey || !window.grecaptcha) {
        console.error("reCAPTCHA not ready");
        return null;
      }

      return new Promise((resolve) => {
        window.grecaptcha.ready(() => {
          window.grecaptcha
            .execute(siteKey, { action })
            .then((token: string) => {
              resolve(token);
            })
            .catch((error: Error) => {
              console.error("reCAPTCHA execute error:", error);
              resolve(null);
            });
        });
      });
    } catch (error) {
      console.error("reCAPTCHA error:", error);
      return null;
    }
  }

  /**
   * Initialize reCAPTCHA (call on app load or login page mount)
   */
  async init(): Promise<boolean> {
    return this.loadScript();
  }
}

// Add grecaptcha v3 types to window
declare global {
  interface Window {
    grecaptcha: {
      ready: (callback: () => void) => void;
      execute: (
        siteKey: string,
        options: { action: string }
      ) => Promise<string>;
    };
  }
}

export const captchaService = new CaptchaService();
