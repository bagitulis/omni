export type Platform = "shopee" | "tiktok" | "lazada";

export interface PlatformMeta {
  key: Platform;
  label: string;
  color: string;
}

export const PLATFORM_META: PlatformMeta[] = [
  { key: "shopee", label: "Shopee", color: "volcano" },
  { key: "tiktok", label: "TikTok", color: "default" },
  { key: "lazada", label: "Lazada", color: "geekblue" },
];
