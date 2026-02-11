import {
  Modal,
  Button,
  Space,
  Statistic,
  Table,
  Select,
  Row,
  Col,
  Spin,
  Alert,
  Typography,
} from "antd";
import {
  DollarOutlined,
  DownloadOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";
import {
  useWalletBalance,
  useWalletTransactions,
  useExportWalletToSheets,
  type WalletTransaction,
} from "@/hooks/useWallet";

const { Text } = Typography;

interface WalletModalProps {
  open: boolean;
  onClose: () => void;
}

export function WalletModal({ open, onClose }: WalletModalProps) {
  const currentYear = new Date().getFullYear();
  const currentMonth = new Date().getMonth() + 1;

  const [selectedMonth, setSelectedMonth] = useState<number>(currentMonth);
  const [selectedYear, setSelectedYear] = useState<number>(currentYear);

  // Fetch wallet balance
  const {
    data: walletData,
    isLoading: balanceLoading,
    error: balanceError,
    refetch: refetchBalance,
  } = useWalletBalance();

  // Fetch wallet transactions with filters
  const {
    data: transactionsData,
    isLoading: transactionsLoading,
    error: transactionsError,
    refetch: refetchTransactions,
  } = useWalletTransactions({
    month: selectedMonth,
    year: selectedYear,
  });

  // Export mutation
  const { mutate: exportToSheets, isPending: exporting } =
    useExportWalletToSheets();

  const handleExport = () => {
    exportToSheets({ month: selectedMonth, year: selectedYear });
  };

  const handleRefresh = () => {
    refetchBalance();
    refetchTransactions();
  };

  // Table columns
  const columns: ColumnsType<WalletTransaction> = [
    {
      title: "Date",
      dataIndex: "created_at",
      key: "created_at",
      width: 130,
      render: (date) => new Date(date).toLocaleDateString(),
    },
    {
      title: "Type",
      dataIndex: "type",
      key: "type",
      width: 100,
    },
    {
      title: "Amount",
      dataIndex: "amount",
      key: "amount",
      width: 120,
      align: "right",
      render: (amount, record) => {
        const isIncome = record.type === "income" || record.type === "refund";
        return (
          <Text strong style={{ color: isIncome ? "#16a34a" : "#dc2626" }}>
            {isIncome ? "+" : "-"}
            {walletData?.currency} {Math.abs(amount).toLocaleString()}
          </Text>
        );
      },
    },
    {
      title: "Description",
      dataIndex: "description",
      key: "description",
      ellipsis: true,
    },
  ];

  // Generate month options (last 12 months)
  const monthOptions = Array.from({ length: 12 }, (_, i) => ({
    label: new Date(0, i).toLocaleString("en", { month: "long" }),
    value: i + 1,
  }));

  // Generate year options (current year and 2 previous years)
  const yearOptions = Array.from({ length: 3 }, (_, i) => ({
    label: String(currentYear - i),
    value: currentYear - i,
  }));

  return (
    <Modal
      title="Wallet"
      open={open}
      onCancel={onClose}
      width={800}
      footer={
        <Space>
          <Button onClick={onClose}>Close</Button>
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            loading={exporting}
            onClick={handleExport}
          >
            Export to Sheets
          </Button>
        </Space>
      }
    >
      {/* Balance Summary */}
      {balanceError ? (
        <Alert
          message="Failed to load wallet balance"
          description={(balanceError as Error).message}
          type="error"
          showIcon
          style={{ marginBottom: 24 }}
        />
      ) : (
        <Spin spinning={balanceLoading}>
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={8}>
              <Statistic
                title="Total Balance"
                value={walletData?.total_balance || 0}
                prefix={<DollarOutlined />}
                suffix={walletData?.currency || "IDR"}
                valueStyle={{ color: "#0369a1", fontSize: 20 }}
              />
            </Col>
            <Col span={8}>
              <Statistic
                title="Available"
                value={walletData?.available_balance || 0}
                suffix={walletData?.currency || "IDR"}
                valueStyle={{ color: "#16a34a", fontSize: 20 }}
              />
            </Col>
            <Col span={8}>
              <Statistic
                title="Pending"
                value={walletData?.pending_balance || 0}
                suffix={walletData?.currency || "IDR"}
                valueStyle={{ color: "#d97706", fontSize: 20 }}
              />
            </Col>
          </Row>
        </Spin>
      )}

      {/* Filters */}
      <Space style={{ marginBottom: 16 }}>
        <Select
          value={selectedMonth}
          onChange={setSelectedMonth}
          options={monthOptions}
          style={{ width: 140 }}
        />
        <Select
          value={selectedYear}
          onChange={setSelectedYear}
          options={yearOptions}
          style={{ width: 100 }}
        />
        <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
          Refresh
        </Button>
      </Space>

      {/* Transaction Table */}
      {transactionsError ? (
        <Alert
          message="Failed to load transactions"
          description={(transactionsError as Error).message}
          type="error"
          showIcon
        />
      ) : (
        <Table<WalletTransaction>
          columns={columns}
          dataSource={transactionsData?.transactions || []}
          rowKey="transaction_id"
          loading={transactionsLoading}
          pagination={{
            pageSize: 10,
            showTotal: (total) => `Total ${total} transactions`,
            showSizeChanger: false,
          }}
          size="small"
          bordered
        />
      )}
    </Modal>
  );
}
