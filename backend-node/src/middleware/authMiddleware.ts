import { Request, Response, NextFunction } from "express";
import { JWTService } from "../services/jwtService";

declare global {
  namespace Express {
    interface Request {
      userId?: string;
      username?: string;
      role?: string;
      tenantId?: string;
      isServiceAccount?: boolean;
      serviceId?: string;
    }
  }
}

const jwtService = new JWTService();

/**
 * Auth middleware - extracts and verifies JWT token
 */
export function authMiddleware(
  req: Request,
  res: Response,
  next: NextFunction
): void {
  try {
    const authHeader = req.headers.authorization;

    if (!authHeader) {
      res.status(401).json({ error: "No token provided" });
      return;
    }

    const token = jwtService.extractToken(authHeader);

    if (!token) {
      res.status(401).json({ error: "Invalid token format" });
      return;
    }

    const payload = jwtService.verifyToken(token);

    req.userId = payload.userId;
    req.username = payload.username;
    req.role = payload.role;
    req.tenantId = payload.tenantId;

    // Handle service accounts (n8n, etc)
    if (payload.role === "service") {
      req.isServiceAccount = true;
      req.serviceId = (payload as any).service || "unknown";

      // Service accounts can specify tenant via header
      const headerTenantId = req.headers["x-tenant-id"] as string;
      if (headerTenantId) {
        req.tenantId = headerTenantId;
      }
    }

    next();
  } catch (error: any) {
    res.status(401).json({ error: error.message });
  }
}

export function roleMiddleware(...allowedRoles: string[]) {
  return (req: Request, res: Response, next: NextFunction): void => {
    const userRole = req.role;

    if (!userRole || !allowedRoles.includes(userRole)) {
      res.status(403).json({ error: "Forbidden - Insufficient permissions" });
      return;
    }

    next();
  };
}

export function requireAuth(
  req: Request,
  res: Response,
  next: NextFunction
): void {
  if (!req.userId) {
    res.status(401).json({ error: "Authentication required" });
    return;
  }
  next();
}
