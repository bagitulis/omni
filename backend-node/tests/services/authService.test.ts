/**
 * AuthService Unit Tests
 * Tests authentication, registration, password hashing, and account locking
 */

import {
  AuthService,
  UserRegistration,
  UserLogin,
} from "../../src/services/authService";
import { PrismaClient } from "@prisma/client";

// Mock PrismaClient
jest.mock("@prisma/client", () => {
  const mockPrisma = {
    user: {
      findFirst: jest.fn(),
      findUnique: jest.fn(),
      create: jest.fn(),
      update: jest.fn(),
    },
  };
  return {
    PrismaClient: jest.fn(() => mockPrisma),
  };
});

describe("AuthService", () => {
  let authService: AuthService;
  let mockPrisma: any;

  beforeEach(() => {
    mockPrisma = new PrismaClient();
    authService = new AuthService(mockPrisma);
    jest.clearAllMocks();
  });

  describe("hashPassword", () => {
    it("should hash password successfully", async () => {
      const password = "TestPassword123!";
      const hash = await authService.hashPassword(password);

      expect(hash).toBeDefined();
      expect(hash).not.toBe(password);
      expect(hash.length).toBeGreaterThan(20);
    });

    it("should create different hashes for same password", async () => {
      const password = "TestPassword123!";
      const hash1 = await authService.hashPassword(password);
      const hash2 = await authService.hashPassword(password);

      expect(hash1).not.toBe(hash2);
    });
  });

  describe("comparePassword", () => {
    it("should return true for correct password", async () => {
      const password = "TestPassword123!";
      const hash = await authService.hashPassword(password);
      const isMatch = await authService.comparePassword(password, hash);

      expect(isMatch).toBe(true);
    });

    it("should return false for incorrect password", async () => {
      const password = "TestPassword123!";
      const wrongPassword = "WrongPassword456!";
      const hash = await authService.hashPassword(password);
      const isMatch = await authService.comparePassword(wrongPassword, hash);

      expect(isMatch).toBe(false);
    });
  });

  describe("register", () => {
    const validRegistration: UserRegistration = {
      username: "testuser",
      email: "test@example.com",
      password: "SecurePass123!",
    };

    it("should register new user successfully", async () => {
      mockPrisma.user.findFirst.mockResolvedValue(null);
      mockPrisma.user.create.mockResolvedValue({
        id: "1",
        username: validRegistration.username,
        email: validRegistration.email,
        password: "hashedPassword",
        role: "owner",
        failedLoginAttempts: 0,
        accountLockedUntil: null,
        lastFailedLogin: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      });

      const result = await authService.register(validRegistration);

      expect(result).toEqual({
        id: "1",
        username: validRegistration.username,
        email: validRegistration.email,
        role: "owner",
      });
      expect(mockPrisma.user.findFirst).toHaveBeenCalledWith({
        where: {
          OR: [
            { username: validRegistration.username },
            { email: validRegistration.email },
          ],
        },
      });
      expect(mockPrisma.user.create).toHaveBeenCalled();
    });

    it("should throw error if username already exists", async () => {
      mockPrisma.user.findFirst.mockResolvedValue({
        id: "1",
        username: validRegistration.username,
      });

      await expect(authService.register(validRegistration)).rejects.toThrow(
        "Username or email already exists"
      );
    });

    it("should throw error if email already exists", async () => {
      mockPrisma.user.findFirst.mockResolvedValue({
        id: "1",
        email: validRegistration.email,
      });

      await expect(authService.register(validRegistration)).rejects.toThrow(
        "Username or email already exists"
      );
    });

    it("should throw error for weak password", async () => {
      mockPrisma.user.findFirst.mockResolvedValue(null);

      const weakPasswordData = {
        ...validRegistration,
        password: "weak",
      };

      await expect(authService.register(weakPasswordData)).rejects.toThrow();
    });

    it("should throw error for password without uppercase", async () => {
      mockPrisma.user.findFirst.mockResolvedValue(null);

      const noUppercaseData = {
        ...validRegistration,
        password: "lowercase123!",
      };

      await expect(authService.register(noUppercaseData)).rejects.toThrow();
    });

    it("should throw error for password without numbers", async () => {
      mockPrisma.user.findFirst.mockResolvedValue(null);

      const noNumberData = {
        ...validRegistration,
        password: "NoNumbers!",
      };

      await expect(authService.register(noNumberData)).rejects.toThrow();
    });
  });

  describe("login", () => {
    const validLogin: UserLogin = {
      username: "testuser",
      password: "TestPassword123!",
    };

    it("should login user successfully with correct credentials", async () => {
      const hashedPassword = await authService.hashPassword(
        validLogin.password
      );

      mockPrisma.user.findUnique.mockResolvedValue({
        id: "1",
        username: validLogin.username,
        email: "test@example.com",
        password: hashedPassword,
        role: "owner",
        failedLoginAttempts: 0,
        accountLockedUntil: null,
        lastFailedLogin: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      });

      mockPrisma.user.update.mockResolvedValue({});

      const result = await authService.login(validLogin);

      expect(result).toEqual({
        id: "1",
        username: validLogin.username,
        email: "test@example.com",
        role: "owner",
      });
      expect(mockPrisma.user.findUnique).toHaveBeenCalledWith({
        where: { username: validLogin.username },
      });
    });

    it("should throw error for non-existent user", async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await expect(authService.login(validLogin)).rejects.toThrow(
        "Invalid username or password"
      );
    });

    it("should throw error for incorrect password", async () => {
      const correctPassword = "CorrectPassword123!";
      const hashedPassword = await authService.hashPassword(correctPassword);

      mockPrisma.user.findUnique.mockResolvedValue({
        id: "1",
        username: validLogin.username,
        password: hashedPassword,
        failedLoginAttempts: 0,
        accountLockedUntil: null,
      });

      const wrongLogin = {
        username: validLogin.username,
        password: "WrongPassword456!",
      };

      await expect(authService.login(wrongLogin)).rejects.toThrow(
        "Invalid username or password"
      );
    });

    it("should throw error if account is locked", async () => {
      const futureDate = new Date(Date.now() + 15 * 60000); // 15 minutes from now

      mockPrisma.user.findUnique.mockResolvedValue({
        id: "1",
        username: validLogin.username,
        accountLockedUntil: futureDate,
      });

      await expect(authService.login(validLogin)).rejects.toThrow(
        /Account locked/
      );
    });

    it("should increment failed login attempts on wrong password", async () => {
      const correctPassword = "CorrectPassword123!";
      const hashedPassword = await authService.hashPassword(correctPassword);

      mockPrisma.user.findUnique.mockResolvedValue({
        id: "1",
        username: validLogin.username,
        password: hashedPassword,
        failedLoginAttempts: 2,
        accountLockedUntil: null,
      });

      mockPrisma.user.update.mockResolvedValue({});

      const wrongLogin = {
        username: validLogin.username,
        password: "WrongPassword456!",
      };

      await expect(authService.login(wrongLogin)).rejects.toThrow();

      expect(mockPrisma.user.update).toHaveBeenCalledWith({
        where: { id: "1" },
        data: {
          failedLoginAttempts: 3,
          lastFailedLogin: expect.any(Date),
        },
      });
    });

    it("should reset failed login attempts on successful login", async () => {
      const hashedPassword = await authService.hashPassword(
        validLogin.password
      );

      mockPrisma.user.findUnique.mockResolvedValue({
        id: "1",
        username: validLogin.username,
        email: "test@example.com",
        password: hashedPassword,
        role: "owner",
        failedLoginAttempts: 3,
        accountLockedUntil: null,
      });

      mockPrisma.user.update.mockResolvedValue({});

      await authService.login(validLogin);

      expect(mockPrisma.user.update).toHaveBeenCalledWith({
        where: { id: "1" },
        data: {
          failedLoginAttempts: 0,
          lastFailedLogin: null,
          accountLockedUntil: null,
        },
      });
    });
  });

  describe("formatUserResponse", () => {
    it("should format user response correctly", () => {
      const user = {
        id: "1",
        username: "testuser",
        email: "test@example.com",
        password: "hashedPassword",
        role: "owner",
        failedLoginAttempts: 0,
        accountLockedUntil: null,
        lastFailedLogin: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      // Access private method through any
      const result = (authService as any).formatUserResponse(user);

      expect(result).toEqual({
        id: "1",
        username: "testuser",
        email: "test@example.com",
        role: "owner",
      });
      expect(result.password).toBeUndefined();
    });
  });
});
