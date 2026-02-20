import {
  Card,
  Descriptions,
  Empty,
  Space,
  Spin,
  Typography,
} from "antd";
import type { LinkConfig } from "./GoogleSheetsLinkRow";

const { Text } = Typography;

interface SheetMeta {
  name: string;
  sheet_id: number;
  column_count: number;
  row_count: number;
}

interface SheetMetadataCardProps {
  linkConfigs: LinkConfig[];
  sheetsMetadata?: Partial<Record<string, SheetMeta[]>>;
  lastUpdated?: string;
  isLoading: boolean;
}

/**
 * Renders the "Sheet Metadata" card section of GoogleSheetsTab.
 * Extracted for SRP — metadata display is a distinct concern from link management.
 */
export function SheetMetadataCard({
  linkConfigs,
  sheetsMetadata,
  lastUpdated,
  isLoading,
}: SheetMetadataCardProps) {
  const hasMetadata =
    sheetsMetadata &&
    Object.values(sheetsMetadata).some((list) => list && list.length > 0);

  return (
    <Card title="Sheet Metadata" style={{ marginTop: 16 }}>
      <Spin spinning={isLoading}>
        {!hasMetadata && <Empty description="No sheet metadata available" />}

        {hasMetadata && (
          <Space direction="vertical" size={12} style={{ width: "100%" }}>
            {linkConfigs.map((config) => {
              const metadataList = sheetsMetadata?.[config.type] ?? [];

              return (
                <Card
                  key={`${config.type}-metadata`}
                  size="small"
                  title={config.label}
                >
                  {metadataList.length === 0 ? (
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      No metadata found.
                    </Text>
                  ) : (
                    <Space direction="vertical" size={8} style={{ width: "100%" }}>
                      {metadataList.map((sheet) => (
                        <Descriptions
                          key={`${config.type}-${sheet.sheet_id}`}
                          size="small"
                          column={2}
                          bordered
                        >
                          <Descriptions.Item label="Sheet Name">
                            {sheet.name}
                          </Descriptions.Item>
                          <Descriptions.Item label="Sheet ID">
                            {sheet.sheet_id}
                          </Descriptions.Item>
                          <Descriptions.Item label="Column Count">
                            {sheet.column_count}
                          </Descriptions.Item>
                          <Descriptions.Item label="Row Count">
                            {sheet.row_count}
                          </Descriptions.Item>
                        </Descriptions>
                      ))}
                    </Space>
                  )}
                </Card>
              );
            })}

            {lastUpdated && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                Last updated: {new Date(lastUpdated).toLocaleString()}
              </Text>
            )}
          </Space>
        )}
      </Spin>
    </Card>
  );
}
