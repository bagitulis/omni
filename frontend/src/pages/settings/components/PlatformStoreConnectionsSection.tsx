import { Card, Col, Row } from "antd";
import type { CredentialPlatformSummary } from "@/api/credentials";
import { PlatformCard } from "./PlatformCard";
import type { PlatformConnectionSummary } from "./PlatformCard";
import {
  formatCredentialExpiry,
  toConnectionSummary,
} from "../tabs/platformsTabUtils";

interface PlatformCardViewModel {
  platform: CredentialPlatformSummary;
  color: string;
  icon: React.ReactNode;
  name: string;
}

interface PlatformStoreConnectionsSectionProps {
  platformCards: PlatformCardViewModel[];
  storeActionReason: string | null;
  destructiveActionKey: string | null;
  onConnect: (platform: PlatformConnectionSummary) => void;
  onDisconnect: (platform: PlatformConnectionSummary) => void;
  onViewHistory: (platform: CredentialPlatformSummary) => void;
}

export function PlatformStoreConnectionsSection({
  platformCards,
  storeActionReason,
  destructiveActionKey,
  onConnect,
  onDisconnect,
  onViewHistory,
}: PlatformStoreConnectionsSectionProps) {
  return (
    <Card title="Store Connections" size="small" style={{ marginBottom: 24 }}>
      <Row gutter={[16, 16]}>
        {platformCards.map(({ platform, color, icon, name }) => (
          <Col xs={24} sm={12} lg={8} key={platform.platform}>
            <PlatformCard
              platform={toConnectionSummary(platform)}
              color={color}
              icon={icon}
              name={name}
              onConnect={onConnect}
              onDisconnect={onDisconnect}
              actionDisabledReason={storeActionReason}
              destructiveActionKey={destructiveActionKey}
              formatExpiry={formatCredentialExpiry}
              onViewHistory={(selectedPlatform) =>
                onViewHistory({
                  platform: selectedPlatform.platform,
                  region: platform.region,
                  status: platform.status,
                  app_configured: platform.app_configured,
                  secret_mask: platform.secret_mask,
                  app_secret_mask: platform.app_secret_mask,
                  stores: platform.stores,
                  app_config: platform.app_config,
                  audit_summary: platform.audit_summary,
                })
              }
            />
          </Col>
        ))}
      </Row>
    </Card>
  );
}
