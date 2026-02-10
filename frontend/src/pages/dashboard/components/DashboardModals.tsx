import { useModalsStore } from "@/stores/modalsStore";
import {
  PriceModal,
  ExportOrdersModal,
  WalletModal,
  TokenModal,
} from "@/components/modals";

export function DashboardModals() {
  const { isModalOpen, closeModal, getModalData } = useModalsStore();

  const handleTokenOperation = (operation: string) => {
    console.log("Token operation:", operation);
    // TODO: Implement token operation logic
  };

  return (
    <>
      <PriceModal
        open={isModalOpen("price")}
        onCancel={() => closeModal("price")}
        selectedProducts={[]}
      />
      <ExportOrdersModal
        open={isModalOpen("exportOrders")}
        onClose={() => closeModal("exportOrders")}
      />
      <WalletModal
        open={isModalOpen("wallet")}
        onClose={() => closeModal("wallet")}
        balance={getModalData("wallet")?.balance || 0}
      />
      <TokenModal
        open={isModalOpen("token")}
        onClose={() => closeModal("token")}
        platform={getModalData("token")?.platform}
        onTokenOperation={handleTokenOperation}
      />
    </>
  );
}
