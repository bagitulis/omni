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
import {
  SettingOutlined,
  DownloadOutlined,
  FileTextOutlined,
} from "@ant-design/icons";
import type { AnalyticsSettings } from "@/types/analytics";
import { getAvailableColumns } from "@/api/inventory";
import { formatCurrency } from "@/lib/analyticsHelpers";
import { logger } from "@/lib/logger";

const { Text, Paragraph } = Typography;

interface Props {
  open: boolean;
  settings: AnalyticsSettings;
  onClose: () => void;
  onSave: (settings: AnalyticsSettings) => void;
  loading?: boolean;
}

export function TiktokAnalyticsSettingsModal({
  open,
  settings,
  onClose,
  onSave,
  loading,
}: Props) {
  const { token } = theme.useToken();
  const [form] = Form.useForm<AnalyticsSettings>();
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
      // Validation failed
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
      title={
        <Space>
          <SettingOutlined />
          <span>TikTok Analytics Settings</span>
        </Space>
      }
      width={500}
      onCancel={onClose}
      onOk={handleSave}
      confirmLoading={loading}
      okText={
        <span>
          <DownloadOutlined /> Save Settings
        </span>
      }
      cancelText="Cancel"
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={settings}
        style={{ marginTop: 20 }}
      >
        <Form.Item
          name="price_column"
          label="Price Column Name"
          rules={[{ required: true, message: "Please select a price column" }]}
          help="Column name in inventory that contains marketplace price"
        >
          <Select
            placeholder="-- Select Column --"
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

        <Form.Item label="Formula: Expected Income">
          <div
            style={{
              padding: "12px 16px",
              background: token.colorFillTertiary,
              borderRadius: token.borderRadius,
              border: `1px solid ${token.colorBorder}`,
              marginBottom: 8,
            }}
          >
            <Text code style={{ color: token.colorSuccessText }}>
              (Harga Marketplace - {deduction || 0}) × {multiplier || 0}
            </Text>
          </div>
        </Form.Item>

        <Space
          align="start"
          size={16}
          style={{ width: "100%", display: "flex" }}
        >
          <Form.Item
            name="formula_deduction"
            label="Admin Fee (Rp)"
            rules={[{ required: true }]}
            style={{ flex: 1, marginBottom: 0 }}
            help="TikTok admin fee deduction"
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
            help="After-fee percentage (0.84 = 84%)"
          >
            <InputNumber
              style={{ width: "100%" }}
              min={0}
              max={1}
              step={0.01}
            />
          </Form.Item>
        </Space>

        <div
          style={{
            marginTop: 20,
            padding: 16,
            background: token.colorFillQuaternary,
            borderRadius: token.borderRadius,
            border: `1px solid ${token.colorBorderSecondary}`,
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 6,
              marginBottom: 12,
            }}
          >
            <FileTextOutlined style={{ color: token.colorTextSecondary }} />
            <Text
              strong
              style={{ fontSize: "0.9rem", color: token.colorTextSecondary }}
            >
              Example Calculation
            </Text>
          </div>

          <div style={{ fontSize: "0.9rem" }}>
            <Paragraph style={{ margin: "6px 0" }}>
              If Harga Marketplace = <Text strong>Rp 100,000</Text>
            </Paragraph>
            <Paragraph style={{ margin: "6px 0" }}>
              Expected = (100,000 - {deduction || 0}) × {multiplier || 0}
            </Paragraph>
            <div
              style={{
                marginTop: 12,
                paddingTop: 12,
                borderTop: `1px dashed ${token.colorBorder}`,
                color: token.colorSuccessText,
                fontSize: "1rem",
                fontWeight: 600,
              }}
            >
              = {formatCurrency(exampleResult)}
            </div>
          </div>
        </div>
      </Form>
    </Modal>
  );
}

export default TiktokAnalyticsSettingsModal;
