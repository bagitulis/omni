/**
 * Route Configuration Module
 * Centralizes all route imports and middleware setup
 */
import express, { Express } from "express";
import helmet from "helmet";
import compression from "compression";
import cookieParser from "cookie-parser";
import { corsMiddleware } from "../middleware/cors";
import { enforceHttps } from "../middleware/httpsRedirect";
import { tenantMiddleware } from "../middleware/tenantMiddleware";
import { createRawBodyParser } from "../middleware/rawBody";
import {
  csrfProtection,
  csrfTokenEndpoint,
} from "../middleware/csrfProtection";

// Route imports
import healthRoutes from "../routes/health";
import tokenStatusRoutes from "../routes/tokenStatus";
import tokenOperationRoutes from "../controllers/tokenOperationController";
import orderSyncRoutes from "../controllers/orderSyncController";
import settingsRoutes from "../routes/settings";
import googleSheetsRoutes from "../routes/googleSheets";
import spreadsheetRegistryRoutes from "../routes/spreadsheetRegistry";
import inventorySheetRouter from "../routes/inventorySheet";
import inventoryColumnsRouter from "../routes/inventoryColumnsRoutes";
import inventoryDataRouter from "../routes/inventoryDataRoutes";
import inventoryConfigRouter from "../routes/inventoryConfigRoutes";
import inventorySyncRouter from "../routes/inventorySyncRoutes";
import routeMappingRoutes from "../routes/routeMapping";
import productRoutes from "../routes/products";
import productCreateRoutes from "../routes/productCreate";
import skuCheckRoutes from "../routes/skuCheckRoutes";
import skuBatchCheckRoutes from "../routes/skuBatchCheckRoutes";
import productCloneRoutes from "../routes/productCloneRoutes";
import filterPreferenceRoutes from "../routes/filterPreference";
import stockRoutes from "../routes/stockRoutes";
import priceRoutes from "../routes/priceRoutes";
import shopeeWalletRouter from "../routes/shopeeWalletRoutes";
import shopeeShippingRouter from "../routes/shopeeShippingRoutes";
import shopeeOperationsRouter from "../routes/shopeeOperationsRoutes";
import { walletSheetsRouter } from "../routes/walletSheetsRoutes";
import { shippingFeeSheetsRouter } from "../routes/shippingFeeSheetsRoutes";
import monitoringRoutes from "../routes/monitoringRoutes";
import jobQueueRouter from "../routes/jobQueueRoutes";
import autoFunctionRouter from "../routes/autoFunctionRoutes";
import routeConfigRouter from "../routes/routeConfigRoutes";
import { routeExecutionConfigRouter } from "../routes/routeExecutionConfigRoutes";
import authRoutes from "../routes/authRoutes";
import userManagementRouter from "../routes/userManagementRoutes";
import auditRouter from "../routes/auditRoutes";
import googleServiceAccountRoutes from "../routes/googleServiceAccountRoutes";
import sheetConfigRoutes from "../routes/sheetConfigRoutes";
import wholesaleRoutes from "../routes/wholesaleRoutes";
import analyticsRouter from "../routes/analyticsRoutes";
import tiktokAnalyticsRouter from "../routes/tiktokAnalyticsRoutes";
import tiktokAdsRouter from "../routes/tiktokAdsRoutes";
import shopeeAdsRouter from "../routes/shopeeAdsRoutes";
import adsReportsRouter from "../routes/adsReportsRoutes";
import platformAuthRoutes from "../routes/platformAuthRoutes";
import webhookRoutes from "../routes/webhookRoutes";
import shopSetupRoutes from "../routes/shopSetupRoutes";
import n8nWebhookRoutes from "../routes/n8nWebhookRoutes";
import securityRoutes from "../routes/securityRoutes";
import captchaRoutes from "../routes/captchaRoutes";

/**
 * Configure express middleware
 */
export function configureMiddleware(app: Express): void {
  // Security middleware
  app.use(helmet());
  app.use(compression());

  // Health check BEFORE CORS - allows Docker health probes and monitoring
  app.use("/api", healthRoutes);

  // CORS
  app.use(corsMiddleware);

  // HTTPS enforcement (production only)
  if (process.env.NODE_ENV === "production") {
    app.use(enforceHttps);
  }

  // Cookie parser (required for CSRF)
  app.use(cookieParser());

  // Body parsers with raw body capture for webhooks
  app.use(createRawBodyParser());
  app.use(express.urlencoded({ extended: true, limit: "10mb" }));

  // CSRF Protection (after body parser, before routes)
  // Production only - can be enabled in dev with CSRF_ENABLED=true
  if (
    process.env.NODE_ENV === "production" ||
    process.env.CSRF_ENABLED === "true"
  ) {
    app.use(csrfProtection);
  }

  // Tenant context middleware (must be before routes)
  app.use(tenantMiddleware);
}

