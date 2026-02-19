import { useState, useCallback } from "react";
import { Button, Card, Typography, Space, Alert, Spin, Tag } from "antd";
import {
  importFromStaging,
  type ImportFromStagingResult,
} from "@/api/products";
import { fetchProductListFromApi } from "@/api/shopeeDb";
import { searchProducts } from "@/api/tiktokDb";
import { syncProductsToDb } from "@/api/lazadaDb";

type Platform = "shopee" | "tiktok" | "lazada";
type LoadingKey = `${Platform}-sync` | `${Platform}-import`;

export function PlatformSyncPanel() {
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

  const handleSyncToDb = useCallback(async (platform: Platform) => {
    const key: LoadingKey = `${platform}-sync`;
    setLoading((prev) => ({ ...prev, [key]: true }));
    setErrors((prev) => ({ ...prev, [key]: undefined }));
    setSyncMessages((prev) => ({ ...prev, [platform]: undefined }));

    try {
      let message = "";
      if (platform === "shopee") {
        const result = await fetchProductListFromApi();
        message = result.message ?? `Synced ${result.processed ?? 0} products`;
      } else if (platform === "tiktok") {
        const result = await searchProducts();
        message = `Synced ${result.length} products`;
      } else if (platform === "lazada") {
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

  const handleImportToMaster = useCallback(async (platform: Platform) => {
    const key: LoadingKey = `${platform}-import`;
    setLoading((prev) => ({ ...prev, [key]: true }));
    setErrors((prev) => ({ ...prev, [key]: undefined }));
    setResults((prev) => ({ ...prev, [platform]: undefined }));

    try {
      const result = await importFromStaging(platform);
      setResults((prev) => ({ ...prev, [platform]: result }));
    } catch (err) {
      setErrors((prev) => ({
        ...prev,
        [key]: err instanceof Error ? err.message : String(err),
      }));
    } finally {
      setLoading((prev) => ({ ...prev, [key]: false }));
    }
  }, []);

  const renderResult = (platform: Platform) => {
    const result = results[platform];
    if (!result) return null;

    return (
      <div style={{ marginTop: 8 }}>
        <Typography.Text type="success" style={{ fontSize: "12px" }}>
          ✓ Created: {result.products_created} products, {result.skus_created}{" "}
          SKUs, {result.links_created} links
        </Typography.Text>
        <br />
        <Typography.Text type="secondary" style={{ fontSize: "12px" }}>
          Matched: {result.products_matched} | Skipped:{" "}
          {result.products_skipped} SKUs skipped
        </Typography.Text>
        {result.errors.length > 0 && (
          <div style={{ marginTop: 4 }}>
            <Typography.Text type="danger" style={{ fontSize: "12px" }}>
              Errors: {result.errors.length}
            </Typography.Text>
            <div
              style={{ maxHeight: "100px", overflowY: "auto", marginTop: 4 }}
            >
              {result.errors.map((err, idx) => (
                <Alert
                  key={`${idx}-${err.substring(0, 10)}`}
                  message={err}
                  type="error"
                  style={{ marginBottom: 2, padding: "4px 8px" }}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    );
  };

  const platforms: { key: Platform; label: string; color: string }[] = [
    { key: "shopee", label: "Shopee", color: "#ee4d2d" },
    { key: "tiktok", label: "TikTok", color: "#000000" },
    { key: "lazada", label: "Lazada", color: "#0f146d" },
  ];

  return (
    <Card
      title="Platform Synchronization"
      size="small"
      style={{ borderRadius: "3px" }}
    >
      <Space direction="vertical" style={{ width: "100%" }} size="middle">
        {platforms.map((p) => (
          <div
            key={p.key}
            style={{
              padding: "8px",
              border: "1px solid #f0f0f0",
              borderRadius: "3px",
            }}
          >
            <Space align="center" style={{ marginBottom: 8 }}>
              <Tag color={p.color} style={{ borderRadius: "3px" }}>
                {p.label}
              </Tag>
              <Button
                size="small"
                onClick={() => handleSyncToDb(p.key)}
                loading={loading[`${p.key}-sync`]}
                style={{ borderRadius: "3px", fontSize: "12px" }}
              >
                Sync to DB
              </Button>
              <Button
                size="small"
                type="primary"
                onClick={() => handleImportToMaster(p.key)}
                loading={loading[`${p.key}-import`]}
                style={{ borderRadius: "3px", fontSize: "12px" }}
              >
                Import to Master
              </Button>
            </Space>

            {loading[`${p.key}-sync`] || loading[`${p.key}-import`] ? (
              <div style={{ marginTop: 8 }}>
                <Spin size="small" />
              </div>
            ) : null}

            {syncMessages[p.key] && (
              <div style={{ marginTop: 8 }}>
                <Alert
                  message={syncMessages[p.key]}
                  type="success"
                  showIcon
                  style={{
                    borderRadius: "3px",
                    fontSize: "12px",
                    padding: "4px 8px",
                  }}
                />
              </div>
            )}

            {errors[`${p.key}-sync`] && (
              <div style={{ marginTop: 8 }}>
                <Alert
                  message={errors[`${p.key}-sync`]}
                  type="error"
                  showIcon
                  style={{
                    borderRadius: "3px",
                    fontSize: "12px",
                    padding: "4px 8px",
                  }}
                />
              </div>
            )}

            {errors[`${p.key}-import`] && (
              <div style={{ marginTop: 8 }}>
                <Alert
                  message={errors[`${p.key}-import`]}
                  type="error"
                  showIcon
                  style={{
                    borderRadius: "3px",
                    fontSize: "12px",
                    padding: "4px 8px",
                  }}
                />
              </div>
            )}

            {renderResult(p.key)}
          </div>
        ))}
      </Space>
    </Card>
  );
}
