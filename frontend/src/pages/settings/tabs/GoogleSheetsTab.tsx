import { useEffect, useMemo, useState } from "react";
import {
  Button,
  Card,
  Descriptions,
  Empty,
  Form,
  Input,
  Space,
  Spin,
  Switch,
  Tag,
  Typography,
  message,
} from "antd";
import {
  useGoogleSheetsDetails,
  useGoogleSheetsLinks,
  useSaveLinks,
  useValidateLink,
} from "@/hooks/useGoogleSheets";
import type { SpreadsheetLinks, ValidationResult } from "@/types/googleSheets";

const { Text, Title } = Typography;

type LinkType = "inventory" | "wallet" | "shipping" | "order";
type LinkField = keyof SpreadsheetLinks;

interface LinkConfig {
  type: LinkType;
  field: LinkField;
  label: string;
}

const linkConfigs: LinkConfig[] = [
  {
    type: "inventory",
    field: "inventory_url",
    label: "Inventory Spreadsheet URL",
  },
  {
    type: "wallet",
    field: "wallet_url",
    label: "Wallet Spreadsheet URL",
  },
  {
    type: "shipping",
    field: "shipping_url",
    label: "Shipping Spreadsheet URL",
  },
  {
    type: "order",
    field: "order_url",
    label: "Order Spreadsheet URL",
  },
];

export default function GoogleSheetsTab() {
  const [form] = Form.useForm<SpreadsheetLinks>();
  const [isLocked, setIsLocked] = useState(false);
  const [validationResults, setValidationResults] = useState<
    Partial<Record<LinkType, ValidationResult>>
  >({});

  const { data: savedLinks, isLoading: isLinksLoading } =
    useGoogleSheetsLinks();
  const { data: details, isLoading: isDetailsLoading } =
    useGoogleSheetsDetails();
  const validateLinkMutation = useValidateLink();
  const saveLinksMutation = useSaveLinks();

  useEffect(() => {
    if (!savedLinks) {
      return;
    }

    form.setFieldsValue(savedLinks);
  }, [form, savedLinks]);

  const hasMetadata = useMemo(() => {
    if (!details?.sheets_metadata) {
      return false;
    }

    return Object.values(details.sheets_metadata).some(
      (metadataList) => metadataList.length > 0,
    );
  }, [details]);

  const handleValidate = (config: LinkConfig) => {
    const rawValue = form.getFieldValue(config.field);
    const spreadsheetUrl = rawValue?.trim();

    if (!spreadsheetUrl) {
      message.warning(`Please enter ${config.label.toLowerCase()}`);
      return;
    }

    validateLinkMutation.mutate(
      {
        spreadsheet_url: spreadsheetUrl,
        type: config.type,
      },
      {
        onSuccess: (result) => {
          setValidationResults((previous) => ({
            ...previous,
            [config.type]: result,
          }));
        },
      },
    );
  };

  const handleSaveAll = () => {
    const values = form.getFieldsValue();

    saveLinksMutation.mutate({
      inventory_url: values.inventory_url?.trim() || null,
      wallet_url: values.wallet_url?.trim() || null,
      shipping_url: values.shipping_url?.trim() || null,
      order_url: values.order_url?.trim() || null,
    });
  };

  const renderValidationResult = (type: LinkType) => {
    const result = validationResults[type];

    if (!result) {
      return null;
    }

    return (
      <div style={{ marginTop: 8 }}>
        <Space direction="vertical" size={4} style={{ width: "100%" }}>
          <Text style={{ fontSize: 12 }}>
            Spreadsheet: <Text strong>{result.name}</Text>
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Sheets found: {result.sheets.length}
          </Text>
          <Space size={[4, 4]} wrap>
            {result.sheets.map((sheet) => (
              <Tag key={`${result.spreadsheet_id}-${sheet.sheet_id}`}>
                {sheet.name}
              </Tag>
            ))}
          </Space>
        </Space>
      </div>
    );
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0, fontSize: 14 }}>
          Google Sheets Settings
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Configure spreadsheet links and review discovered sheet metadata.
        </Text>
      </div>

      <Card
        title="Spreadsheet Links"
        extra={
          <Space size={8}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Locked
            </Text>
            <Switch checked={isLocked} onChange={setIsLocked} />
          </Space>
        }
      >
        <Spin spinning={isLinksLoading}>
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              inventory_url: "",
              wallet_url: "",
              shipping_url: "",
              order_url: "",
            }}
          >
            {linkConfigs.map((config) => (
              <Form.Item
                key={config.field}
                label={config.label}
                name={config.field}
                style={{ marginBottom: 16 }}
              >
                <div>
                  <Space.Compact style={{ width: "100%" }}>
                    <Input
                      placeholder="https://docs.google.com/spreadsheets/d/..."
                      disabled={isLocked}
                    />
                    <Button
                      onClick={() => handleValidate(config)}
                      loading={validateLinkMutation.isPending}
                      disabled={isLocked}
                    >
                      Validate
                    </Button>
                  </Space.Compact>
                  {renderValidationResult(config.type)}
                </div>
              </Form.Item>
            ))}

            <Button
              type="primary"
              onClick={handleSaveAll}
              loading={saveLinksMutation.isPending}
              disabled={isLocked}
            >
              Save All Links
            </Button>
          </Form>
        </Spin>
      </Card>

      <Card title="Sheet Metadata" style={{ marginTop: 16 }}>
        <Spin spinning={isDetailsLoading}>
          {!hasMetadata && <Empty description="No sheet metadata available" />}

          {hasMetadata && (
            <Space direction="vertical" size={12} style={{ width: "100%" }}>
              {linkConfigs.map((config) => {
                const metadataList =
                  details?.sheets_metadata?.[config.type] ?? [];

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
                      <Space
                        direction="vertical"
                        size={8}
                        style={{ width: "100%" }}
                      >
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

              {details?.last_updated && (
                <Text type="secondary" style={{ fontSize: 12 }}>
                  Last updated:{" "}
                  {new Date(details.last_updated).toLocaleString()}
                </Text>
              )}
            </Space>
          )}
        </Spin>
      </Card>
    </div>
  );
}
