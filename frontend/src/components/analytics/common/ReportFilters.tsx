import { Row, Col, Select, Segmented } from "antd";
import { ReconciliationOutlined, DollarOutlined } from "@ant-design/icons";
import type { ReportTab } from "@/types/analytics";

interface ReportFiltersProps {
  month: number;
  year: number;
  onMonthChange: (m: number) => void;
  onYearChange: (y: number) => void;
  tab: ReportTab;
  onTabChange: (t: ReportTab) => void;
}

const CURRENT_YEAR = new Date().getFullYear();

const monthOptions = Array.from({ length: 12 }, (_, i) => ({
  value: i + 1,
  label: new Date(2000, i, 1).toLocaleDateString("en-US", { month: "long" }),
}));

const yearOptions = Array.from({ length: 5 }, (_, i) => ({
  value: CURRENT_YEAR - i,
  label: String(CURRENT_YEAR - i),
}));

const tabOptions = [
  {
    value: "reconciliation" as ReportTab,
    label: (
      <span>
        <ReconciliationOutlined /> Reconciliation
      </span>
    ),
  },
  {
    value: "shipping_fee" as ReportTab,
    label: (
      <span>
        <DollarOutlined /> Shipping Fee
      </span>
    ),
  },
];

/**
 * Filter controls for the report page:
 * Month/year selection and tab switching (reconciliation / shipping fee).
 */
export function ReportFilters({
  month,
  year,
  onMonthChange,
  onYearChange,
  tab,
  onTabChange,
}: ReportFiltersProps) {
  return (
    <Row gutter={[16, 16]} align="middle" wrap>
      <Col>
        <Select
          value={month}
          onChange={onMonthChange}
          options={monthOptions}
          style={{ width: 140 }}
          size="small"
        />
      </Col>
      <Col>
        <Select
          value={year}
          onChange={onYearChange}
          options={yearOptions}
          style={{ width: 100 }}
          size="small"
        />
      </Col>
      <Col flex="auto">
        <Segmented
          value={tab}
          onChange={(val) => onTabChange(val as ReportTab)}
          options={tabOptions}
          size="small"
        />
      </Col>
    </Row>
  );
}
