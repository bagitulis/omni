import type { FC } from "react";
import { Card, Tag, Space, Alert, Descriptions, Image, theme } from "antd";
import { ArrowRightOutlined, WarningOutlined } from "@ant-design/icons";
import type { Platform } from "@/types/shared";
import { type CloneDifference, DiffSummarySection } from "./CloneDiffSummary";

interface ClonePreviewDiffProps {
  source: {
    title: string;
    description: string;
    price: number;
    stock: number;
    images: string[];
    sku: string;
  };
  target: {
    platform: Platform;
    title: string; // May be truncated by platform limits
    description: string;
    price: number;
    stock: number;
    images: string[];
  };
  differences: CloneDifference[];
  warnings: string[];
}

// Platform-specific styling (from Vue ClonePreviewPanel.vue)
const PLATFORM_STYLES: Record<Platform, { gradient: string; label: string }> = {
  shopee: {
    gradient: "linear-gradient(135deg, #ff6b2c 0%, #ff5511 100%)",
    label: "Shopee",
  },
  tiktok: {
    gradient: "linear-gradient(135deg, #000000 0%, #333333 100%)",
    label: "TikTok",
  },
  lazada: {
    gradient: "linear-gradient(135deg, #0f146d 0%, #1a237e 100%)",
    label: "Lazada",
  },
};

// Difference highlight color (from Vue — yellow background)
const DIFF_HIGHLIGHT = "#fef3c7"; // Amber-50

const formatPrice = (val: number): string =>
  `Rp ${Math.round(val).toLocaleString("id-ID")}`;

export const ClonePreviewDiff: FC<ClonePreviewDiffProps> = ({
  source,
  target,
  differences,
  warnings,
}) => {
  const { token } = theme.useToken();
  const platformStyle = PLATFORM_STYLES[target.platform];

  // Check if a field has differences
  const isDifferent = (field: string): boolean =>
    differences.some((d) => d.field === field);

  const getDiffReason = (field: string): string | undefined =>
    differences.find((d) => d.field === field)?.reason;

  return (
    <div>
      {/* Warnings */}
      {warnings.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          {warnings.map((warning) => (
            <Alert
              key={warning}
              message={warning}
              type="warning"
              showIcon
              icon={<WarningOutlined />}
              style={{ marginBottom: 8 }}
            />
          ))}
        </div>
      )}

      {/* Side-by-side comparison */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr auto 1fr",
          gap: 16,
          alignItems: "start",
        }}
      >
        {/* Source panel */}
        <Card
          size="small"
          title={
            <Space>
              <Tag>Source</Tag>
              <span style={{ fontSize: 13 }}>{source.sku}</span>
            </Space>
          }
          styles={{ body: { padding: 12 } }}
        >
          <Descriptions column={1} size="small">
            <Descriptions.Item label="Title">
              <span
                style={{
                  backgroundColor: isDifferent("title")
                    ? DIFF_HIGHLIGHT
                    : "transparent",
                  padding: isDifferent("title") ? "2px 4px" : 0,
                  borderRadius: 3,
                }}
              >
                {source.title}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Price">
              <span
                style={{
                  backgroundColor: isDifferent("price")
                    ? DIFF_HIGHLIGHT
                    : "transparent",
                  padding: isDifferent("price") ? "2px 4px" : 0,
                  borderRadius: 3,
                }}
              >
                {formatPrice(source.price)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Stock">
              <span
                style={{
                  backgroundColor: isDifferent("stock")
                    ? DIFF_HIGHLIGHT
                    : "transparent",
                  padding: isDifferent("stock") ? "2px 4px" : 0,
                  borderRadius: 3,
                }}
              >
                {source.stock}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Images">
              <Space size={4}>
                {source.images.slice(0, 4).map((img) => (
                  <Image
                    key={img}
                    src={img}
                    width={48}
                    height={48}
                    style={{ objectFit: "cover", borderRadius: 3 }}
                    preview={false}
                  />
                ))}
                {source.images.length > 4 && (
                  <Tag>+{source.images.length - 4}</Tag>
                )}
              </Space>
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Arrow */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            height: "100%",
            paddingTop: 60,
          }}
        >
          <ArrowRightOutlined
            style={{ fontSize: 24, color: token.colorTextSecondary }}
          />
        </div>

        {/* Target panel */}
        <Card
          size="small"
          title={
            <Space>
              <Tag
                style={{
                  background: platformStyle.gradient,
                  color: "white",
                  border: "none",
                }}
              >
                {platformStyle.label}
              </Tag>
              <span style={{ fontSize: 13 }}>Clone Target</span>
            </Space>
          }
          styles={{ body: { padding: 12 } }}
        >
          <Descriptions column={1} size="small">
            <Descriptions.Item
              label={
                <Space>
                  Title
                  {isDifferent("title") && (
                    <WarningOutlined
                      style={{ color: "#faad14", fontSize: 12 }}
                    />
                  )}
                </Space>
              }
            >
              <span
                style={{
                  backgroundColor: isDifferent("title")
                    ? DIFF_HIGHLIGHT
                    : "transparent",
                  padding: isDifferent("title") ? "2px 4px" : 0,
                  borderRadius: 3,
                }}
              >
                {target.title}
                {isDifferent("title") && (
                  <div
                    style={{
                      fontSize: 11,
                      color: "#d48806",
                      marginTop: 2,
                    }}
                  >
                    {getDiffReason("title")}
                  </div>
                )}
              </span>
            </Descriptions.Item>
            <Descriptions.Item
              label={
                <Space>
                  Price
                  {isDifferent("price") && (
                    <WarningOutlined
                      style={{ color: "#faad14", fontSize: 12 }}
                    />
                  )}
                </Space>
              }
            >
              <span
                style={{
                  backgroundColor: isDifferent("price")
                    ? DIFF_HIGHLIGHT
                    : "transparent",
                  padding: isDifferent("price") ? "2px 4px" : 0,
                  borderRadius: 3,
                }}
              >
                {formatPrice(target.price)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Stock">
              <span
                style={{
                  backgroundColor: isDifferent("stock")
                    ? DIFF_HIGHLIGHT
                    : "transparent",
                  padding: isDifferent("stock") ? "2px 4px" : 0,
                  borderRadius: 3,
                }}
              >
                {target.stock}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Images">
              <Space size={4}>
                {target.images.slice(0, 4).map((img) => (
                  <Image
                    key={img}
                    src={img}
                    width={48}
                    height={48}
                    style={{ objectFit: "cover", borderRadius: 3 }}
                    preview={false}
                  />
                ))}
                {target.images.length > 4 && (
                  <Tag>+{target.images.length - 4}</Tag>
                )}
              </Space>
            </Descriptions.Item>
          </Descriptions>
        </Card>
      </div>

      <DiffSummarySection differences={differences} />
    </div>
  );
};
