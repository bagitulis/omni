import { Modal, Tabs, Button } from "antd";
import { WholesaleTab } from "@/pages/inventory/components/WholesaleTab";
import { MpqTab } from "@/pages/inventory/components/MpqTab";
import { DeleteTab } from "@/pages/inventory/components/DeleteTab";

interface WholesaleMpqModalProps {
  open: boolean;
  onClose: () => void;
  items?: any[];
}

export function WholesaleMpqModal({ open, onClose }: WholesaleMpqModalProps) {
  const items = [
    {
      key: "wholesale",
      label: "Wholesale",
      children: <WholesaleTab />,
    },
    {
      key: "mpq",
      label: "MPQ",
      children: <MpqTab />,
    },
    {
      key: "delete",
      label: "Delete",
      children: <DeleteTab />,
    },
    {
      key: "settings",
      label: "Settings",
      children: <div>Coming soon</div>,
    },
  ];

  return (
    <Modal
      title="Bulk Pricing"
      open={open}
      onCancel={onClose}
      destroyOnClose
      width={900}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
        <Button key="update" type="primary" disabled>
          Update
        </Button>,
      ]}
      styles={{ body: { height: "600px", overflowY: "auto" } }}
      centered
    >
      <Tabs defaultActiveKey="wholesale" items={items} />
    </Modal>
  );
}
