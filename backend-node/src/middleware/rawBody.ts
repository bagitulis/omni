/**
 * Raw Body Middleware
 * Captures raw request body before JSON parsing for webhook signature verification
 *
 * Use Case:
 * - Shopee/TikTok/Lazada webhooks sign the raw body bytes
 * - express.json() parses body first, making signature verification fail
 * - This middleware saves raw body to req.rawBody for later verification
 */
import { Request, Response, NextFunction } from "express";
import { getLogger } from "../utils/logger";

const logger = getLogger("RawBodyMiddleware");

// Extend Express Request type to include rawBody
declare global {
  namespace Express {
    interface Request {
      rawBody?: string;
    }
  }
}

/**
 * Middleware to capture raw body for specific routes
 * Only applies to webhook routes to avoid memory overhead
 */
export function rawBodyMiddleware(
  req: Request,
  _res: Response,
  next: NextFunction
): void {
  // Only capture raw body for webhook routes
  if (!req.path.includes("/webhook")) {
    return next();
  }

  // Only process JSON content type
  if (
    req.headers["content-type"] &&
    !req.headers["content-type"].includes("application/json")
  ) {
    return next();
  }

  let rawBody = "";

  req.on("data", (chunk: Buffer) => {
    rawBody += chunk.toString("utf8");
  });

  req.on("end", () => {
    if (rawBody) {
      req.rawBody = rawBody;
      logger.debug(`Raw body captured (${rawBody.length} bytes)`);
    }
    next();
  });
}

/**
 * Alternative: express.json with verify function
 * This is more efficient as it hooks into body-parser
 */
export function createRawBodyParser() {
  const express = require("express");

  return express.json({
    limit: "10mb",
    verify: (req: Request, _res: Response, buf: Buffer, _encoding: string) => {
      // Use originalUrl as path might not include full route at middleware level
      const fullPath = req.originalUrl || req.path || "";

      // Only save raw body for webhook routes
      // Match both /webhook and /webhooks patterns
      if (fullPath.includes("/webhook")) {
        req.rawBody = buf.toString("utf8");
        logger.debug(
          `Raw body captured (${buf.length} bytes) for: ${fullPath}`
        );
      }
    },
  });
}
