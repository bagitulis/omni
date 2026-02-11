/**
 * Platform icon and label utilities
 * Uses emoji icons since React frontend doesn't have platform icon assets yet
 */

export function getPlatformIcon(platform: string | undefined): string {
  const icons: Record<string, string> = {
    shopee: "🛍️",
    lazada: "📦",
    tiktok: "🎵",
  };
  return icons[(platform || "").toLowerCase()] || "🛒";
}

export function getPlatformIconEmoji(platform: string | undefined): string {
  return getPlatformIcon(platform);
}

export function getPlatformLabel(platform: string | undefined): string {
  if (!platform) return "Unknown";
  return platform.charAt(0).toUpperCase() + platform.slice(1);
}
