/**
 * Lazada Product Validator
 * Validates pagination and product-related parameters
 * Single Responsibility: Input validation for Lazada products
 */

export class LazadaProductValidator {
  /**
   * Validate pagination parameters
   */
  static validatePagination(
    offset: number,
    limit: number
  ): { offset: number; limit: number } {
    const validOffset = Math.max(0, offset);
    const validLimit = Math.min(Math.max(1, limit), 1000);
    return { offset: validOffset, limit: validLimit };
  }

  /**
   * Validate filter parameter
   */
  static validateFilter(filter: string): string {
    const validFilters = ["all", "live", "inactive", "deleted", "image-missing", "pending", "rejected", "sold-out"];
    return validFilters.includes(filter) ? filter : "live";
  }
}
