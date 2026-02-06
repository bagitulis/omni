import api from "./api";

interface LoginResponse {
  token: string;
  tenant_id: string; // snake_case from backend per AGENTS.MD
  user: {
    id: string;
    username: string;
    email: string;
    role: string;
  };
  requiresCaptcha?: boolean;
  isLocked?: boolean;
  lockMinutesRemaining?: number;
}

interface LoginErrorResponse {
  error: string;
  code?: string;
  requiresCaptcha?: boolean;
  isLocked?: boolean;
  lockMinutesRemaining?: number;
}

interface RegisterResponse extends LoginResponse {}

class AuthService {
  /**
   * Login with optional reCAPTCHA v3 token
   *
   * Flow:
   * - 1-3 failed attempts: No CAPTCHA needed
   * - 4-10 failed attempts: reCAPTCHA token required
   * - 10+ failed attempts: Account locked for 30 minutes
   *
   * @param username - Username
   * @param password - Password
   * @param recaptchaToken - reCAPTCHA v3 token (only required after 3 failed attempts)
   */
  async login(
    username: string,
    password: string,
    recaptchaToken?: string,
  ): Promise<LoginResponse> {
    try {
      const payload: Record<string, string> = { username, password };

      if (recaptchaToken) {
        payload.recaptchaToken = recaptchaToken;
      }

      const response = await api.post<LoginResponse>("/auth/login", payload);
      return response;
    } catch (error: any) {
      const errorData = error.response?.data as LoginErrorResponse;
      const err = new Error(errorData?.error || "Login failed") as any;
      err.code = errorData?.code;
      err.requiresCaptcha = errorData?.requiresCaptcha;
      err.isLocked = errorData?.isLocked;
      err.lockMinutesRemaining = errorData?.lockMinutesRemaining;
      throw err;
    }
  }

  async register(
    username: string,
    email: string,
    password: string,
  ): Promise<RegisterResponse> {
    try {
      const response = await api.post<RegisterResponse>("/auth/register", {
        username,
        email,
        password,
      });
      return response;
    } catch (error: any) {
      throw new Error(error.response?.data?.error || "Registration failed");
    }
  }

  async getCurrentUser() {
    try {
      const response = await api.get("/auth/me");
      return response.data;
    } catch {
      return null;
    }
  }

  async logout() {
    try {
      await api.post("/auth/logout");
    } catch (error) {
      console.error("Logout error:", error);
    } finally {
      // Always clear ALL auth-related data from localStorage
      localStorage.removeItem("authToken");
      localStorage.removeItem("authUser");
      localStorage.removeItem("tenantId");
      localStorage.removeItem("userRole");
      localStorage.removeItem("userName");
    }
  }
}

export const authService = new AuthService();
