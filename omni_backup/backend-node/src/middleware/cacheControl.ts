import { Request, Response, NextFunction } from "express";

/**
 * Middleware to disable caching for auth-related endpoints
 * Prevents 304 Not Modified responses for dynamic content
 */
export function noCacheMiddleware(
  _req: Request,
  res: Response,
  next: NextFunction
): void {
  res.set({
    "Cache-Control": "no-store, no-cache, must-revalidate, proxy-revalidate",
    Pragma: "no-cache",
    Expires: "0",
    "Surrogate-Control": "no-store",
  });
  next();
}
