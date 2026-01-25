import bcrypt from "bcryptjs";
import { PrismaClient } from "@prisma/client";
import { PasswordValidator } from "../utils/passwordValidator";

export interface UserRegistration {
  username: string;
  email: string;
  password: string;
}

export interface UserLogin {
  username: string;
  password: string;
}

export interface LoginResult {
  user: UserResponse;
  requiresCaptcha: boolean;
  isLocked: boolean;
  lockMinutesRemaining?: number;
}

export interface UserResponse {
  id: string;
  username: string;
  email: string;
  role: string;
}

export class AuthService {
  private passwordValidator = new PasswordValidator();

  // Login attempt thresholds
  private static readonly CAPTCHA_THRESHOLD = 3; // Show CAPTCHA after 3 failed attempts
  private static readonly LOCKOUT_THRESHOLD = 10; // Lock account after 10 failed attempts
  private static readonly LOCKOUT_DURATION_MS = 30 * 60 * 1000; // 30 minutes lockout

  constructor(private prisma: PrismaClient) {}

  /**
   * Hash password using bcrypt
   */
  async hashPassword(password: string): Promise<string> {
    const saltRounds = 10;
    return bcrypt.hash(password, saltRounds);
  }

  /**
   * Compare password with hash
   */
  async comparePassword(password: string, hash: string): Promise<boolean> {
    return bcrypt.compare(password, hash);
  }

  /**
   * Register new user
   */
  async register(data: UserRegistration): Promise<UserResponse> {
    // Check if user already exists
    const existingUser = await this.prisma.user.findFirst({
      where: {
        OR: [{ username: data.username }, { email: data.email }],
      },
    });

    if (existingUser) {
      throw new Error("Username or email already exists");
    }

    // Validate password with strong policy
    const validation = this.passwordValidator.validate(data.password);
    if (!validation.valid) {
      throw new Error(validation.errors[0]);
    }

    // Hash password
    const hashedPassword = await this.hashPassword(data.password);

    // Create user
    const user = await this.prisma.user.create({
      data: {
        username: data.username,
        email: data.email,
        password: hashedPassword,
        role: "owner",
        failedLoginAttempts: 0,
      },
    });

    return this.formatUserResponse(user);
  }

  /**
   * Login user and return user info
   *
   * Security flow:
   * - 1-3 failed attempts: Normal login
   * - 4-10 failed attempts: Require reCAPTCHA
   * - 10+ failed attempts: Lock account for 30 minutes
   */
  async login(data: UserLogin): Promise<LoginResult> {
    const user = await this.prisma.user.findUnique({
      where: { username: data.username },
    });

    if (!user) {
      throw new Error("Invalid username or password");
    }

    const failedAttempts = user.failedLoginAttempts || 0;

    // Check if account is locked
    if (user.accountLockedUntil) {
      const lockedUntil = new Date(user.accountLockedUntil);
      if (lockedUntil > new Date()) {
        const minutesLeft = Math.ceil(
          (lockedUntil.getTime() - Date.now()) / 60000
        );
        throw new Error(
          `Account locked due to too many failed attempts. Try again in ${minutesLeft} minute${
            minutesLeft !== 1 ? "s" : ""
          }.`
        );
      } else {
        // Lock expired, reset attempts
        await this.prisma.user.update({
          where: { id: user.id },
          data: {
            failedLoginAttempts: 0,
            accountLockedUntil: null,
            lastFailedLogin: null,
          },
        });
      }
    }

    // Verify password
    const isPasswordValid = await this.comparePassword(
      data.password,
      user.password
    );

    if (!isPasswordValid) {
      // Increment failed login attempts
      const newAttempts = failedAttempts + 1;
      const updateData: any = {
        failedLoginAttempts: newAttempts,
        lastFailedLogin: new Date(),
      };

      // Lock account after reaching threshold
      if (newAttempts >= AuthService.LOCKOUT_THRESHOLD) {
        updateData.accountLockedUntil = new Date(
          Date.now() + AuthService.LOCKOUT_DURATION_MS
        );
      }

      await this.prisma.user.update({
        where: { id: user.id },
        data: updateData,
      });

      // Throw error with info about CAPTCHA/lock status
      if (newAttempts >= AuthService.LOCKOUT_THRESHOLD) {
        throw new Error(
          "Account locked due to too many failed attempts. Try again in 30 minutes."
        );
      }

      throw new Error("Invalid username or password");
    }

    // Reset failed login attempts on successful login
    if (failedAttempts > 0) {
      await this.prisma.user.update({
        where: { id: user.id },
        data: {
          failedLoginAttempts: 0,
          accountLockedUntil: null,
          lastFailedLogin: null,
        },
      });
    }

    return {
      user: this.formatUserResponse(user),
      requiresCaptcha: false,
      isLocked: false,
    };
  }

  /**
   * Check login status for a user (CAPTCHA required, locked, etc.)
   */
  async getLoginStatus(username: string): Promise<{
    requiresCaptcha: boolean;
    isLocked: boolean;
    lockMinutesRemaining?: number;
  }> {
    const user = await this.prisma.user.findUnique({
      where: { username },
      select: {
        failedLoginAttempts: true,
        accountLockedUntil: true,
      },
    });

    if (!user) {
      // Don't reveal if user exists - return default
      return { requiresCaptcha: false, isLocked: false };
    }

    const failedAttempts = user.failedLoginAttempts || 0;

    // Check if locked
    if (user.accountLockedUntil) {
      const lockedUntil = new Date(user.accountLockedUntil);
      if (lockedUntil > new Date()) {
        const minutesLeft = Math.ceil(
          (lockedUntil.getTime() - Date.now()) / 60000
        );
        return {
          requiresCaptcha: true,
          isLocked: true,
          lockMinutesRemaining: minutesLeft,
        };
      }
    }

    // Check if CAPTCHA required (after 3 failed attempts)
    const requiresCaptcha = failedAttempts >= AuthService.CAPTCHA_THRESHOLD;

    return { requiresCaptcha, isLocked: false };
  }

  /**
   * Get user by ID
   */
  async getUserById(userId: string): Promise<UserResponse | null> {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
    });

    return user ? this.formatUserResponse(user) : null;
  }

  /**
   * Get user by username
   */
  async getUserByUsername(username: string): Promise<UserResponse | null> {
    const user = await this.prisma.user.findUnique({
      where: { username },
    });

    return user ? this.formatUserResponse(user) : null;
  }

  /**
   * Format user response (exclude password)
   */
  private formatUserResponse(user: any): UserResponse {
    return {
      id: user.id,
      username: user.username,
      email: user.email,
      role: user.role,
    };
  }
}
