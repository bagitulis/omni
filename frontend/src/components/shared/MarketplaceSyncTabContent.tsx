import type { FC } from "react";
import {
  Alert,
  Button,
  Checkbox,
  InputNumber,
  Radio,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import { BulbOutlined } from "@ant-design/icons";
import type { ColumnType } from "antd/es/table";
import type { Platform } from "@/types/shared";
import { PLATFORM_OPTIONS, type PerPlatformRow } from "./stockSyncColumns";
import type { PricePerPlatformRow } from "./priceSyncColumns";

// ─── Types ───────────────────────────────────────────────────────────────────

type SyncMode = "uniform" | "per_platform";

interface ModeToggleProps {
  mode: SyncMode;
  onModeChange: (m: SyncMode) => void;
  recoLoading: boolean;
  onApplyRecommendations: () => void;
  recoCount: number;
  tooltipText: string;
}

interface UniformSectionProps {
  value: number;
  onChange: (v: number) => void;
  platforms: Record<Platform, boolean>;
  onPlatformChange: (p: Platform, checked: boolean) => void;
  label: string;
}

// ─── Shared Sub-Components ───────────────────────────────────────────────────

const ModeToggle: FC<ModeToggleProps> = ({
  mode,
  onModeChange,
  recoLoading,
  onApplyRecommendations,
  recoCount,
  tooltipText,
}) => (
  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
    <Radio.Group value={mode} onChange={(e) => onModeChange(e.target.value as SyncMode)}>
      <Radio.Button value="uniform">Uniform</Radio.Button>
      <Radio.Button value="per_platform">Per Platform</Radio.Button>
    </Radio.Group>
    {mode === "per_platform" && (
      <Tooltip title={tooltipText}>
        <Button
          icon={<BulbOutlined />}
          size="small"
          onClick={onApplyRecommendations}
          disabled={recoLoading || recoCount === 0}
          loading={recoLoading}
        >
          Fill from Inventory
        </Button>
      </Tooltip>
    )}
  </div>
);

const UniformSection: FC<UniformSectionProps> = ({
  value,
  onChange,
  platforms,
  onPlatformChange,
  label,
}) => (
  <div>
    <div style={{ marginBottom: 16 }}>
      <Typography.Text strong style={{ display: "block", marginBottom: 4 }}>
        {label}
      </Typography.Text>
      <InputNumber
        min={0}
        value={value}
        onChange={(v) => onChange(v ?? 0)}
        style={{ width: 200 }}
        formatter={(v) => v != null ? `${Number(v).toLocaleString("id-ID")}` : ""}
        parser={(v) => Number((v ?? "").replace(/\./g, "")) || 0}
      />
    </div>
    <div>
      <Typography.Text strong style={{ display: "block", marginBottom: 4 }}>
        Target Platforms
      </Typography.Text>
      <Space>
        {PLATFORM_OPTIONS.map((opt) => (
          <Checkbox
            key={opt.key}
            checked={platforms[opt.key]}
            onChange={(e) => onPlatformChange(opt.key, e.target.checked)}
          >
            {opt.label}
          </Checkbox>
        ))}
      </Space>
    </div>
  </div>
);

const PerPlatformTable: FC<{
  columns: ColumnType<PerPlatformRow | PricePerPlatformRow>[];
  data: (PerPlatformRow | PricePerPlatformRow)[];
}> = ({ columns, data }) => (
  <Table
    columns={columns}
    dataSource={data}
    size="small"
    pagination={false}
    scroll={{ y: 340 }}
    bordered
  />
);

// ─── Stock Tab ───────────────────────────────────────────────────────────────

export interface StockTabContentProps {
  lockedSkuCount: number;
  totalLockedQty: number;
  stockMode: SyncMode;
  setStockMode: (m: SyncMode) => void;
  stockRecoLoading: boolean;
  onApplyStockRecommendations: () => void;
  stockRecoCount: number;
  uniformStock: number;
  setUniformStock: (v: number) => void;
  stockUniformPlatforms: Record<Platform, boolean>;
  setStockUniformPlatforms: (p: Platform, checked: boolean) => void;
  stockColumns: ColumnType<PerPlatformRow | PricePerPlatformRow>[];
  stockRows: PerPlatformRow[];
}

export const StockTabContent: FC<StockTabContentProps> = ({
  lockedSkuCount,
  totalLockedQty,
  stockMode,
  setStockMode,
  stockRecoLoading,
  onApplyStockRecommendations,
  stockRecoCount,
  uniformStock,
  setUniformStock,
  stockUniformPlatforms,
  setStockUniformPlatforms,
  stockColumns,
  stockRows,
}) => (
  <div>
    {lockedSkuCount > 0 && (
      <Alert
        type="warning"
        showIcon
        style={{ marginBottom: 12 }}
        message={
          <span>
            🔒 <strong>Lock Stock Active</strong> —{" "}
            <Tag color="error" style={{ borderRadius: 3 }}>{lockedSkuCount} SKUs</Tag>
            <Tag color="warning" style={{ borderRadius: 3 }}>{totalLockedQty} pcs locked</Tag>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              Recommendations auto-adjusted
            </Typography.Text>
          </span>
        }
      />
    )}
    <ModeToggle
      mode={stockMode}
      onModeChange={setStockMode}
      recoLoading={stockRecoLoading}
      onApplyRecommendations={onApplyStockRecommendations}
      recoCount={stockRecoCount}
      tooltipText="Auto-fills from inventory. Locked orders deducted, allocated by ratio settings."
    />
    {stockMode === "uniform"
      ? (
        <UniformSection
          value={uniformStock}
          onChange={setUniformStock}
          platforms={stockUniformPlatforms}
          onPlatformChange={setStockUniformPlatforms}
          label="Stock Quantity"
        />
      )
      : <PerPlatformTable columns={stockColumns} data={stockRows} />
    }
  </div>
);

// ─── Price Tab ───────────────────────────────────────────────────────────────

export interface PriceTabContentProps {
  priceMode: SyncMode;
  setPriceMode: (m: SyncMode) => void;
  priceRecoLoading: boolean;
  onApplyPriceRecommendations: () => void;
  priceRecoCount: number;
  uniformPrice: number;
  setUniformPrice: (v: number) => void;
  priceUniformPlatforms: Record<Platform, boolean>;
  setPriceUniformPlatforms: (p: Platform, checked: boolean) => void;
  priceColumns: ColumnType<PerPlatformRow | PricePerPlatformRow>[];
  priceRows: PricePerPlatformRow[];
}

export const PriceTabContent: FC<PriceTabContentProps> = ({
  priceMode,
  setPriceMode,
  priceRecoLoading,
  onApplyPriceRecommendations,
  priceRecoCount,
  uniformPrice,
  setUniformPrice,
  priceUniformPlatforms,
  setPriceUniformPlatforms,
  priceColumns,
  priceRows,
}) => (
  <div>
    <ModeToggle
      mode={priceMode}
      onModeChange={setPriceMode}
      recoLoading={priceRecoLoading}
      onApplyRecommendations={onApplyPriceRecommendations}
      recoCount={priceRecoCount}
      tooltipText="Auto-fills per-platform prices from inventory sheet (HARGA_SHOPEE, HARGA_TIKTOK, HARGA_LAZADA)."
    />
    {priceMode === "uniform"
      ? (
        <UniformSection
          value={uniformPrice}
          onChange={setUniformPrice}
          platforms={priceUniformPlatforms}
          onPlatformChange={setPriceUniformPlatforms}
          label="Price"
        />
      )
      : <PerPlatformTable columns={priceColumns} data={priceRows} />
    }
  </div>
);
