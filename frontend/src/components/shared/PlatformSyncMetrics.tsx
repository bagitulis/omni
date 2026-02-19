import { Alert, Space, Tag, Typography } from "antd";
import type { ImportFromStagingResult } from "@/api/products";
import type { Platform } from "./platformSyncPanelHelpers";

interface PlatformSyncMetricsProps {
  platform: Platform;
  result?: ImportFromStagingResult;
}

export function PlatformSyncMetrics({
  platform,
  result,
}: PlatformSyncMetricsProps) {
  if (!result) {
    return null;
  }

  return (
    <div style={{ marginTop: 10 }} data-testid={`platform-result-${platform}`}>
      <Space wrap size={[6, 6]}>
        <Tag color="success" bordered={false}>
          Created P: {result.products_created}
        </Tag>
        <Tag color="processing" bordered={false}>
          Matched: {result.products_matched}
        </Tag>
        <Tag color="blue" bordered={false}>
          SKUs: {result.skus_created}
        </Tag>
        <Tag color="purple" bordered={false}>
          Links: {result.links_created}
        </Tag>
      </Space>

      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
        Skipped products: {result.products_skipped} | Skipped SKUs:{" "}
        {result.skus_skipped}
      </Typography.Text>

      {result.errors.length > 0 ? (
        <Alert
          message={`${result.errors.length} import error(s)`}
          description={
            <div style={{ maxHeight: 96, overflowY: "auto", marginTop: 4 }}>
              {result.errors.map((err) => (
                <Typography.Text
                  key={`${platform}-${err}`}
                  style={{ display: "block", fontSize: 12 }}
                >
                  {err}
                </Typography.Text>
              ))}
            </div>
          }
          type="error"
          showIcon
          style={{ marginTop: 8 }}
        />
      ) : null}
    </div>
  );
}
