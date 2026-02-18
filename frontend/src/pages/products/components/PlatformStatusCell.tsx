import { Space } from "antd";
import { PlatformIndicator } from "@/components/shared/PlatformIndicator";
import { usePlatformStatus } from "@/hooks/usePlatformStatus";
import type { Platform, UnifiedProductRow } from "@/types/shared";

const PLATFORMS: Platform[] = ["shopee", "tiktok", "lazada"];

interface PlatformStatusCellProps {
  product: UnifiedProductRow;
}

export function PlatformStatusCell({ product }: PlatformStatusCellProps) {
  const statusByPlatform = usePlatformStatus(product);

  return (
    <Space size={4}>
      {PLATFORMS.map((platform) => (
        <span key={`${product.id}-${platform}`} className="platform-indicator">
          <PlatformIndicator data={statusByPlatform[platform]} size="small" />
        </span>
      ))}
    </Space>
  );
}
