import { useState } from "react";
import { Card, Col, Row, Slider, Space, Switch, Typography, theme } from "antd";

const { Text } = Typography;

interface CacheSettingsState {
  enabled: boolean;
  ttl: number;
  auto_refresh: boolean;
}

export function RouteManagementCacheSettings() {
  const { token } = theme.useToken();
  const [cache_settings, setCacheSettings] = useState<CacheSettingsState>({
    enabled: true,
    ttl: 300,
    auto_refresh: false,
  });

  return (
    <Card
      size="small"
      title="Route Cache Settings"
      style={{ borderRadius: token.borderRadius }}
    >
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Text>Enable Caching</Text>
          </Col>
          <Col>
            <Switch
              checked={cache_settings.enabled}
              onChange={(checked) =>
                setCacheSettings((prev) => ({
                  ...prev,
                  enabled: checked,
                }))
              }
            />
          </Col>
        </Row>
        <Row gutter={16} align="middle">
          <Col xs={24} md={10}>
            <Text>Cache TTL: {cache_settings.ttl}s</Text>
          </Col>
          <Col xs={24} md={14}>
            <Slider
              disabled={!cache_settings.enabled}
              min={0}
              max={3600}
              step={30}
              value={cache_settings.ttl}
              tooltip={{ formatter: (value) => `${value ?? 0}s` }}
              onChange={(value) =>
                setCacheSettings((prev) => ({
                  ...prev,
                  ttl: typeof value === "number" ? value : prev.ttl,
                }))
              }
            />
          </Col>
        </Row>
        <Row justify="space-between" align="middle">
          <Col>
            <Text>Auto Refresh</Text>
          </Col>
          <Col>
            <Switch
              disabled={!cache_settings.enabled}
              checked={cache_settings.auto_refresh}
              onChange={(checked) =>
                setCacheSettings((prev) => ({
                  ...prev,
                  auto_refresh: checked,
                }))
              }
            />
          </Col>
        </Row>
      </Space>
    </Card>
  );
}
