import {
  Modal,
  Form,
  Select,
  InputNumber,
  Typography,
  Space,
  theme,
} from "antd";
import { useState, useEffect, useMemo, useCallback } from "react";
import type { AnalyticsSettings } from "@/types/analytics";
import { getAvailableColumns } from "@/api/inventory";
import { formatCurrency } from "@/lib/analyticsHelpers";
import { logger } from "@/lib/logger";

const { Text, Title, Paragraph } = Typography;

interface Props {
  open: boolean;
  settings: AnalyticsSettings;
  onClose: () => void;
  onSave: (settings: AnalyticsSettings) => void;
  loading?: boolean;
}

export function SettingsModal({
  open,
  settings,
  onClose,
  onSave,
  loading,
}: Props) {
  const [form] = Form.useForm<AnalyticsSettings>();
  const [modalApi, contextHolder] = Modal.useModal();
  const { token } = theme.useToken();
  const [columns, setColumns] = useState<string[]>([]);
  const [fetchingColumns, setFetchingColumns] = useState(false);

  // Watch form values for example calculation
  const deduction = Form.useWatch("formula_deduction", form);
  const multiplier = Form.useWatch("formula_multiplier", form);

  const fetchColumns = useCallback(async () => {
    try {
      setFetchingColumns(true);
      const cols = await getAvailableColumns();
      setColumns(cols);
    } catch (error) {
      logger.error("Failed to fetch columns:", { error });
    } finally {
      setFetchingColumns(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      form.setFieldsValue(settings);
      fetchColumns();
    }
  }, [open, settings, form, fetchColumns]);

  const sortedColumns = useMemo(() => {
    return [...columns].sort((a, b) => {
      const aPriority = /harga|price|cost|nilai/i.test(a) ? 1 : 0;
      const bPriority = /harga|price|cost|nilai/i.test(b) ? 1 : 0;
      return bPriority - aPriority;
    });
  }, [columns]);

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      onSave(values);
    } catch (error) {
      modalApi.error({
        title: "Invalid settings",
        content: "Please correct the highlighted fields before saving.",
      });
    }
  };

  // Example calculation
  const examplePrice = 100000;
  const exampleResult = useMemo(() => {
    const d = deduction || 0;
    const m = multiplier || 0;
    return (examplePrice - d) * m;
  }, [deduction, multiplier]);

  return (
    <Modal
      open={open}
      title="Analytics Settings"
      width={500}
      onCancel={onClose}
      onOk={handleSave}
      confirmLoading={loading}
      okText="Save Settings"
    >
      {contextHolder}
      <Form
        form={form}
        layout="vertical"
        initialValues={settings}
        style={{ marginTop: 20 }}
      >
        <Form.Item
          name="price_column"
          label="Inventory Price Column"
          rules={[{ required: true, message: "Please select a price column" }]}
          tooltip="Select the column from your inventory that represents the base price (HPP)"
        >
          <Select
            placeholder="Select a column"
            loading={fetchingColumns}
            showSearch
            filterOption={(input, option) =>
              (option?.label ?? "").toLowerCase().includes(input.toLowerCase())
            }
            options={sortedColumns.map((col) => ({
              label: col,
              value: col,
            }))}
          />
        </Form.Item>

        <div
          style={{
            background: token?.colorFillAlter,
            padding: 16,
            borderRadius: token?.borderRadius ?? 6,
            marginBottom: 24,
            border: `1px solid ${token?.colorBorder}`,
          }}
        >
          <Title level={5} style={{ fontSize: 14, marginTop: 0 }}>
            Expected Income Formula
          </Title>
          <Paragraph style={{ fontFamily: "monospace", fontSize: 13 }}>
            (Marketplace Price - <Text type="danger">{deduction || 0}</Text>) ×{" "}
            <Text type="success">{multiplier || 0}</Text>
          </Paragraph>

          <Space align="start" size={16} style={{ width: "100%" }}>
            <Form.Item
              name="formula_deduction"
              label="Admin Fee (Rp)"
              rules={[{ required: true }]}
              style={{ flex: 1, marginBottom: 0 }}
            >
              <InputNumber<number>
                style={{ width: "100%" }}
                min={0}
                formatter={(value) =>
                  `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
                }
                parser={(value) => {
                  const val = value?.replace(/\Rp\s?|(,*)/g, "") || "";
                  return Number(val);
                }}
              />
            </Form.Item>

            <Form.Item
              name="formula_multiplier"
              label="Multiplier"
              rules={[{ required: true }]}
              style={{ flex: 1, marginBottom: 0 }}
              tooltip="Usually 0.90 to 0.95 depending on platform fees"
            >
              <InputNumber
                style={{ width: "100%" }}
                min={0}
                max={1}
                step={0.01}
              />
            </Form.Item>
          </Space>
        </div>

        <div style={{ marginTop: 16 }}>
          <Text strong style={{ fontSize: 12 }}>
            Example Calculation:
          </Text>
          <div
            style={{
              marginTop: 4,
              padding: 8,
              background: token?.colorSuccessBg,
              border: `1px solid ${token?.colorSuccessBorder}`,
              borderRadius: 4,
              fontSize: 12,
            }}
          >
            If Marketplace Price = {formatCurrency(examplePrice)} → Expected ={" "}
            <Text strong>{formatCurrency(exampleResult)}</Text>
          </div>
        </div>
      </Form>
    </Modal>
  );
}
