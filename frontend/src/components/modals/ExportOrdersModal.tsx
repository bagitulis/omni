import React, { useState } from "react";
import {
  Modal,
  Form,
  DatePicker,
  Select,
  Radio,
  Button,
  Alert,
  Space,
  Typography,
} from "antd";
import dayjs, { Dayjs } from "dayjs";
import { useExportOrders } from "@/hooks/useExports";
import { ExportOrdersParams } from "@/api/exports";

const { RangePicker } = DatePicker;
const { Text } = Typography;

interface ExportOrdersModalProps {
  open: boolean;
  onClose: () => void;
}

const PLATFORM_OPTIONS = [
  { label: "Shopee", value: "shopee" },
  { label: "Lazada", value: "lazada" },
  { label: "TikTok", value: "tiktok" },
];

const STATUS_OPTIONS = [
  { label: "Pending", value: "pending" },
  { label: "Processing", value: "processing" },
  { label: "Shipped", value: "shipped" },
  { label: "Completed", value: "completed" },
  { label: "Cancelled", value: "cancelled" },
];

export const ExportOrdersModal: React.FC<ExportOrdersModalProps> = ({
  open,
  onClose,
}) => {
  const [form] = Form.useForm();
  const { mutate: exportOrders, isPending } = useExportOrders();

  // Default date range: last 7 days
  const [dates, setDates] = useState<[Dayjs, Dayjs]>([
    dayjs().subtract(7, "day"),
    dayjs(),
  ]);

  const handleExport = async () => {
    try {
      const values = await form.validateFields();

      const params: ExportOrdersParams = {
        platform: values.platform?.length ? values.platform : undefined,
        status: values.status?.length ? values.status : undefined,
        date_from: dates[0].format("YYYY-MM-DD"),
        date_to: dates[1].format("YYYY-MM-DD"),
        format: values.format,
      };

      exportOrders(params, {
        onSuccess: () => {
          onClose();
        },
      });
    } catch {
      // Form validation error, do nothing
    }
  };

  const handleClose = () => {
    form.resetFields();
    setDates([dayjs().subtract(7, "day"), dayjs()]);
    onClose();
  };

  return (
    <Modal
      title="Export Orders"
      open={open}
      onCancel={handleClose}
      destroyOnHidden
      footer={[
        <Button key="cancel" onClick={handleClose} disabled={isPending}>
          Cancel
        </Button>,
        <Button
          key="export"
          type="primary"
          loading={isPending}
          onClick={handleExport}
          disabled={!dates || dates.length !== 2}
        >
          Export
        </Button>,
      ]}
      width={500}
    >
      <Space direction="vertical" size="large" style={{ width: "100%" }}>
        <Alert
          message="Export orders to CSV or Excel file"
          description="Select date range, platforms, and status filters below."
          type="info"
          showIcon
        />

        <Form
          form={form}
          layout="vertical"
          initialValues={{
            format: "excel",
            platform: [], // Empty means All
            status: [], // Empty means All
          }}
        >
          <Form.Item
            label="Date Range"
            required
            tooltip="Select the date range for the export"
          >
            <RangePicker
              value={dates}
              onChange={(values) => {
                if (values && values[0] && values[1]) {
                  setDates([values[0], values[1]]);
                }
              }}
              style={{ width: "100%" }}
              allowClear={false}
              disabled={isPending}
            />
          </Form.Item>

          <Form.Item label="Platform" name="platform">
            <Select
              mode="multiple"
              placeholder="All Platforms"
              options={PLATFORM_OPTIONS}
              allowClear
              disabled={isPending}
              maxTagCount="responsive"
            />
          </Form.Item>

          <Form.Item label="Order Status" name="status">
            <Select
              mode="multiple"
              placeholder="All Statuses"
              options={STATUS_OPTIONS}
              allowClear
              disabled={isPending}
              maxTagCount="responsive"
            />
          </Form.Item>

          <Form.Item
            label="Export Format"
            name="format"
            rules={[{ required: true, message: "Please select a format" }]}
          >
            <Radio.Group disabled={isPending}>
              <Radio value="csv">CSV</Radio>
              <Radio value="excel">Excel (.xlsx)</Radio>
            </Radio.Group>
          </Form.Item>
        </Form>

        {isPending && (
          <div style={{ textAlign: "center", padding: "10px" }}>
            <Text type="secondary">Generating export file, please wait...</Text>
          </div>
        )}
      </Space>
    </Modal>
  );
};

export default ExportOrdersModal;
