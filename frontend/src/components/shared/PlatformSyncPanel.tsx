import { useCallback, useMemo, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Col,
  Row,
  Space,
  Spin,
  Tag,
  Typography,
  theme,
} from "antd";
import { syncProductsToDb } from "@/api/lazadaDb";
import {
  importFromStaging,
  type ImportFromStagingResult,
} from "@/api/products";
import { syncShopeeProducts } from "@/api/shopeeDb";
import { searchProducts } from "@/api/tiktokDb";
import { PLATFORM_META, type Platform } from "./platformSyncPanelHelpers";
import { PlatformSyncMetrics } from "./PlatformSyncMetrics";

type LoadingKey = `${Platform}-sync` | `${Platform}-import`;

interface PlatformSyncPanelProps {
  onImportCompleted?: () => Promise<void> | void;
}

export function PlatformSyncPanel({
  onImportCompleted,
}: PlatformSyncPanelProps) {
  const { token } = theme.useToken();
  const [loading, setLoading] = useState<Partial<Record<LoadingKey, boolean>>>(
    {},
  );
  const [results, setResults] = useState<
    Partial<Record<Platform, ImportFromStagingResult>>
  >({});
  const [syncMessages, setSyncMessages] = useState<
    Partial<Record<Platform, string>>
  >({});
  const [errors, setErrors] = useState<Partial<Record<LoadingKey, string>>>({});

  const totals = useMemo(() => {
    return Object.values(results).reduce(
      (acc, result) => {
        if (!result) {
          return acc;
        }

        return {
          productsCreated: acc.productsCreated + result.products_created,
          productsMatched: acc.productsMatched + result.products_matched,
          skusCreated: acc.skusCreated + result.skus_created,
          linksCreated: acc.linksCreated + result.links_created,
          errors: acc.errors + result.errors.length,
        };
      },
      {
        productsCreated: 0,
        productsMatched: 0,
        skusCreated: 0,
        linksCreated: 0,
        errors: 0,
      },
    );
  }, [results]);

  const handleSyncToDb = useCallback(async (platform: Platform) => {
    const key: LoadingKey = `${platform}-sync`;
    setLoading((prev) => ({ ...prev, [key]: true }));
    setErrors((prev) => ({ ...prev, [key]: undefined }));
    setSyncMessages((prev) => ({ ...prev, [platform]: undefined }));

    try {
      let message = "";
      if (platform === "shopee") {
        const result = await syncShopeeProducts();
        message = result.message ?? `Synced ${result.processed ?? 0} products`;
      } else if (platform === "tiktok") {
        const result = await searchProducts();
        message = `Synced ${result.length} products`;
      } else {
        const result = await syncProductsToDb();
        message = result.message ?? `Synced ${result.processed ?? 0} products`;
      }

      setSyncMessages((prev) => ({ ...prev, [platform]: message }));
    } catch (err) {
      setErrors((prev) => ({
        ...prev,
        [key]: err instanceof Error ? err.message : String(err),
      }));
    } finally {
      setLoading((prev) => ({ ...prev, [key]: false }));
    }
  }, []);

  const handleImportToMaster = useCallback(
    async (platform: Platform) => {
      const key: LoadingKey = `${platform}-import`;
      setLoading((prev) => ({ ...prev, [key]: true }));
      setErrors((prev) => ({ ...prev, [key]: undefined }));
      setResults((prev) => ({ ...prev, [platform]: undefined }));

      try {
        const result = await importFromStaging(platform);
        setResults((prev) => ({ ...prev, [platform]: result }));
        if (onImportCompleted) {
          await onImportCompleted();
        }
      } catch (err) {
        setErrors((prev) => ({
          ...prev,
          [key]: err instanceof Error ? err.message : String(err),
        }));
      } finally {
        setLoading((prev) => ({ ...prev, [key]: false }));
      }
    },
    [onImportCompleted],
  );

  return (
    <div style={{ width: "100%" }} data-testid="platform-sync-panel">
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            gap: 16,
            flexWrap: "wrap",
          }}
        >
          <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
            <Typography.Text style={{ fontSize: 16, fontWeight: 600 }}>
              Platform Synchronization
            </Typography.Text>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              Sync platform data to staging, then import to master catalog.
            </Typography.Text>
          </div>

          <Space wrap size={[8, 8]}>
            <Tag color="success" bordered={false}>
              Created: {totals.productsCreated}
            </Tag>
            <Tag color="processing" bordered={false}>
              Matched: {totals.productsMatched}
            </Tag>
            <Tag color="blue" bordered={false}>
              SKUs: {totals.skusCreated}
            </Tag>
            <Tag color="purple" bordered={false}>
              Links: {totals.linksCreated}
            </Tag>
            <Tag
              color={totals.errors > 0 ? "error" : "default"}
              bordered={false}
            >
              Errors: {totals.errors}
            </Tag>
          </Space>
        </div>

        <Row gutter={[16, 16]}>
          {PLATFORM_META.map((platform) => {
            const syncLoading = Boolean(loading[`${platform.key}-sync`]);
            const importLoading = Boolean(loading[`${platform.key}-import`]);
            const isBusy = syncLoading || importLoading;
            const hasError =
              Boolean(errors[`${platform.key}-sync`]) ||
              Boolean(errors[`${platform.key}-import`]);

            return (
              <Col xs={24} md={12} xl={8} key={platform.key}>
                <Card
                  size="small"
                  style={{ borderRadius: token.borderRadius, height: "100%" }}
                  data-testid={`platform-sync-card-${platform.key}`}
                >
                  <Space
                    direction="vertical"
                    size={12}
                    style={{ width: "100%" }}
                  >
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                        gap: 8,
                      }}
                    >
                      <Space align="center" size={8}>
                        <Tag
                          color={platform.color}
                          bordered={false}
                          style={{ margin: 0 }}
                        >
                          {platform.label}
                        </Tag>
                        <Typography.Text
                          type="secondary"
                          style={{ fontSize: 12 }}
                        >
                          Staging to Master
                        </Typography.Text>
                      </Space>

                      {isBusy ? <Spin size="small" /> : null}
                    </div>

                    <Space.Compact block>
                      <Button
                        onClick={() => handleSyncToDb(platform.key)}
                        loading={syncLoading}
                        disabled={importLoading}
                        data-testid={`platform-sync-db-${platform.key}`}
                      >
                        Sync to DB
                      </Button>
                      <Button
                        type="primary"
                        onClick={() => handleImportToMaster(platform.key)}
                        loading={importLoading}
                        disabled={syncLoading}
                        data-testid={`platform-import-master-${platform.key}`}
                      >
                        Import
                      </Button>
                    </Space.Compact>

                    {syncMessages[platform.key] ? (
                      <Alert
                        message={syncMessages[platform.key]}
                        type="success"
                        showIcon
                      />
                    ) : null}

                    {errors[`${platform.key}-sync`] ? (
                      <Alert
                        message={errors[`${platform.key}-sync`]}
                        type="error"
                        showIcon
                      />
                    ) : null}

                    {errors[`${platform.key}-import`] ? (
                      <Alert
                        message={errors[`${platform.key}-import`]}
                        type="error"
                        showIcon
                      />
                    ) : null}

                    {!isBusy && !hasError && !syncMessages[platform.key] ? (
                      <Typography.Text
                        type="secondary"
                        style={{ fontSize: 12 }}
                      >
                        Run sync first when staging data is stale, then import.
                      </Typography.Text>
                    ) : null}

                    <PlatformSyncMetrics
                      platform={platform.key}
                      result={results[platform.key]}
                    />
                  </Space>
                </Card>
              </Col>
            );
          })}
        </Row>
      </Space>
    </div>
  );
}
