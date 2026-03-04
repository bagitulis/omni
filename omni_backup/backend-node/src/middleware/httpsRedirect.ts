import { Request, Response, NextFunction } from "express";

/**
 * HTTPS Redirect Middleware
 * Enforces HTTPS in production environment
 * Redirects HTTP requests to HTTPS
 *
 * Can be disabled with DISABLE_HTTPS_REDIRECT=true for Docker/reverse proxy setups
 */
export const enforceHttps = (
  req: Request,
  res: Response,
  next: NextFunction
): void => {
  // Skip if disabled (for Docker/reverse proxy setups)
  if (process.env.DISABLE_HTTPS_REDIRECT === "true") {
    return next();
  }

  // Only enforce in production
  if (process.env.NODE_ENV !== "production") {
    return next();
  }

  // Check if request is secure
  const isSecure =
    req.secure ||
    req.get("x-forwarded-proto") === "https" ||
    req.get("x-forwarded-ssl") === "on";

  if (!isSecure) {
    // Redirect to HTTPS
    const httpsUrl = `https://${req.hostname}${req.url}`;
    return res.redirect(301, httpsUrl);
  }

  next();
};
