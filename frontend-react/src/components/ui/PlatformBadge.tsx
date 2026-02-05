import { Tag } from "antd";

interface PlatformBadgeProps {
  platform: string;
}

const PLATFORM_CONFIG: Record<string, { color: string; label: string }> = {
  shopee: { color: "#ee4d2d", label: "Shopee" },
  tiktok: { color: "#000000", label: "TikTok" },
  lazada: { color: "#0f146d", label: "Lazada" },
  tokopedia: { color: "#42b549", label: "Tokopedia" },
};

export function PlatformBadge({ platform }: PlatformBadgeProps) {
  const normalizedPlatform = platform.toLowerCase();
  const config = PLATFORM_CONFIG[normalizedPlatform] || {
    color: "#888",
    label: platform,
  };

  return (
    <Tag
      color={config.color}
      style={{
        marginRight: 0,
        color: "white",
        fontWeight: 500,
        border: "none",
        fontSize: "10px",
        borderRadius: "3px", // Sharp Ginee style
      }}
    >
      {config.label}
    </Tag>
  );
}
