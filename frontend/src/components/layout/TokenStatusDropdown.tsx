import { useState, useEffect } from "react";
import {
  Dropdown,
  Button,
  Card,
  Space,
  Typography,
  Tag,
  Spin,
  Tooltip,
  Flex,
  Badge,
} from "antd";
import {
  KeyOutlined,
  ReloadOutlined,
  CheckCircleFilled,
  CloseCircleFilled,
  WarningFilled,
} from "@ant-design/icons";
import apiClient from "@/api/client";

const { Text } = Typography;

interface TokenStatusData {
  isExpired: boolean;
  expiresAt: string | null;
  refreshTokenExpiresAt: string | null;
  status?: string;
  valid?: boolean;
}

interface TokenStatusMap {
  shopee?: TokenStatusData;
  tiktok?: TokenStatusData;
  lazada?: TokenStatusData;
  [key: string]: TokenStatusData | undefined;
}

// Backend response format from /platform-auth/status
interface BackendPlatformStatus {
  connected: boolean;
  expired: boolean;
  expires_soon: boolean;
  expires_at: number;
  refresh_token_expires_at?: number;
  shop_id?: string;
  shop_name?: string;
}

interface BackendTokenStatusResponse {
  [platform: string]: BackendPlatformStatus;
}

const PLATFORM_CONFIG: Record<string, { color: string; label: string }> = {
  shopee: { color: "#ee4d2d", label: "Shopee" },
  tiktok: { color: "#000000", label: "TikTok" },
  lazada: { color: "#0f146d", label: "Lazada" },
};

