/**
 * CheckboxDataConverter.ts
 * Frontend utility untuk handle checkbox display logic
 * 
 * Frontend display → checkbox input element
 * Database storage → "true"/"false" string atau boolean
 */

export class CheckboxDataConverter {
  /**
   * Convert value ke boolean untuk checkbox checked state
   */
  static toBoolean(value: any): boolean {
    if (typeof value === "boolean") return value;
    if (value === null || value === undefined) return false;
    const str = String(value).toLowerCase();
    return ["true", "1", "yes", "✓"].includes(str);
  }

  /**
   * Check if value is boolean-like
   * Only "true"/"false" strings and direct boolean type, NOT "0"/"1"
   */
  static isBooleanLike(value: any): boolean {
    // Direct boolean type check
    if (typeof value === "boolean") return true;
    
    if (value === null || value === undefined) return false;
    
    // Only check for "true" and "false" strings, NOT "0" and "1"
    const str = String(value).toLowerCase();
    return str === "true" || str === "false";
  }
}
