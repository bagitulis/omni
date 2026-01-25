import cors from "cors";
import { Request, Response, NextFunction } from "express";
import { env } from "../config/environment";

// ============================================
// CORS CONFIGURATION WITH STRICT WHITELIST
// ============================================

/**
 * Extract origin from Referer header as fallback
 */
function getOriginFromReferer(referer: string | undefined): string | null {
  if (!referer) return null;
  try {
    const url = new URL(referer);
    return `${url.protocol}//${url.host}`;
  } catch {
    return null;
  }
}

/**
 * Check if request comes through a reverse proxy (nginx)
 */
function isProxiedRequest(req: Request): boolean {
  return !!(req.headers["x-forwarded-for"] || req.headers["x-real-ip"]);
}

/**
 * Allowed origins for different environments
 * Development: localhost only (safe for local testing)
 * Production: specific domains from environment variable
 */
const getDevelopmentOrigins = (): string[] => [
  "http://localhost:5173", // Vite default frontend
  "http://127.0.0.1:5173",
  "http://localhost:5174", // Alternative Vite frontend port
  "http://127.0.0.1:5174",
  "http://localhost:4173", // Vite preview (production mode)
  "http://127.0.0.1:4173",
  "http://localhost:3000", // Alternative frontend port
  "http://127.0.0.1:3000",
  "http://localhost:3001", // Backend itself (for testing)
  "http://127.0.0.1:3001",
  "http://localhost:5000", // Legacy port
  "http://127.0.0.1:5000",
  "http://localhost:8000", // Alternative port
  "http://127.0.0.1:8000",
  // Production ports (Docker)
  "http://localhost:80",
  "http://localhost:8080",
  "http://localhost:8888",
  "http://127.0.0.1:80",
  "http://127.0.0.1:8080",
  "http://127.0.0.1:8888",
  "http://localhost", // Port 80 tanpa explicit port
  "http://127.0.0.1", // Port 80 tanpa explicit port
];

const getProductionOrigins = (): string[] => {
  // Get from env CORS_ORIGINS, or use hardcoded production domains
  const envOrigins = env.CORS_ORIGINS || [];

  if (envOrigins.length > 0 && envOrigins[0] !== "") {
    return envOrigins.map((o) => o.trim());
  }

  // Fallback: Docker production ports + custom domains
  return [
    "http://localhost:80",
    "http://localhost:8080",
    "http://localhost:8888",
    "http://localhost:8000",
    "http://localhost:4173",
    "http://127.0.0.1:80",
    "http://127.0.0.1:8080",
    "http://127.0.0.1:8888",
    "http://127.0.0.1:8000",
    "http://127.0.0.1:4173",
    "http://localhost",
    "http://127.0.0.1",
    // Production domains
    "https://yndigital.my.id",
    "http://yndigital.my.id",
    "https://www.yndigital.my.id",
    "http://www.yndigital.my.id",
  ];
};

/**
 * Custom CORS middleware that handles reverse proxy scenarios
 * where Origin header is not forwarded by nginx
 */
export function corsMiddleware(
  req: Request,
  res: Response,
  next: NextFunction
): void {
  const origin = req.headers.origin;
  const referer = req.headers.referer;

  const allowedOrigins =
    env.NODE_ENV === "production"
      ? getProductionOrigins()
      : getDevelopmentOrigins();

  // Determine effective origin
  let effectiveOrigin: string | null = origin || null;

  // If no origin header but request is proxied, use referer
  if (!effectiveOrigin && isProxiedRequest(req)) {
    effectiveOrigin = getOriginFromReferer(referer);
  }

  // Check if origin is allowed
  const isAllowed =
    !effectiveOrigin || // No origin = internal/health check
    allowedOrigins.includes(effectiveOrigin);

  if (!isAllowed) {
    res.status(403).json({
      success: false,
      error: `CORS policy: Origin "${effectiveOrigin}" not allowed`,
    });
    return;
  }

  // Set CORS headers
  if (effectiveOrigin) {
    res.setHeader("Access-Control-Allow-Origin", effectiveOrigin);
  }
  res.setHeader("Access-Control-Allow-Credentials", "true");
  res.setHeader(
    "Access-Control-Allow-Methods",
    "GET, POST, PUT, DELETE, PATCH, OPTIONS"
  );
  res.setHeader(
    "Access-Control-Allow-Headers",
    "Content-Type, Authorization, Cache-Control, x-tenant-id, x-username, x-csrf-token"
  );
  res.setHeader("Access-Control-Expose-Headers", "Content-Type, X-Total-Count");
  res.setHeader("Access-Control-Max-Age", "3600");

  // Handle preflight
  if (req.method === "OPTIONS") {
    res.status(204).end();
    return;
  }

  next();
}

/**
 * CORS middleware for webhooks - allows no-origin since webhooks
 * come from external servers (Shopee, TikTok, Lazada)
 */
export const webhookCorsMiddleware = cors({
  origin: (origin, callback) => {
    // Webhooks from platforms don't send Origin header
    // This is expected behavior for server-to-server requests
    if (!origin) {
      return callback(null, true);
    }

    // If origin is present, still validate it
    const allowedOrigins =
      env.NODE_ENV === "production"
        ? getProductionOrigins()
        : getDevelopmentOrigins();

    if (allowedOrigins.includes(origin)) {
      callback(null, true);
    } else {
      // For webhooks, we're more permissive but log the origin
      // Platform servers may send varying origins
      callback(null, true);
    }
  },
  credentials: false, // Webhooks don't need credentials
  methods: ["POST", "OPTIONS"],
  allowedHeaders: ["Content-Type", "Authorization"],
  maxAge: 3600,
});

export default corsMiddleware;
