import { Alert, Button, Checkbox, InputNumber, Space, Table, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { Platform } from "@/types/shared";
import { PLATFORM_OPTIONS, type PerPlatformRow } from "./stockSyncColumns";

interface UniformStockSectionProps {
  uniformStock: number;
  onUniformStockChange: (value: number) => void;
  uniformPlatforms: Record<Platform, boolean>;
  onUniformPlatformChange: (platform: Platform, checked: boolean) => void;
  skippedSkuCount: number;
  totalSelectedSkus: number;
  validSkuCount: number;
  operationCount: number;
  secondaryTextColor: string;
}

export function UniformStockSection({
  uniformStock,
  onUniformStockChange,
  uniformPlatforms,
  onUniformPlatformChange,
  skippedSkuCount,
  totalSelectedSkus,
  validSkuCount,
  operationCount,
  secondaryTextColor,
}: UniformStockSectionProps) {
  return (
    <div>
      <Alert
        message="All selected SKUs will be synced with the same stock value to checked platforms."
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <div style={{ marginBottom: 16 }}>
        <div style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>
          Stock Quantity
        </div>
        <InputNumber
          min={0}
          value={uniformStock}
          onChange={(value) => onUniformStockChange(value ?? 0)}
          style={{ width: 200 }}
        />
      </div>

      <div>
        <div style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>
          Target Platforms
        </div>
        <Space>
          {PLATFORM_OPTIONS.map((platformOption) => (
            <Checkbox
              key={platformOption.key}
              checked={uniformPlatforms[platformOption.key]}
              onChange={(event) =>
                onUniformPlatformChange(
                  platformOption.key,
                  event.target.checked,
                )
              }
            >
              {platformOption.label}
            </Checkbox>
          ))}
        </Space>
      </div>

      <div style={{ marginTop: 16, color: secondaryTextColor }}>
        {skippedSkuCount > 0 ? (
          <Alert
            message={`${skippedSkuCount} SKU(s) skipped because they are not linked to selected platforms.`}
            type="warning"
            showIcon
            style={{ marginBottom: 12 }}
          />
        ) : null}
        <Tag>{totalSelectedSkus} selected SKUs</Tag>
        <Tag>{validSkuCount} ready SKUs</Tag>
        <Tag color="processing">{operationCount} sync operations</Tag>
      </div>
    </div>
  );
}

interface PerPlatformStockSectionProps {
  recommendationsLoading: boolean;
  onApplyRecommendations: () => void;
  perPlatformColumns: ColumnsType<PerPlatformRow>;
  perPlatformData: PerPlatformRow[];
}

export function PerPlatformStockSection({
  recommendationsLoading,
  onApplyRecommendations,
  perPlatformColumns,
  perPlatformData,
}: PerPlatformStockSectionProps) {
  return (
    <div>
      <Alert
        message="Set stock and target platforms individually per SKU."
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <Alert
        message={
          recommendationsLoading
            ? "Loading inventory recommendations & locked orders..."
            : "Inventory recommendation available (locked orders auto-deducted). Apply to all rows."
        }
        type="warning"
        showIcon
        style={{ marginBottom: 12 }}
        action={
          <Button
            size="small"
            onClick={onApplyRecommendations}
            disabled={recommendationsLoading}
          >
            Apply Recommendation
          </Button>
        }
      />
      <Table
        columns={perPlatformColumns}
        dataSource={perPlatformData}
        size="small"
        pagination={false}
        scroll={{ y: 300 }}
        bordered
      />
    </div>
  );
}
