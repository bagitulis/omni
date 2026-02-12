import { useState, useEffect, useCallback } from "react";
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
  message,
  theme,
} from "antd";
import { KeyOutlined, ReloadOutlined } from "@ant-design/icons";
import apiClient from "@/api/client";
import {
  TokenStatusMap,
  BackendTokenStatusResponse,
  PLATFORM_CONFIG,
} from "./TokenStatusDropdown.types";
import {
  getTokenStatusType,
  formatTimeRemaining,
  getStatusIcon,
  getStatusColor,
} from "./TokenStatusDropdown.utils";

const { Text } = Typography;

export function TokenStatusDropdown() {
  const { token } = theme.useToken();
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [tokenStatus, setTokenStatus] = useState<TokenStatusMap | null>(null);

  const loadTokenStatus = useCallback(async () => {
    setLoading(true);
    try {
      const response = await apiClient.get<BackendTokenStatusResponse>(
        "/platform-auth/status",
      );
      if (response.success && response.data) {
        const transformed: TokenStatusMap = {};
        Object.entries(response.data).forEach(([platform, platformData]) => {
          transformed[platform] = {
            isExpired: platformData.expired,
            expiresAt:
              platformData.expires_at && platformData.expires_at > 0
                ? new Date(platformData.expires_at).toISOString()
                : null,
            refreshTokenExpiresAt:
              platformData.refresh_token_expires_at &&
              platformData.refresh_token_expires_at > 0
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
      // Error silently handled - loading state resets in finally block
    } finally {
      setLoading(false);
    }
  }, []);

  const refreshTokens = async () => {
    setLoading(true);
    message.loading({ content: "Refreshing tokens...", key: "refresh" });
    try {
      await apiClient.post("/tokens/refresh-all?force=true");
      await loadTokenStatus();
      message.success({
        content: "Tokens refreshed successfully",
        key: "refresh",
      });
    } catch (error) {
      message.error({ content: "Failed to refresh tokens", key: "refresh" });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && !tokenStatus) {
      loadTokenStatus();
    }
  }, [open, tokenStatus, loadTokenStatus]);

  const hasIssues = tokenStatus
    ? Object.values(tokenStatus).some((s) => {
        const status = getTokenStatusType(s);
        return status === "expired" || status === "expiring";
      })
    : false;

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
      style={{ width: 320, boxShadow: token.boxShadowSecondary }}
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
              color: "default",
              label: platform,
            };
            const status = getTokenStatusType(data);

            return (
              <div
                key={platform}
                style={{
                  padding: "10px 16px",
                  borderBottom: `1px solid ${token.colorBorder}`,
                  background:
                    status === "expired"
                      ? token.colorErrorBg
                      : status === "expiring"
                        ? token.colorWarningBg
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
                          : status === "not_configured"
                            ? "Not Configured"
                            : "Unknown"}
                  </Tag>
                </Flex>

                <Flex
                  vertical
                  gap={2}
                  style={{
                    marginTop: 8,
                    fontSize: 11,
                    color: token.colorTextSecondary,
                  }}
                >
                  <Flex justify="space-between">
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      Access Token:
                    </Text>
                    <Text
                      style={{
                        fontSize: 11,
                        color: data?.isExpired
                          ? token.colorError
                          : token.colorSuccess,
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
                style={{ color: hasIssues ? token.colorWarning : undefined }}
              />
            }
            style={{ height: 40, width: 40 }}
          />
        </Badge>
      </Tooltip>
    </Dropdown>
  );
}
