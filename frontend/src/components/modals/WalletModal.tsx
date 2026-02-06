import {
  Drawer,
  Button,
  Space,
  Divider,
  Statistic,
  Table,
  Modal,
  Form,
  InputNumber,
  Typography,
} from "antd";
import { DollarOutlined, PlusOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";

const { Text } = Typography;

export interface Transaction {
  id: string;
  type: "topup" | "purchase" | "refund";
  amount: number;
  description: string;
  timestamp: string;
}

interface WalletModalProps {
  open: boolean;
  onClose: () => void;
  balance: number;
  transactions?: Transaction[];
  onTopUp?: (amount: number) => Promise<void>;
}

const COLOR_MAP: Record<string, string> = {
  topup: "green",
  purchase: "red",
  refund: "blue",
};

export function WalletModal({
  open,
  onClose,
  balance,
  transactions = [],
  onTopUp,
}: WalletModalProps) {
  const [topupOpen, setTopupOpen] = useState(false);
  const [topupLoading, setTopupLoading] = useState(false);
  const [topupForm] = Form.useForm();

  const handleTopupSubmit = async () => {
    try {
      const { amount } = await topupForm.validateFields();
      if (!onTopUp) return;
      setTopupLoading(true);
      await onTopUp(amount);
      setTopupLoading(false);
      topupForm.resetFields();
      setTopupOpen(false);
    } catch (error) {
      setTopupLoading(false);
    }
  };

  const columns: ColumnsType<Transaction> = [
    {
      title: "Type",
      dataIndex: "type",
      key: "type",
      width: 80,
      render: (type) => <span style={{ color: COLOR_MAP[type] }}>{type}</span>,
    },
    {
      title: "Amount",
      dataIndex: "amount",
      key: "amount",
      width: 100,
      render: (amount, record) => {
        const isIncome = record.type === "topup" || record.type === "refund";
        return (
          <Text strong style={{ color: isIncome ? "green" : "red" }}>
            {isIncome ? "+" : "-"}
            {amount.toLocaleString()}
          </Text>
        );
      },
    },
    {
      title: "Description",
      dataIndex: "description",
      key: "description",
      render: (text) => <Text ellipsis>{text}</Text>,
    },
    {
      title: "Date",
      dataIndex: "timestamp",
      key: "timestamp",
      width: 130,
      render: (ts) => new Date(ts).toLocaleString(),
    },
  ];

  return (
    <>
      <Drawer
        title="Wallet"
        placement="right"
        width={500}
        onClose={onClose}
        open={open}
        footer={
          <Space style={{ display: "flex", justifyContent: "flex-end" }}>
            <Button onClick={onClose}>Close</Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setTopupOpen(true)}
            >
              Top Up
            </Button>
          </Space>
        }
      >
        <div style={{ marginBottom: 32 }}>
          <Statistic
            title="Wallet Balance"
            value={balance}
            prefix={<DollarOutlined />}
            precision={2}
            valueStyle={{ color: "#0369a1" }}
          />
        </div>
        <Divider />
        <Typography.Title level={5}>Transaction History</Typography.Title>
        <Table<Transaction>
          columns={columns}
          dataSource={transactions}
          rowKey="id"
          pagination={false}
          size="small"
          bordered
        />
      </Drawer>

      <Modal
        title="Top Up Wallet"
        open={topupOpen}
        onCancel={() => setTopupOpen(false)}
        footer={[
          <Button key="cancel" onClick={() => setTopupOpen(false)}>
            Cancel
          </Button>,
          <Button
            key="submit"
            type="primary"
            loading={topupLoading}
            onClick={handleTopupSubmit}
          >
            Confirm
          </Button>,
        ]}
      >
        <Form form={topupForm} layout="vertical">
          <Form.Item
            name="amount"
            label="Amount"
            rules={[
              { required: true, message: "Please enter amount" },
              { type: "number", min: 10, message: "Minimum top up is 10" },
            ]}
          >
            <InputNumber
              prefix="$"
              placeholder="Enter amount"
              min={10}
              step={10}
              precision={2}
              style={{ width: "100%" }}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
