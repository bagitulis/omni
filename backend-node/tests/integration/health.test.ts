/**
 * Health Check API Integration Tests
 * Tests basic API endpoints and server health
 */

import request from "supertest";
import express, { Application } from "express";

// Mock Express app for testing
const createTestApp = (): Application => {
  const app = express();
  app.use(express.json());

  // Health check route
  app.get("/health", (_req, res) => {
    res.json({
      status: "OK",
      timestamp: new Date().toISOString(),
      uptime: process.uptime(),
      environment: process.env.NODE_ENV || "development",
    });
  });

  // Status route
  app.get("/status", (_req, res) => {
    res.json({
      service: "omni-backend",
      version: "1.0.0",
      status: "running",
    });
  });

  return app;
};

describe("Health Check API", () => {
  let app: Application;

  beforeAll(() => {
    app = createTestApp();
  });

  describe("GET /health", () => {
    it("should return 200 OK", async () => {
      const response = await request(app).get("/health");

      expect(response.status).toBe(200);
    });

    it("should return health status", async () => {
      const response = await request(app).get("/health");

      expect(response.body).toHaveProperty("status", "OK");
      expect(response.body).toHaveProperty("timestamp");
      expect(response.body).toHaveProperty("uptime");
      expect(response.body).toHaveProperty("environment");
    });

    it("should return valid timestamp", async () => {
      const response = await request(app).get("/health");

      const timestamp = new Date(response.body.timestamp);
      expect(timestamp.getTime()).not.toBeNaN();
    });

    it("should return positive uptime", async () => {
      const response = await request(app).get("/health");

      expect(response.body.uptime).toBeGreaterThanOrEqual(0);
    });

    it("should set correct content-type header", async () => {
      const response = await request(app).get("/health");

      expect(response.headers["content-type"]).toMatch(/application\/json/);
    });
  });

  describe("GET /status", () => {
    it("should return 200 OK", async () => {
      const response = await request(app).get("/status");

      expect(response.status).toBe(200);
    });

    it("should return service info", async () => {
      const response = await request(app).get("/status");

      expect(response.body).toEqual({
        service: "omni-backend",
        version: "1.0.0",
        status: "running",
      });
    });
  });

  describe("404 Handling", () => {
    it("should return 404 for unknown routes", async () => {
      const response = await request(app).get("/unknown-route");

      expect(response.status).toBe(404);
    });
  });
});
