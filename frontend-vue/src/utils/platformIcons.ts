// Platform icon utilities
import ShopeeIcon from "@/assets/icons/shopee.webp";
import LazadaIcon from "@/assets/icons/lazada.webp";
import TiktokIcon from "@/assets/icons/tiktok.webp";

export const getPlatformIconSvg = (platform: string | undefined): string => {
  const icons: Record<string, string> = {
    shopee: ShopeeIcon,
    lazada: LazadaIcon,
    tiktok: TiktokIcon,
  };
  return icons[(platform || "").toLowerCase()] || ShopeeIcon;
};

export const getPlatformIconEmoji = (platform: string | undefined): string => {
  const icons: Record<string, string> = {
    shopee: "🛍️",
    lazada: "📦",
    tiktok: "🎵",
  };
  return icons[(platform || "").toLowerCase()] || "🛒";
};

export const getPlatformLabel = (platform: string | undefined): string => {
  if (!platform) return "Unknown";
  return platform.charAt(0).toUpperCase() + platform.slice(1);
};
