import { Grid, Pagination, Typography } from "antd";

interface Props {
  current: number;
  pageSize: number;
  total: number;
  onChange: (page: number, pageSize: number) => void;
  onShowSizeChange: (current: number, size: number) => void;
}

export function InventoryPagination({
  current,
  pageSize,
  total,
  onChange,
  onShowSizeChange,
}: Props) {
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  const start = total === 0 ? 0 : (current - 1) * pageSize + 1;
  const end = total === 0 ? 0 : Math.min(current * pageSize, total);

  return (
    <div
      style={{
        marginTop: 16,
        display: "flex",
        justifyContent: isMobile ? "flex-start" : "space-between",
        alignItems: isMobile ? "stretch" : "center",
        flexDirection: isMobile ? "column" : "row",
        gap: 12,
        flexWrap: "wrap",
      }}
    >
      <Typography.Text>
        Showing {start}-{end} of {total} items
      </Typography.Text>

      <Pagination
        current={current}
        pageSize={pageSize}
        total={total}
        onChange={onChange}
        onShowSizeChange={onShowSizeChange}
        showSizeChanger
        pageSizeOptions={[10, 25, 50, 100]}
        size={isMobile ? "small" : "default"}
      />
    </div>
  );
}
