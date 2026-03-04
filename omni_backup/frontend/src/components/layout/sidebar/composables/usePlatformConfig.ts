interface Platform {
  value: string;
  label: string;
  icon: string;
}

export function usePlatformConfig() {
  const platforms: Platform[] = [
    { value: "lazada", label: "Lazada", icon: "📦" },
    { value: "shopee", label: "Shopee", icon: "🛍️" },
    { value: "tiktok", label: "TikTok", icon: "🎵" },
  ];

  const productPlatforms: Platform[] = [
    { value: "lazada", label: "Lazada", icon: "📦" },
    { value: "shopee", label: "Shopee", icon: "🛍️" },
    { value: "tiktok", label: "TikTok", icon: "🎵" },
  ];

  return {
    platforms,
    productPlatforms,
  };
}
