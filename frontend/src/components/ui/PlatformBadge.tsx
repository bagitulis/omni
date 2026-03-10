import { Tag } from "antd";

interface PlatformBadgeProps {
  platform: string;
}

const PLATFORM_CONFIG: Record<string, { color: string; label: string }> = {
  shopee: { color: "processing", label: "Shopee" },
  tiktok: { color: "default", label: "TikTok" },
  lazada: { color: "warning", label: "Lazada" },
};

export function PlatformBadge({ platform }: PlatformBadgeProps) {
  const normalizedPlatform = platform.toLowerCase();
  const config = PLATFORM_CONFIG[normalizedPlatform] || {
    color: "default",
    label: platform,
  };

  return (
    <Tag
      color={config.color}
      style={{
        marginRight: 0,
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
