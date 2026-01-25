/**
 * Auth API Integration Tests
 * Tests authentication endpoints
 */

import request from "supertest";
import express, { Application } from "express";
import { AuthService } from "../../src/services/authService";

// Mock AuthService
jest.mock("../../src/services/authService");

const createAuthApp = (): Application => {
  const app = express();
  app.use(express.json());

  // Register route
  app.post("/api/auth/register", async (req, res) => {
    try {
      const authService = new AuthService(null as any);
      const user = await authService.register(req.body);
      res.status(201).json({ success: true, user });
    } catch (error: any) {
      res.status(400).json({ success: false, error: error.message });
    }
  });

  // Login route
  app.post("/api/auth/login", async (req, res) => {
    try {
      const authService = new AuthService(null as any);
      const user = await authService.login(req.body);
      res.status(200).json({ success: true, user });
    } catch (error: any) {
      res.status(401).json({ success: false, error: error.message });
    }
  });

  return app;
};

describe("Auth API Integration Tests", () => {
  let app: Application;
  let mockAuthService: jest.Mocked<AuthService>;

  beforeAll(() => {
    app = createAuthApp();
  });

  beforeEach(() => {
    jest.clearAllMocks();
    mockAuthService = new AuthService(null as any) as jest.Mocked<AuthService>;
  });

  describe("POST /api/auth/register", () => {
    it("should register new user successfully", async () => {
      const mockUser = {
        id: "1",
        username: "testuser",
        email: "test@example.com",
        role: "owner",
      };

      mockAuthService.register = jest.fn().mockResolvedValue(mockUser);
      (AuthService as jest.Mock).mockImplementation(() => mockAuthService);

      const response = await request(app).post("/api/auth/register").send({
        username: "testuser",
        email: "test@example.com",
        password: "SecurePass123!",
      });

      expect(response.status).toBe(201);
      expect(response.body.success).toBe(true);
      expect(response.body.user).toEqual(mockUser);
    });

    it("should return 400 for duplicate username", async () => {
      mockAuthService.register = jest
        .fn()
        .mockRejectedValue(new Error("Username or email already exists"));
      (AuthService as jest.Mock).mockImplementation(() => mockAuthService);

      const response = await request(app).post("/api/auth/register").send({
        username: "duplicate",
        email: "dup@example.com",
        password: "SecurePass123!",
      });

      expect(response.status).toBe(400);
      expect(response.body.success).toBe(false);
      expect(response.body.error).toContain("already exists");
    });

    it("should return 400 for weak password", async () => {
      mockAuthService.register = jest
        .fn()
        .mockRejectedValue(new Error("Password must be at least 8 characters"));
      (AuthService as jest.Mock).mockImplementation(() => mockAuthService);

      const response = await request(app).post("/api/auth/register").send({
        username: "testuser",
        email: "test@example.com",
        password: "weak",
      });

      expect(response.status).toBe(400);
      expect(response.body.success).toBe(false);
      expect(response.body.error).toBeDefined();
    });

    it("should validate required fields", async () => {
      const response = await request(app).post("/api/auth/register").send({});

      expect(response.status).toBe(400);
    });
  });

  describe("POST /api/auth/login", () => {
    it("should login user successfully", async () => {
      const mockUser = {
        id: "1",
        username: "testuser",
        email: "test@example.com",
        role: "owner",
      };

      mockAuthService.login = jest.fn().mockResolvedValue(mockUser);
      (AuthService as jest.Mock).mockImplementation(() => mockAuthService);

      const response = await request(app).post("/api/auth/login").send({
        username: "testuser",
        password: "SecurePass123!",
      });

      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body.user).toEqual(mockUser);
    });

    it("should return 401 for invalid credentials", async () => {
      mockAuthService.login = jest
        .fn()
        .mockRejectedValue(new Error("Invalid username or password"));
      (AuthService as jest.Mock).mockImplementation(() => mockAuthService);

      const response = await request(app).post("/api/auth/login").send({
        username: "testuser",
        password: "WrongPassword123!",
      });

      expect(response.status).toBe(401);
      expect(response.body.success).toBe(false);
      expect(response.body.error).toContain("Invalid");
    });

    it("should return 401 for locked account", async () => {
      mockAuthService.login = jest
        .fn()
        .mockRejectedValue(
          new Error(
            "Account locked. Too many attempts, please try again in 15 minutes"
          )
        );
      (AuthService as jest.Mock).mockImplementation(() => mockAuthService);

      const response = await request(app).post("/api/auth/login").send({
        username: "lockeduser",
        password: "Password123!",
      });

      expect(response.status).toBe(401);
      expect(response.body.error).toContain("Account locked");
    });
  });

  describe("Request Validation", () => {
    it("should reject requests with invalid JSON", async () => {
      const response = await request(app)
        .post("/api/auth/login")
        .send("invalid-json")
        .set("Content-Type", "application/json");

      expect(response.status).toBe(400);
    });

    it("should accept valid JSON requests", async () => {
      mockAuthService.login = jest
        .fn()
        .mockRejectedValue(new Error("Invalid username or password"));
      (AuthService as jest.Mock).mockImplementation(() => mockAuthService);

      const response = await request(app)
        .post("/api/auth/login")
        .send({
          username: "test",
          password: "test",
        })
        .set("Content-Type", "application/json");

      expect(response.headers["content-type"]).toMatch(/application\/json/);
    });
  });
});
