import { Table, TableProps } from "antd";
import { useVirtualScroll } from "@/hooks/useVirtualScroll";

interface VirtualTableProps<T> extends TableProps<T> {
  enableVirtual?: boolean;
  containerHeight?: number;
  offsetBottom?: number;
}

/**
 * A wrapper around Ant Design Table that adds virtual scrolling support.
 * Automatically calculates the scroll height based on the viewport.
 */
export function VirtualTable<T extends object>({
  enableVirtual = true,
  containerHeight,
  offsetBottom,
  scroll,
  ...props
}: VirtualTableProps<T>) {
  const { tableProps } = useVirtualScroll({
    enabled: enableVirtual,
    containerHeight,
    offsetBottom,
  });

  const mergedScroll = {
    ...scroll,
    ...tableProps.scroll,
    // Preserve x scroll if it exists in props.scroll
    x: scroll?.x ?? tableProps.scroll.x,
  };

  return (
    <Table<T> {...props} virtual={tableProps.virtual} scroll={mergedScroll} />
  );
}
