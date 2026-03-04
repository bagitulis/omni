/**
 * Order Data Converter
 * Converts data types for JSON serialization
 * Single Responsibility: Data type conversion
 */

export class OrderDataConverter {
  /**
   * Convert BigInt values to strings for JSON serialization
   * Recursively handles objects, arrays, and primitives
   */
  static convertBigIntToString(obj: any): any {
    if (typeof obj === "bigint") {
      return obj.toString();
    }
    if (Array.isArray(obj)) {
      return obj.map((item) => this.convertBigIntToString(item));
    }
    if (obj !== null && typeof obj === "object") {
      const result: any = {};
      for (const key in obj) {
        result[key] = this.convertBigIntToString(obj[key]);
      }
      return result;
    }
    return obj;
  }

  /**
   * Convert orders array with BigInt fields to JSON-safe format
   */
  static convertOrdersToJSON(orders: any[]): any[] {
    return orders.map((order) => this.convertBigIntToString(order));
  }
}
