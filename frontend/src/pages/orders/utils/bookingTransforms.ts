/**
 * Booking Transforms
 * Pure utility functions for formatting booking data for display
 */

/**
 * Format booking status string to human-readable form.
 * Examples: "BOOKED" → "Booked", "PICKED_UP" → "Picked Up"
 */
export function formatBookingStatus(status: string): string {
  if (!status) return "";
  return status
    .toLowerCase()
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

/**
 * Format match status string to human-readable form.
 * "MATCHED" → "Matched", "NOT_MATCHED" → "Not Matched"
 */
export function formatMatchStatus(status: string): string {
  if (!status) return "";
  if (status.toUpperCase() === "MATCHED") return "Matched";
  if (status.toUpperCase() === "NOT_MATCHED") return "Not Matched";
  return formatBookingStatus(status);
}

/**
 * Format Unix seconds timestamp to date string.
 * Returns "YYYY-MM-DD HH:mm:ss" or empty string for invalid/zero input.
 */
export function formatBookingTime(unixSeconds: number): string {
  if (!unixSeconds || unixSeconds <= 0) return "";
  const date = new Date(unixSeconds * 1000);
  if (Number.isNaN(date.getTime())) return "";
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  const seconds = String(date.getSeconds()).padStart(2, "0");
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

/**
 * Get Ant Design Tag color for booking status.
 */
export function getBookingStatusColor(
  status: string,
): "blue" | "processing" | "success" | "warning" | "error" | "default" {
  const upper = status.toUpperCase();
  switch (upper) {
    case "BOOKED":
    case "BOOKING":
      return "blue";
    case "PICKED_UP":
      return "processing";
    case "DELIVERED":
      return "success";
    case "CANCELLED":
    case "CANCELED":
      return "error";
    default:
      return "default";
  }
}

/**
 * Get Ant Design Tag color for match status.
 */
export function getMatchStatusColor(
  status: string,
): "success" | "error" | "warning" | "default" {
  const upper = status.toUpperCase();
  switch (upper) {
    case "MATCHED":
      return "success";
    case "NOT_MATCHED":
      return "error";
    default:
      return "default";
  }
}
