import { Empty, Table } from "antd";
import type { OmniExtension } from "@/api/extensions";
import { getExtensionTableColumns } from "./ExtensionTableColumns";

export interface ExtensionTableProps {
  extensions: OmniExtension[];
  loading: boolean;
  onUnpair: (extensionId: string) => void;
  unpairing: boolean;
}

export function ExtensionTable({
  extensions,
  loading,
  onUnpair,
  unpairing,
}: ExtensionTableProps) {
  return (
    <Table<OmniExtension>
      rowKey="extension_id"
      columns={getExtensionTableColumns({ onUnpair, unpairing })}
      dataSource={extensions}
      loading={loading}
      pagination={false}
      locale={{
        emptyText: (
          <Empty description="No extensions paired yet. Use 'Pair New Extension' to start." />
        ),
      }}
    />
  );
}

export default ExtensionTable;
