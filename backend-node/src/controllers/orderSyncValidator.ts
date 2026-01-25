/**
 * Order Sync Validator
 * Validates request parameters for order sync operations
 * Single Responsibility: Input validation for order sync endpoints
 */

type CategoryType = "unpaid" | "unprocess" | "processed";
type PlatformType = "shopee" | "lazada" | "tiktok";

export class OrderSyncValidator {
  private static readonly VALID_CATEGORIES: CategoryType[] = [
    "unpaid",
    "unprocess",
    "processed",
  ];

  private static readonly VALID_PLATFORMS: PlatformType[] = [
    "shopee",
    "lazada",
    "tiktok",
  ];

  /**
   * Validate category parameter
   */
  static validateCategory(category: string): {
    isValid: boolean;
    error?: string;
  } {
    if (!this.VALID_CATEGORIES.includes(category as CategoryType)) {
      return {
        isValid: false,
        error: "Invalid category. Must be: unpaid, unprocess, or processed",
      };
    }
    return { isValid: true };
  }

  /**
   * Validate platform parameter
   */
  static validatePlatform(platform: string): {
    isValid: boolean;
    error?: string;
  } {
    if (!this.VALID_PLATFORMS.includes(platform as PlatformType)) {
      return {
        isValid: false,
        error: "Invalid platform. Must be: shopee, lazada, or tiktok",
      };
    }
    return { isValid: true };
  }

  /**
   * Validate days parameter
   */
  static validateDays(days: any): {
    isValid: boolean;
    value: number;
    error?: string;
  } {
    const daysNum = Number(days);
    if (isNaN(daysNum) || daysNum < 1 || daysNum > 365) {
      return {
        isValid: false,
        value: 7,
        error: "Invalid days. Must be between 1 and 365",
      };
    }
    return { isValid: true, value: daysNum };
  }

  /**
   * Validate order IDs array
   */
  static validateOrderIds(orderIds: any): {
    isValid: boolean;
    error?: string;
  } {
    if (!Array.isArray(orderIds) || orderIds.length === 0) {
      return {
        isValid: false,
        error: "orderIds must be a non-empty array",
      };
    }
    return { isValid: true };
  }

  /**
   * Extract tenant ID from request
   * No hardcoded default - caller must provide or use empty string
   */
  static extractTenantId(req: any, defaultTenant: string = ""): string {
    return (
      req.tenantId || (req.headers["x-tenant-id"] as string) || defaultTenant
    );
  }

  /**
   * List of special account types without platform credentials
   * ⚠️ NOTE: These are NOT real tenants!
   * - 'system' = refers to GlobalConfig operations (system.db has no orders!)
   * - 'developer', 'admin' = special user roles
   *
   * Real tenants with platform access: yumna_bertigamart, tika_nusseyba
   */
  private static readonly NON_PLATFORM_ACCOUNTS = [
    "system",
    "developer",
    "admin",
  ];

  /**
   * Validate that the tenant has platform access
   * System and developer accounts cannot directly sync orders
   * They must switch to a real tenant (yumna_bertigamart, tika_nusseyba) first
   */
  static validateTenantHasPlatformAccess(tenantId: string): {
    isValid: boolean;
    error?: string;
  } {
    if (this.NON_PLATFORM_ACCOUNTS.includes(tenantId.toLowerCase())) {
      return {
        isValid: false,
        error:
          `'${tenantId}' is not a real tenant with platform credentials. ` +
          `Please switch to a real tenant (yumna_bertigamart or tika_nusseyba) to access platform operations.`,
      };
    }
    return { isValid: true };
  }
}
