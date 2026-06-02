/** Platform region constants for credential management */

export const SHOPEE_REGIONS = [
  { value: "sg", label: "Singapore" },
  { value: "my", label: "Malaysia" },
  { value: "ph", label: "Philippines" },
  { value: "th", label: "Thailand" },
  { value: "vn", label: "Vietnam" },
  { value: "id", label: "Indonesia" },
  { value: "tw", label: "Taiwan" },
  { value: "br", label: "Brazil" },
  { value: "mx", label: "Mexico" },
  { value: "co", label: "Colombia" },
  { value: "cl", label: "Chile" },
] as const;

export const LAZADA_REGIONS = [
  { value: "id", label: "Indonesia" },
  { value: "my", label: "Malaysia" },
  { value: "ph", label: "Philippines" },
  { value: "sg", label: "Singapore" },
  { value: "th", label: "Thailand" },
  { value: "vn", label: "Vietnam" },
] as const;

/** Platform display names */
export const PLATFORM_DISPLAY_NAMES: Record<string, string> = {
  shopee: "Shopee",
  lazada: "Lazada",
  tiktok: "TikTok Shop",
};

/** Platform-specific credential field labels */
export const PLATFORM_FIELD_LABELS: Record<string, Record<string, string>> = {
  shopee: {
    partner_id: "Partner ID",
    partner_key: "Partner Key",
    region: "Region",
  },
  lazada: {
    app_key: "App Key (Client ID)",
    app_secret: "App Secret",
    region: "Region",
  },
  tiktok: {
    app_key: "App Key",
    app_secret: "App Secret",
  },
};
