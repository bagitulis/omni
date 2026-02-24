import { useEffect, useState } from "react";
import {
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  TimePicker,
} from "antd";
import type { Dayjs } from "dayjs";
import dayjs from "dayjs";
import type { AutoFunctionConfig } from "@/types/scriptMonitor";
import {
  getAvailableAutoFunctions,
  type AvailableAutoFunction,
} from "@/api/scriptMonitor";

type EditorMode = "create" | "edit";

interface FormValues {
  name: string;
  enabled: boolean;
  interval_minutes: number;
  start_time: Dayjs | null;
  end_time: Dayjs | null;
}

interface Props {
  open: boolean;
  mode: EditorMode;
  initialData?: AutoFunctionConfig;
  existingNames: string[];
  onClose: () => void;
  onCreate: (config: AutoFunctionConfig, onSuccess: () => void) => void;
  onUpdate: (
    payload: { name: string; config: Partial<AutoFunctionConfig> },
    onSuccess: () => void,
  ) => void;
}

function toTimeValue(value?: string): Dayjs | null {
  if (!value) {
    return null;
  }

  const normalized = value.length === 5 ? `${value}:00` : value;
  const parsed = dayjs(`1970-01-01T${normalized}`);
  return parsed.isValid() ? parsed : null;
}

function toTimeString(value: Dayjs | null): string | undefined {
  if (!value) {
    return undefined;
  }

  return value.format("HH:mm:ss");
}

export function ConfigEditorModal({
  open,
  mode,
  initialData,
  existingNames,
  onClose,
  onCreate,
  onUpdate,
}: Props) {
  const [form] = Form.useForm<FormValues>();
  const isEditMode = mode === "edit";
  const [availableFunctions, setAvailableFunctions] = useState<
    AvailableAutoFunction[]
  >([]);

  useEffect(() => {
    if (open && !isEditMode) {
      getAvailableAutoFunctions()
        .then(setAvailableFunctions)
        .catch(() => setAvailableFunctions([]));
    }
  }, [open, isEditMode]);

  // Filter out already-configured functions
  const selectOptions = availableFunctions
    .filter((fn) => !existingNames.includes(fn.name))
    .map((fn) => ({
      value: fn.name,
      label: `${fn.name} — ${fn.description}`,
    }));

  useEffect(() => {
    if (!open) {
      return;
    }

    if (isEditMode && initialData) {
      form.setFieldsValue({
        name: initialData.name,
        enabled: initialData.enabled,
        interval_minutes: initialData.interval_minutes,
        start_time: toTimeValue(initialData.start_time),
        end_time: toTimeValue(initialData.end_time),
      });
      return;
    }

    form.setFieldsValue({
      name: "",
      enabled: true,
      interval_minutes: 15,
      start_time: null,
      end_time: null,
    });
  }, [open, isEditMode, initialData, form]);

  const handleSubmit = async () => {
    const values = await form.validateFields();

    const payload: Partial<AutoFunctionConfig> = {
      enabled: values.enabled,
      interval_minutes: values.interval_minutes,
      start_time: toTimeString(values.start_time),
      end_time: toTimeString(values.end_time),
    };

    if (isEditMode && initialData) {
      onUpdate(
        {
          name: initialData.name,
          config: payload,
        },
        onClose,
      );
      return;
    }

    onCreate(
      {
        id: 0,
        name: values.name,
        enabled: values.enabled,
        interval_minutes: values.interval_minutes,
        start_time: payload.start_time,
        end_time: payload.end_time,
      },
      onClose,
    );
  };

  return (
    <Modal
      open={open}
      title={isEditMode ? "Edit Auto-Function" : "Add Auto-Function"}
      onCancel={onClose}
      destroyOnHidden
      footer={
        <Space>
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" onClick={handleSubmit}>
            {isEditMode ? "Save Changes" : "Create"}
          </Button>
        </Space>
      }
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          name: "",
          enabled: true,
          interval_minutes: 15,
          start_time: null,
          end_time: null,
        }}
      >
        <Form.Item
          label="Function Name"
          name="name"
          rules={
            isEditMode
              ? []
              : [{ required: true, message: "Please select a function" }]
          }
        >
          {isEditMode ? (
            <Input disabled />
          ) : (
            <Select
              placeholder="Select a function"
              options={selectOptions}
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? "")
                  .toLowerCase()
                  .includes(input.toLowerCase())
              }
            />
          )}
        </Form.Item>

        <Form.Item label="Enabled" name="enabled" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item
          label="Interval (minutes)"
          name="interval_minutes"
          rules={[{ required: true, message: "Interval is required" }]}
        >
          <InputNumber min={1} style={{ width: "100%" }} />
        </Form.Item>

        <Form.Item label="Start Time" name="start_time">
          <TimePicker format="HH:mm" style={{ width: "100%" }} allowClear />
        </Form.Item>

        <Form.Item label="End Time" name="end_time">
          <TimePicker format="HH:mm" style={{ width: "100%" }} allowClear />
        </Form.Item>
      </Form>
    </Modal>
  );
}
