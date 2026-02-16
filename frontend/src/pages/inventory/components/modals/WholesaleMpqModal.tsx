import { Modal, Tabs, Button } from "antd";
import { WholesaleTab } from "@/pages/inventory/components/WholesaleTab";
import { MpqTab } from "@/pages/inventory/components/MpqTab";

interface WholesaleMpqModalProps {
  open: boolean;
  onClose: () => void;
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
      ]}
      styles={{ body: { height: "600px", overflowY: "auto" } }}
      centered
    >
      <Tabs defaultActiveKey="wholesale" items={items} />
    </Modal>
  );
}
