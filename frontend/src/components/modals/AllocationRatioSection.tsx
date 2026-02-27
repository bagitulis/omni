import { useMemo } from "react";
import { Card, Slider, Typography, theme } from "antd";
import {
  calculateMarketplaceAllocation,
  type MarketplaceAllocationSettings,
} from "@/pages/inventory/utils/marketplaceAllocation";

const { Text, Title } = Typography;

interface AllocationRatioSectionProps {
  settings: MarketplaceAllocationSettings;
  onChangeShopeeRatio: (value: number) => void;
  onChangeTiktokRatio: (value: number) => void;
}

/** Allocation ratio sliders, live preview card, and formula reference. */
export function AllocationRatioSection({
  settings,
  onChangeShopeeRatio,
  onChangeTiktokRatio,
}: AllocationRatioSectionProps) {
  const {
    token: {
      colorInfo,
      colorText,
      colorTextSecondary,
      colorBgLayout,
      colorSuccess,
    },
  } = theme.useToken();

  const total = 20;
  const preview = useMemo(
    () => calculateMarketplaceAllocation(total, false, settings),
    [settings],
  );

  return (
    <>
      <div>
        <Title level={5}>Allocation Ratios</Title>
        <Text type="secondary">
          Ratios are used for preview and can be fine-tuned later when backend
          fields are available.
        </Text>

        <div style={{ marginTop: 16 }}>
          <Text>
            Shopee Ratio: {(settings.shopeeRatio * 100).toFixed(0)}%
          </Text>
          <Slider
            min={0}
            max={1}
            step={0.05}
            value={settings.shopeeRatio}
            onChange={onChangeShopeeRatio}
          />
        </div>

        <div style={{ marginTop: 16 }}>
          <Text>
            TikTok Ratio: {(settings.tiktokRatio * 100).toFixed(0)}%
          </Text>
          <Slider
            min={0}
            max={1}
            step={0.05}
            value={settings.tiktokRatio}
            onChange={onChangeTiktokRatio}
          />
        </div>

        <Card
          size="small"
          style={{
            marginTop: 16,
            background: colorBgLayout,
            borderRadius: 3,
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              fontWeight: 500,
            }}
          >
            <span style={{ color: colorInfo }}>Shopee: {preview.shopee}</span>
            <span style={{ color: colorText }}>TikTok: {preview.tiktok}</span>
            <span style={{ color: colorSuccess }}>
              Lazada: {preview.lazada}
            </span>
            <span style={{ color: colorTextSecondary }}>Total: {total}</span>
          </div>
        </Card>
      </div>

      <div>
        <Title level={5}>Formula Reference</Title>
        <ul>
          <li>
            <strong>Shopee:</strong> MIN(CEILING(ratio * Total), Total)
          </li>
          <li>
            <strong>TikTok:</strong> MIN(CEILING(ratio * Total), Total -
            Shopee)
          </li>
          <li>
            <strong>Lazada:</strong> Total - Shopee - TikTok
          </li>
          <li>
            <strong>Auto = TRUE:</strong> All platforms = Total
          </li>
        </ul>
      </div>
    </>
  );
}
