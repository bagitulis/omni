import {
  Modal,
  Form,
  Select,
  Checkbox,
  InputNumber,
  Button,
  TimePicker,
} from "antd";
import dayjs from "dayjs";
import { useEffect } from "react";

const { Option } = Select;

export interface AutoFunctionConfig {
  id?: string;
  name: string;
  interval_minutes: number;
  start_time?: string;
  end_time?: string;
  enabled: boolean;
}

interface ConfigEditorModalProps {
  open: boolean;
  onClose: () => void;
  config: AutoFunctionConfig | null;
  availableFunctions: Array<{ value: string; label: string }>;
  onSave: (config: AutoFunctionConfig) => void;
}

export function ConfigEditorModal({
  open,
  onClose,
  config,
  availableFunctions,
  onSave,
}: ConfigEditorModalProps) {
  const [form] = Form.useForm();

  useEffect(() => {
    if (open && config) {
      form.setFieldsValue({
        ...config,
        start_time: config.start_time
          ? dayjs(config.start_time, "HH:mm")
          : undefined,
        end_time: config.end_time ? dayjs(config.end_time, "HH:mm") : undefined,
      });
    } else {
      form.resetFields();
    }
  }, [open, config, form]);

  const handleFinish = (values: unknown) => {
    const newConfig: AutoFunctionConfig = {
      ...config,
      ...values,
      start_time: values.start_time ? values.start_time.format("HH:mm") : null,
      end_time: values.end_time ? values.end_time.format("HH:mm") : null,
    };
    onSave(newConfig);
    onClose();
  };

  return (
    <Modal
      title={config?.id ? `Configure ${config.name}` : "Add New Auto-Function"}
      open={open}
      onCancel={onClose}
      footer={null}
      width={500}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        initialValues={{
          interval_minutes: 60,
          enabled: true,
        }}
      >
        {!config?.id && (
          <Form.Item
            name="name"
            label="Function Name"
            rules={[{ required: true, message: "Please select a function" }]}
          >
            <Select placeholder="Select a function">
              {availableFunctions.map((func) => (
                <Option key={func.value} value={func.value}>
                  {func.label}
                </Option>
              ))}
            </Select>
          </Form.Item>
        )}

        <Form.Item
          name="interval_minutes"
          label="Interval (minutes)"
          rules={[{ required: true, message: "Please enter interval" }]}
        >
          <InputNumber min={1} style={{ width: "100%" }} />
        </Form.Item>

        <div style={{ display: "flex", gap: 16 }}>
          <Form.Item
            name="start_time"
            label="Start Time (Optional)"
            style={{ flex: 1 }}
          >
            <TimePicker format="HH:mm" style={{ width: "100%" }} />
          </Form.Item>
          <Form.Item
            name="end_time"
            label="End Time (Optional)"
            style={{ flex: 1 }}
          >
            <TimePicker format="HH:mm" style={{ width: "100%" }} />
          </Form.Item>
        </div>

        <Form.Item name="enabled" valuePropName="checked">
          <Checkbox>Enable this function</Checkbox>
        </Form.Item>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            {config?.id ? "Save" : "Create"}
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
