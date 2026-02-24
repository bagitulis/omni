import { useEffect } from "react";
import {
  Button,
  Divider,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  Tag,
  Typography,
} from "antd";
import { useUpdateRouteConfig } from "@/hooks/useRouteConfig";
import type { RouteConfig } from "@/types/routeConfig";

interface RouteConfigModalProps {
  open: boolean;
  onClose: () => void;
  route: RouteConfig | null;
}

type RouteConfigFormValues = Pick<
  RouteConfig,
  | "enabled"
  | "description"
  | "caching_enabled"
  | "cache_ttl"
  | "cache_strategy"
  | "queue_enabled"
  | "queue_max_size"
  | "queue_priority"
  | "rate_limit_enabled"
  | "rate_limit_window"
  | "rate_limit_max"
  | "max_concurrent"
  | "timeout"
  | "min_interval_ms"
  | "retry_enabled"
  | "max_retries"
  | "retry_delay_ms"
>;

export function RouteConfigModal({
  open,
  onClose,
  route,
}: RouteConfigModalProps) {
  const [form] = Form.useForm<RouteConfigFormValues>();
  const updateRouteConfig = useUpdateRouteConfig();

  useEffect(() => {
    if (!route) {
      form.resetFields();
      return;
    }

    form.setFieldsValue({
      enabled: route.enabled,
      description: route.description,
      caching_enabled: route.caching_enabled,
      cache_ttl: route.cache_ttl,
      cache_strategy: route.cache_strategy,
      queue_enabled: route.queue_enabled,
      queue_max_size: route.queue_max_size,
      queue_priority: route.queue_priority,
      rate_limit_enabled: route.rate_limit_enabled,
      rate_limit_window: route.rate_limit_window,
      rate_limit_max: route.rate_limit_max,
      max_concurrent: route.max_concurrent,
      timeout: route.timeout,
      min_interval_ms: route.min_interval_ms,
      retry_enabled: route.retry_enabled,
      max_retries: route.max_retries,
      retry_delay_ms: route.retry_delay_ms,
    });
  }, [form, route]);

  const handleSave = async () => {
    if (!route) {
      return;
    }

    const values = await form.validateFields().catch(() => null);
    if (!values) {
      return;
    }

    updateRouteConfig.mutate(
      { id: route.id, data: values },
      {
        onSuccess: () => {
          onClose();
        },
      },
    );
  };

  return (
    <Modal
      title={
        <Space size={8}>
          <Typography.Text strong>Edit Route Config</Typography.Text>
          {route && (
            <>
              <Tag color="blue">{route.route_method}</Tag>
              <Typography.Text code>{route.route_path}</Typography.Text>
            </>
          )}
        </Space>
      }
      open={open}
      onCancel={onClose}
      width={720}
      footer={
        <Space>
          <Button onClick={onClose}>Cancel</Button>
          <Button
            type="primary"
            loading={updateRouteConfig.isPending}
            onClick={handleSave}
          >
            Save
          </Button>
        </Space>
      }
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        disabled={updateRouteConfig.isPending}
      >
        <Typography.Title level={5}>General</Typography.Title>
        <Form.Item name="enabled" valuePropName="checked" label="Enabled">
          <Switch />
        </Form.Item>
        <Form.Item name="description" label="Description">
          <Input placeholder="Describe this route behavior" />
        </Form.Item>

        <Divider />
        <Typography.Title level={5}>Caching</Typography.Title>
        <Form.Item
          name="caching_enabled"
          valuePropName="checked"
          label="Caching Enabled"
        >
          <Switch />
        </Form.Item>
        <Form.Item name="cache_ttl" label="Cache TTL">
          <InputNumber min={0} style={{ width: "100%" }} addonAfter="s" />
        </Form.Item>
        <Form.Item name="cache_strategy" label="Cache Strategy">
          <Select
            options={[
              { value: "normal", label: "Normal" },
              { value: "aggressive", label: "Aggressive" },
              { value: "passive", label: "Passive" },
            ]}
          />
        </Form.Item>

        <Divider />
        <Typography.Title level={5}>Queue</Typography.Title>
        <Form.Item
          name="queue_enabled"
          valuePropName="checked"
          label="Queue Enabled"
        >
          <Switch />
        </Form.Item>
        <Form.Item name="queue_max_size" label="Queue Max Size">
          <InputNumber min={0} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item name="queue_priority" label="Queue Priority">
          <InputNumber min={0} style={{ width: "100%" }} />
        </Form.Item>

        <Divider />
        <Typography.Title level={5}>Rate Limiting</Typography.Title>
        <Form.Item
          name="rate_limit_enabled"
          valuePropName="checked"
          label="Rate Limit Enabled"
        >
          <Switch />
        </Form.Item>
        <Form.Item name="rate_limit_window" label="Rate Limit Window">
          <InputNumber min={1} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item name="rate_limit_max" label="Rate Limit Max">
          <InputNumber min={1} style={{ width: "100%" }} />
        </Form.Item>

        <Divider />
        <Typography.Title level={5}>Performance</Typography.Title>
        <Form.Item name="max_concurrent" label="Max Concurrent">
          <InputNumber min={1} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item name="timeout" label="Timeout">
          <InputNumber min={1} style={{ width: "100%" }} addonAfter="ms" />
        </Form.Item>
        <Form.Item name="min_interval_ms" label="Minimum Interval">
          <InputNumber min={0} style={{ width: "100%" }} addonAfter="ms" />
        </Form.Item>

        <Divider />
        <Typography.Title level={5}>Retry</Typography.Title>
        <Form.Item
          name="retry_enabled"
          valuePropName="checked"
          label="Retry Enabled"
        >
          <Switch />
        </Form.Item>
        <Form.Item name="max_retries" label="Max Retries">
          <InputNumber min={0} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item name="retry_delay_ms" label="Retry Delay">
          <InputNumber min={0} style={{ width: "100%" }} addonAfter="ms" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
