import { useCallback, useEffect, useState } from "react";
import {
  Button,
  Card,
  Col,
  DatePicker,
  Input,
  Row,
  Select,
  Space,
  Spin,
  Statistic,
  Table,
  Tag,
  Typography,
} from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  DatabaseOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import {
  developerApi,
  type AuditLogEntry,
  type AuditLogParams,
  type SystemHealth,
} from "@/api/developer";
import type { ColumnsType } from "antd/es/table";
import type { Dayjs } from "dayjs";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

const ACTION_OPTIONS = [
  { label: "All Actions", value: "" },
  { label: "Login", value: "login" },
  { label: "Logout", value: "logout" },
  { label: "Create", value: "create" },
  { label: "Update", value: "update" },
  { label: "Delete", value: "delete" },
  { label: "Password Reset", value: "password_reset" },
];

export default function SystemTab() {
  // Health state
  const [health, setHealth] = useState<SystemHealth | null>(null);
  const [healthLoading, setHealthLoading] = useState(false);

  // Audit state
  const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
  const [auditLoading, setAuditLoading] = useState(false);
  const [auditTotal, setAuditTotal] = useState(0);
  const [filters, setFilters] = useState<AuditLogParams>({
    page: 1,
    limit: 10,
    action: "",
    user_id: "",
    date_from: "",
    date_to: "",
  });

  const fetchHealth = useCallback(async () => {
    setHealthLoading(true);
    try {
      const response = await developerApi.getSystemHealth();
      if (response.success && response.data) {
        setHealth(response.data);
      }
    } finally {
      setHealthLoading(false);
    }
  }, []);

  const fetchAuditLogs = useCallback(async (params: AuditLogParams) => {
    setAuditLoading(true);
    try {
      const cleanParams: AuditLogParams = {
        page: params.page,
        limit: params.limit,
      };
      if (params.action) cleanParams.action = params.action;
      if (params.user_id) cleanParams.user_id = params.user_id;
      if (params.date_from) cleanParams.date_from = params.date_from;
      if (params.date_to) cleanParams.date_to = params.date_to;

      const response = await developerApi.getAuditLogs(cleanParams);
      if (response.success && response.data) {
        setAuditLogs(response.data);
        setAuditTotal(response.data.length < (cleanParams.limit || 10)
          ? ((cleanParams.page || 1) - 1) * (cleanParams.limit || 10) + response.data.length
          : (cleanParams.page || 1) * (cleanParams.limit || 10) + 1);
      }
    } finally {
      setAuditLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchHealth();
  }, [fetchHealth]);

  useEffect(() => {
    void fetchAuditLogs(filters);
  }, [fetchAuditLogs, filters]);

  const handleDateChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    setFilters((prev) => ({
      ...prev,
      page: 1,
      date_from: dates?.[0]?.format("YYYY-MM-DD") || "",
      date_to: dates?.[1]?.format("YYYY-MM-DD") || "",
    }));
  };

  const handleActionChange = (value: string) => {
    setFilters((prev) => ({ ...prev, page: 1, action: value }));
  };

  const handleUserSearch = (value: string) => {
    setFilters((prev) => ({ ...prev, page: 1, user_id: value }));
  };

  const auditColumns: ColumnsType<AuditLogEntry> = [
    {
      title: "Timestamp",
      dataIndex: "created_at",
      key: "created_at",
      width: 180,
      render: (val: string) => {
        if (!val) return "-";
        const d = new Date(val);
        return d.toLocaleString();
      },
    },
    {
      title: "User",
      dataIndex: "user_id",
      key: "user_id",
      width: 200,
      ellipsis: true,
    },
    {
      title: "Action",
      dataIndex: "action",
      key: "action",
      width: 140,
      render: (action: string) => {
        const colorMap: Record<string, string> = {
          login: "green",
          logout: "default",
          create: "blue",
          update: "orange",
          delete: "red",
          password_reset: "purple",
        };
        return <Tag color={colorMap[action] || "default"}>{action}</Tag>;
      },
    },
    {
      title: "Details",
      dataIndex: "details",
      key: "details",
      ellipsis: true,
    },
    {
      title: "Tenant",
      dataIndex: "tenant_id",
      key: "tenant_id",
      width: 200,
      ellipsis: true,
    },
  ];

  return (
    <>
      {/* Health Section */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 16,
        }}
      >
        <div>
          <Title level={4} style={{ margin: 0 }}>
            System Health
          </Title>
          <Text type="secondary">Real-time system status overview</Text>
        </div>
        <Button
          icon={<ReloadOutlined />}
          loading={healthLoading}
          onClick={() => void fetchHealth()}
        >
          Refresh
        </Button>
      </div>

      <Spin spinning={healthLoading}>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col xs={12} md={6}>
            <Card>
              <Statistic
                title="DB Status"
                value={health?.db_status === "connected" ? "Connected" : "Disconnected"}
                prefix={<DatabaseOutlined />}
                valueStyle={{
                  color: health?.db_status === "connected" ? "#16a34a" : "#dc2626",
                  fontSize: 14,
                }}
              />
            </Card>
          </Col>
          <Col xs={12} md={6}>
            <Card>
              <Statistic
                title="Memory Usage"
                value={health?.memory_usage ?? 0}
                suffix="MB"
              />
            </Card>
          </Col>
          <Col xs={12} md={6}>
            <Card>
              <Statistic
                title="Goroutines"
                value={health?.goroutines ?? 0}
              />
            </Card>
          </Col>
          <Col xs={12} md={6}>
            <Card>
              <Statistic
                title="Uptime"
                value={health?.uptime ?? "-"}
                valueStyle={{ fontSize: 14 }}
              />
            </Card>
          </Col>
        </Row>

        {health?.components && health.components.length > 0 && (
          <Card title="Components" size="small" style={{ marginBottom: 24 }}>
            <Space wrap>
              {health.components.map((comp) => (
                <Tag
                  key={comp.name}
                  icon={
                    comp.status === "healthy" ? (
                      <CheckCircleOutlined />
                    ) : (
                      <CloseCircleOutlined />
                    )
                  }
                  color={comp.status === "healthy" ? "success" : "error"}
                >
                  {comp.name}
                </Tag>
              ))}
            </Space>
          </Card>
        )}
      </Spin>

      {/* Audit Trail Section */}
      <div style={{ marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>
          Audit Trail
        </Title>
        <Text type="secondary">System activity log</Text>
      </div>

      <Space wrap style={{ marginBottom: 16 }}>
        <RangePicker
          onChange={(dates) =>
            handleDateChange(dates as [Dayjs | null, Dayjs | null] | null)
          }
          style={{ width: 260 }}
        />
        <Select
          value={filters.action}
          onChange={handleActionChange}
          options={ACTION_OPTIONS}
          style={{ width: 160 }}
          placeholder="Filter by action"
        />
        <Input.Search
          placeholder="Search by user ID"
          onSearch={handleUserSearch}
          allowClear
          style={{ width: 220 }}
        />
      </Space>

      <Card>
        <Table<AuditLogEntry>
          rowKey="id"
          dataSource={auditLogs}
          columns={auditColumns}
          loading={auditLoading}
          pagination={{
            current: filters.page,
            pageSize: filters.limit,
            total: auditTotal,
            showSizeChanger: true,
            pageSizeOptions: ["10", "20", "50"],
            onChange: (page, pageSize) => {
              setFilters((prev) => ({ ...prev, page, limit: pageSize }));
            },
          }}
          scroll={{ x: 900 }}
          locale={{ emptyText: auditLoading ? "Loading..." : "No audit logs found" }}
        />
      </Card>
    </>
  );
}