export function TokenStatusDropdown() {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [tokenStatus, setTokenStatus] = useState<TokenStatusMap | null>(null);

  const loadTokenStatus = async () => {
    setLoading(true);
    try {
      // Use existing platform-auth/status endpoint
      const response = await apiClient.get<BackendTokenStatusResponse>(
        "/platform-auth/status",
      );
      if (response.success && response.data) {
        // Transform backend format to frontend format
        const transformed: TokenStatusMap = {};
        Object.entries(response.data).forEach(([platform, platformData]) => {
          transformed[platform] = {
            isExpired: platformData.expired,
            expiresAt: platformData.expires_at
              ? new Date(platformData.expires_at).toISOString()
              : null,
            refreshTokenExpiresAt: platformData.refresh_token_expires_at
              ? new Date(platformData.refresh_token_expires_at).toISOString()
              : null,
            status: platformData.expired
              ? "expired"
              : platformData.expires_soon
                ? "expiring"
                : platformData.connected
                  ? "valid"
                  : "not_configured",
            valid: platformData.connected && !platformData.expired,
          };
        });
        setTokenStatus(transformed);
      }
    } catch (error) {
      console.error("Failed to load token status:", error);
    } finally {
      setLoading(false);
    }
  };

  const refreshTokens = async () => {
    setLoading(true);
    try {
      // Force refresh all platform tokens
      await apiClient.post("/tokens/refresh-all?force=true");
      await loadTokenStatus();
    } catch (error) {
      console.error("Failed to refresh tokens:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && !tokenStatus) {
      loadTokenStatus();
    }
  }, [open]);

  const getTokenStatusType = (
    data: TokenStatusData | undefined,
  ): "valid" | "expiring" | "expired" | "unknown" => {
    if (!data) return "unknown";
    if (data.isExpired || data.status === "expired" || data.valid === false) {
      return "expired";
    }
    if (data.expiresAt) {
      const expiresAt = new Date(data.expiresAt).getTime();
      const hoursLeft = (expiresAt - Date.now()) / (1000 * 60 * 60);
      if (hoursLeft < 24 && hoursLeft > 0) return "expiring";
    }
    return "valid";
  };

  const formatTimeRemaining = (dateString: string | null): string => {
    if (!dateString) return "Unknown";
    const expiresAt = new Date(dateString);
    const diffMs = expiresAt.getTime() - Date.now();
    if (diffMs <= 0) return "Expired";

    const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    const hours = Math.floor(
      (diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60),
    );
    const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const hasIssues = tokenStatus
    ? Object.values(tokenStatus).some((s) => {
        const status = getTokenStatusType(s);
        return status === "expired" || status === "expiring";
      })
    : false;

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "valid":
        return <CheckCircleFilled style={{ color: "#52c41a" }} />;
      case "expiring":
        return <WarningFilled style={{ color: "#faad14" }} />;
      case "expired":
        return <CloseCircleFilled style={{ color: "#ff4d4f" }} />;
      default:
        return <WarningFilled style={{ color: "#999" }} />;
    }
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case "valid":
        return "success";
      case "expiring":
        return "warning";
      case "expired":
        return "error";
      default:
        return "default";
    }
  };

  const dropdownContent = (
    <Card
      size="small"
      title={
        <Flex justify="space-between" align="center">
          <Text strong style={{ fontSize: 13 }}>
            <KeyOutlined style={{ marginRight: 6 }} />
            Platform Token Status
          </Text>
          <Button
            type="text"
            size="small"
            icon={<ReloadOutlined spin={loading} />}
            onClick={refreshTokens}
            disabled={loading}
          >
            Refresh
          </Button>
        </Flex>
      }
      style={{ width: 320, boxShadow: "0 4px 12px rgba(0,0,0,0.15)" }}
      styles={{ body: { padding: "8px 0" } }}
    >
      {loading && !tokenStatus ? (
        <Flex justify="center" style={{ padding: 24 }}>
          <Spin />
        </Flex>
      ) : tokenStatus ? (
        <Space direction="vertical" size={0} style={{ width: "100%" }}>
          {Object.entries(tokenStatus).map(([platform, data]) => {
            const config = PLATFORM_CONFIG[platform] || {
              color: "#666",
              label: platform,
            };
            const status = getTokenStatusType(data);

            return (
              <div
                key={platform}
                style={{
                  padding: "10px 16px",
                  borderBottom: "1px solid #f0f0f0",
                  background:
                    status === "expired"
                      ? "#fff1f0"
                      : status === "expiring"
                        ? "#fffbe6"
                        : "transparent",
                }}
              >
                <Flex justify="space-between" align="center">
                  <Flex align="center" gap={8}>
                    <Tag
                      color={config.color}
                      style={{
                        margin: 0,
                        fontSize: 10,
                        padding: "0 6px",
                        borderRadius: 3,
                      }}
                    >
                      {config.label.toUpperCase()}
                    </Tag>
                    {getStatusIcon(status)}
                  </Flex>
                  <Tag color={getStatusColor(status)} bordered={false}>
                    {status === "valid"
                      ? "Valid"
                      : status === "expiring"
                        ? "Expiring Soon"
                        : status === "expired"
                          ? "Expired"
                          : "Unknown"}
                  </Tag>
                </Flex>

                <Flex
                  vertical
                  gap={2}
                  style={{ marginTop: 8, fontSize: 11, color: "#666" }}
                >
                  <Flex justify="space-between">
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      Access Token:
                    </Text>
                    <Text
                      style={{
                        fontSize: 11,
                        color: data?.isExpired ? "#ff4d4f" : "#52c41a",
                      }}
                    >
                      {data?.isExpired ? "Expired" : "Valid"} •{" "}
                      {formatTimeRemaining(data?.expiresAt ?? null)}
                    </Text>
                  </Flex>
                  <Flex justify="space-between">
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      Refresh Token:
                    </Text>
                    <Text style={{ fontSize: 11 }}>
                      {formatTimeRemaining(data?.refreshTokenExpiresAt ?? null)}
                    </Text>
                  </Flex>
                </Flex>
              </div>
            );
          })}
        </Space>
      ) : (
        <Flex justify="center" style={{ padding: 24 }}>
          <Text type="secondary">Click refresh to load status</Text>
        </Flex>
      )}
    </Card>
  );

  return (
    <Dropdown
      dropdownRender={() => dropdownContent}
      trigger={["click"]}
      open={open}
      onOpenChange={setOpen}
      placement="bottomRight"
    >
      <Tooltip title="Platform Token Status">
        <Badge dot={hasIssues} offset={[-2, 2]}>
          <Button
            type="text"
            icon={
              <KeyOutlined
                style={{ color: hasIssues ? "#faad14" : undefined }}
              />
            }
            style={{ height: 40, width: 40 }}
          />
        </Badge>
      </Tooltip>
    </Dropdown>
  );
}
