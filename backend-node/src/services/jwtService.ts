import jwt from "jsonwebtoken";
import { env } from "../config/environment";

export interface JWTPayload {
  userId: string;
  username: string;
  role: string;
  tenantId: string;
  iat?: number;
  exp?: number;
}

export class JWTService {
  private secret: string;
  private expiresIn: string = "24h";

  constructor() {
    // Use validated secret from environment (no fallback)
    this.secret = env.JWT_SECRET;
  }

  /**
   * Generate JWT token
   */
  generateToken(payload: Omit<JWTPayload, "iat" | "exp">): string {
    return jwt.sign(payload, this.secret, {
      expiresIn: this.expiresIn,
    } as any);
  }

  /**
   * Verify JWT token
   */
  verifyToken(token: string): JWTPayload {
    try {
      return jwt.verify(token, this.secret as jwt.Secret) as JWTPayload;
    } catch (error: any) {
      throw new Error(`Invalid token: ${error.message}`);
    }
  }

  /**
   * Extract token from Bearer header
   */
  extractToken(authHeader?: string): string | null {
    if (!authHeader) return null;

    const parts = authHeader.split(" ");
    if (parts.length !== 2 || parts[0] !== "Bearer") {
      return null;
    }

    return parts[1];
  }

  /**
   * Decode token without verification (for debugging)
   */
  decodeToken(token: string): JWTPayload | null {
    try {
      return jwt.decode(token) as JWTPayload;
    } catch {
      return null;
    }
  }
}
