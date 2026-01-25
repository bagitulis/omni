/**
 * Shared Authentication Service for E2E Tests
 * Prevents rate limiting by reusing a single auth token
 */
import axios from "axios";

class SharedAuthService {
  private static instance: SharedAuthService;
  private token: string | null = null;
  private tokenExpiry: number = 0;
  private readonly baseUrl = "http://localhost:3000/api";
  private readonly credentials = { username: "yumna", password: "password123" };

  private constructor() {}

  static getInstance(): SharedAuthService {
    if (!SharedAuthService.instance) {
      SharedAuthService.instance = new SharedAuthService();
    }
    return SharedAuthService.instance;
  }

  async getToken(): Promise<string> {
    if (this.token && Date.now() < this.tokenExpiry) {
      return this.token;
    }

    try {
      const response = await axios.post(
        `${this.baseUrl}/auth/login`,
        this.credentials
      );
      this.token = response.data.token;
      this.tokenExpiry = Date.now() + 3600000; // 1 hour
      console.log("✅ Shared authentication successful");
      return this.token!;
    } catch (error: any) {
      throw new Error(`Auth failed: ${error.message}`);
    }
  }

  clearToken(): void {
    this.token = null;
    this.tokenExpiry = 0;
  }
}

export const sharedAuth = SharedAuthService.getInstance();
