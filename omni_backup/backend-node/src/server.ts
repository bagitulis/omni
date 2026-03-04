/**
 * Main Server Entry Point
 * Simplified and modular Express server configuration
 */
import express, { Express, Request, Response } from "express";
import { env } from "./config/environment";
import { backendLogger } from "./utils/backendLogger";
import { errorHandler } from "./middleware/errorHandler";
import { configureMiddleware, registerRoutes } from "./config/routeConfig";
import { initializeServices } from "./config/serviceInitializer";

const app: Express = express();

/**
 * Configure Express middleware
 */
configureMiddleware(app);

/**
 * Initialize backend services on startup
 */
initializeServices().catch((error) => {
  backendLogger.error("Failed to initialize services:", error);
  process.exit(1);
});

/**
 * Register all application routes
 */
registerRoutes(app);

/**
 * Root API endpoint - API documentation
 */
app.get("/api", (_req: Request, res: Response) => {
  res.json({
    name: "Omni Backend API",
    version: "1.0.0",
    description: "Multi-tenant E-commerce Platform Integration API",
    status: "operational",
    endpoints: {
      health: "/api/health",
      auth: {
        login: "POST /api/auth/login",
        register: "POST /api/auth/register",
        tenants: "GET /api/auth/tenants",
      },
      orders: {
        sync: "POST /api/orders/sync/:category",
        list: "GET /api/orders/:platform",
      },
      products: {
        list: "GET /api/products/:platform",
        sync: "POST /api/products/:platform/sync",
      },
      inventory: {
        list: "GET /api/inventory/data",
        update: "PUT /api/inventory/data/:id",
      },
      monitoring: {
        metrics: "GET /api/monitoring/metrics",
        health: "GET /api/monitoring/health",
      },
    },
    documentation: "https://github.com/your-repo/omni-backend",
  });
});

/**
 * 404 handler
 */
app.use((req: Request, res: Response) => {
  res.status(404).json({
    success: false,
    error: {
      message: "Route not found",
      code: "NOT_FOUND",
      statusCode: 404,
      path: req.path,
      method: req.method,
    },
  });
});

/**
 * Global error handler (must be last middleware)
 */
app.use(errorHandler);

/**
 * Start server
 */
const PORT = env.PORT;
const server = app.listen(PORT, "0.0.0.0", () => {
  backendLogger.success("Server", "=".repeat(60));
  backendLogger.success("Server", "🚀 OMNI BACKEND SERVER STARTED");
  backendLogger.success("Server", "=".repeat(60));
  backendLogger.info("Server", `Port: ${PORT}`);
  backendLogger.info("Server", `Environment: ${env.NODE_ENV}`);
  backendLogger.info("Server", `CORS Origins: ${env.CORS_ORIGINS.join(", ")}`);
  backendLogger.success("Server", "=".repeat(60));
});

/**
 * Graceful shutdown handlers
 */
const gracefulShutdown = (signal: string) => {
  backendLogger.warning("Server", `${signal} received, shutting down gracefully`);
  server.close(() => {
    backendLogger.info("Server", "Server closed successfully");
    process.exit(0);
  });
  
  // Force shutdown after 10 seconds
  setTimeout(() => {
    backendLogger.error("Server", "Forced shutdown after timeout");
    process.exit(1);
  }, 10000);
};

process.on("SIGTERM", () => gracefulShutdown("SIGTERM"));
process.on("SIGINT", () => gracefulShutdown("SIGINT"));

export default app;
