import {
  Modal,
  Form,
  Input,
  Select,
  Button,
  Checkbox,
  InputNumber,
  Card,
  Row,
  Col,
} from "antd";
import { useEffect } from "react";

const { Option } = Select;

export interface RouteConfig {
  id?: string;
  route_path: string;
  route_method: string;
  category?: string;
  description?: string;
  caching_enabled: boolean;
  cache_ttl?: number;
  cache_strategy?: string;
  queue_enabled: boolean;
  queue_max_size?: number;
  max_concurrent?: number;
  rate_limit_enabled: boolean;
  rate_limit_window?: number;
  rate_limit_max?: number;
}

interface RouteManagementModalProps {
  open: boolean;
  onClose: () => void;
  route: RouteConfig | null;
  onSave: (route: RouteConfig) => void;
}

export function RouteManagementModal({
  open,
  onClose,
  route,
  onSave,
}: RouteManagementModalProps) {
  const [form] = Form.useForm();
  const isEdit = !!route?.id;

  useEffect(() => {
    if (open) {
      if (route) {
        form.setFieldsValue(route);
      } else {
        form.resetFields();
        form.setFieldsValue({
          route_method: "GET",
          caching_enabled: true,
          queue_enabled: true,
          rate_limit_enabled: false,
          cache_ttl: 0,
          cache_strategy: "standard",
          queue_max_size: 100,
          max_concurrent: 5,
          rate_limit_window: 60,
          rate_limit_max: 100,
        });
      }
    }
  }, [open, route, form]);

  const handleFinish = (values: Record<string, unknown>) => {
    const merged: RouteConfig = {
      ...(route ?? ({} as Partial<RouteConfig>)),
      ...(values as Partial<RouteConfig>),
    } as RouteConfig;
    onSave(merged);
    onClose();
  };

  return (
    <Modal
      title={isEdit ? "Edit Route" : "Add New Route"}
      open={open}
      onCancel={onClose}
      footer={null}
      width={700}
    >
      <Form form={form} layout="vertical" onFinish={handleFinish}>
        {/* Basic Info */}
        <Card size="small" style={{ marginBottom: 16 }}>
          <Form.Item
            name="route_path"
            label="Route Path"
            rules={[{ required: true, message: "Please enter route path" }]}
          >
            <Input placeholder="/api/orders" disabled={isEdit} />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="route_method" label="Method">
                <Select>
                  <Option value="GET">GET</Option>
                  <Option value="POST">POST</Option>
                  <Option value="PUT">PUT</Option>
                  <Option value="DELETE">DELETE</Option>
                  <Option value="PATCH">PATCH</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="category" label="Category">
                <Input placeholder="orders, products, etc" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="Description">
            <Input placeholder="Route description" />
          </Form.Item>
        </Card>

        {/* Caching */}
        <Card
          size="small"
          title={
            <Form.Item name="caching_enabled" valuePropName="checked" noStyle>
              <Checkbox>Enable Caching</Checkbox>
            </Form.Item>
          }
          style={{ marginBottom: 16 }}
        >
          <Form.Item
            noStyle
            shouldUpdate={(prev, curr) =>
              prev.caching_enabled !== curr.caching_enabled
            }
          >
            {({ getFieldValue }) =>
              getFieldValue("caching_enabled") && (
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item label="TTL (seconds)" name="cache_ttl">
                      <InputNumber
                        min={0}
                        max={3600}
                        style={{ width: "100%" }}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="Strategy" name="cache_strategy">
                      <Select>
                        <Option value="standard">Standard</Option>
                        <Option value="aggressive">Aggressive</Option>
                        <Option value="minimal">Minimal</Option>
                      </Select>
                    </Form.Item>
                  </Col>
                </Row>
              )
            }
          </Form.Item>
        </Card>

        {/* Queue */}
        <Card
          size="small"
          title={
            <Form.Item name="queue_enabled" valuePropName="checked" noStyle>
              <Checkbox>Enable Queue</Checkbox>
            </Form.Item>
          }
          style={{ marginBottom: 16 }}
        >
          <Form.Item
            noStyle
            shouldUpdate={(prev, curr) =>
              prev.queue_enabled !== curr.queue_enabled
            }
          >
            {({ getFieldValue }) =>
              getFieldValue("queue_enabled") && (
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item label="Max Size" name="queue_max_size">
                      <InputNumber
                        min={1}
                        max={1000}
                        style={{ width: "100%" }}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="Max Concurrent" name="max_concurrent">
                      <InputNumber min={1} max={50} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                </Row>
              )
            }
          </Form.Item>
        </Card>

        {/* Rate Limit */}
        <Card
          size="small"
          title={
            <Form.Item
              name="rate_limit_enabled"
              valuePropName="checked"
              noStyle
            >
              <Checkbox>Enable Rate Limiting</Checkbox>
            </Form.Item>
          }
          style={{ marginBottom: 16 }}
        >
          <Form.Item
            noStyle
            shouldUpdate={(prev, curr) =>
              prev.rate_limit_enabled !== curr.rate_limit_enabled
            }
          >
            {({ getFieldValue }) =>
              getFieldValue("rate_limit_enabled") && (
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item label="Window (sec)" name="rate_limit_window">
                      <InputNumber
                        min={1}
                        max={3600}
                        style={{ width: "100%" }}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="Max Requests" name="rate_limit_max">
                      <InputNumber
                        min={1}
                        max={10000}
                        style={{ width: "100%" }}
                      />
                    </Form.Item>
                  </Col>
                </Row>
              )
            }
          </Form.Item>
        </Card>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            {isEdit ? "Update" : "Create"}
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