/**
 * Register all application routes
 */
export function registerRoutes(app: Express): void {
  // CSRF Token endpoint (must be before CSRF protection kicks in)
  app.get("/api/csrf-token", csrfTokenEndpoint);

  // CAPTCHA endpoint (no auth required, before login)
  app.use("/api/captcha", captchaRoutes);

  // Token Status (health already mounted before CORS for docker probes)
  app.use("/api", tokenStatusRoutes);

  // Webhooks - MUST be before productRoutes which has router.use(authMiddleware)
  // External platform webhooks don't require authentication
  app.use("/api/webhooks", webhookRoutes);

  // n8n Integration - No auth required (uses secret key)
  app.use("/api", n8nWebhookRoutes);

  // Platform OAuth - MUST be before productRoutes (no auth required for callbacks)
  app.use("/api/platform-auth", platformAuthRoutes);

  // Authentication & Authorization
  app.use("/api/auth", authRoutes);
  app.use("/api/admin/users", userManagementRouter); // Changed from /users to /admin/users
  app.use("/api/admin/shop-setup", shopSetupRoutes); // Shop setup and credentials
  app.use("/api/audit", auditRouter);

  // Platform Integration
  app.use("/api/token", tokenOperationRoutes);
  app.use("/api/orders", orderSyncRoutes);
  app.use("/api/settings", settingsRoutes);

  // Google Sheets Integration
  app.use("/api/google", googleSheetsRoutes); // Changed from /google-sheets to /google
  app.use("/api/google/service-accounts", googleServiceAccountRoutes);
  app.use("/api/google/registry", spreadsheetRegistryRoutes); // Changed from /spreadsheet-registry to /google/registry
  app.use("/api/google/sheet-config", sheetConfigRoutes); // Moved to google namespace for consistency

  // Inventory Management - Order matters! More specific routes first
  app.use("/api/inventory/sheet", inventorySheetRouter);
  app.use("/api/inventory/columns", inventoryColumnsRouter);
  app.use("/api/inventory/data", inventoryDataRouter);
  app.use("/api/inventory/config", inventoryConfigRouter);
  app.use("/api/inventory/sync", inventorySyncRouter);
  // SKU batch check routes (platform-status, batch-check-sku, etc.)
  app.use("/api", skuBatchCheckRoutes);
  // inventoryDataRouter last - has /:keyValue catch-all pattern
  app.use("/api/inventory", inventoryDataRouter);

  // Product Management
  app.use("/api", productRoutes); // Changed from /products to /api (routes already include platform prefix)
  app.use("/api", productCreateRoutes); // Shopee & Lazada create product routes
  app.use("/api", productCloneRoutes); // Product cloning between platforms
  app.use("/api/products/sku-check", skuCheckRoutes); // Nested under products for better organization
  app.use("/api/wholesale", wholesaleRoutes);

  // Operations - Stock and Price updates
  // Mount on both original paths AND /api/inventory for frontend compatibility
  // Frontend consistently uses /api/inventory as base for all inventory operations
  app.use("/api/stock", stockRoutes);
  app.use("/api/inventory", stockRoutes); // Stock: /api/inventory/update-stock, update-stock-batch
  app.use("/api/price", priceRoutes);
  app.use("/api/inventory", priceRoutes); // Price: /api/inventory/update-price, update-price-batch
  app.use("/api/filter-preferences", filterPreferenceRoutes); // Standardized to plural

  // Shopee Specific
  app.use("/api/shopee/wallet", shopeeWalletRouter);
  app.use("/api/shopee/shipping", shopeeShippingRouter);
  app.use("/api/shopee/operations", shopeeOperationsRouter);

  // Sheets Integration
  app.use("/api/sheets/operations", walletSheetsRouter); // Sheets operations namespace
  app.use("/api/sheets/operations", shippingFeeSheetsRouter); // Sheets operations namespace

  // Analytics
  app.use("/api/analytics", analyticsRouter);
  app.use("/api/analytics/tiktok", tiktokAnalyticsRouter);
  app.use("/api/analytics/tiktok-ads", tiktokAdsRouter);
  app.use("/api/analytics/shopee-ads", shopeeAdsRouter);
  app.use("/api/reports", adsReportsRouter);

  // System Management
  app.use("/api/route-mapping", routeMappingRoutes);
  app.use("/api/routes-config", routeConfigRouter); // Changed from /route-config to /routes-config (plural)
  app.use("/api/route-execution-config", routeExecutionConfigRouter);
  app.use("/api/jobs", jobQueueRouter); // Changed from /job-queue to /jobs
  app.use("/api/jobs/auto-functions", autoFunctionRouter); // Changed from /auto-function to /jobs/auto-functions
  app.use("/api/monitoring", monitoringRoutes);

  // Security Monitoring
  app.use("/api/security", securityRoutes);
}
