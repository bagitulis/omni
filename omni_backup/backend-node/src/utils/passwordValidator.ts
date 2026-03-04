/**
 * Password Validator
 * Enforces strong password policy
 */
export interface PasswordValidationResult {
  valid: boolean;
  errors: string[];
}

export class PasswordValidator {
  private minLength = 8;
  private requireUppercase = true;
  private requireLowercase = true;
  private requireNumbers = true;
  private requireSpecialChars = false; // Optional

  /**
   * Validate password against policy
   */
  validate(password: string): PasswordValidationResult {
    const errors: string[] = [];

    if (!password || password.length < this.minLength) {
      errors.push(`Password must be at least ${this.minLength} characters`);
    }

    if (this.requireUppercase && !/[A-Z]/.test(password)) {
      errors.push("Password must contain at least one uppercase letter");
    }

    if (this.requireLowercase && !/[a-z]/.test(password)) {
      errors.push("Password must contain at least one lowercase letter");
    }

    if (this.requireNumbers && !/\d/.test(password)) {
      errors.push("Password must contain at least one number");
    }

    if (this.requireSpecialChars && !/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      errors.push("Password must contain at least one special character");
    }

    // Check against common passwords
    const commonPasswords = [
      "password",
      "12345678",
      "qwerty123",
      "admin123",
      "password1",
      "password123",
    ];

    if (commonPasswords.includes(password.toLowerCase())) {
      errors.push("Password is too common, please choose a stronger password");
    }

    return {
      valid: errors.length === 0,
      errors,
    };
  }
}
