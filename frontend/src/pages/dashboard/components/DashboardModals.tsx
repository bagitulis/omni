import { useModalsStore } from "@/stores/modalsStore";
import {
  PriceModal,
  ExportOrdersModal,
  WalletModal,
  TokenModal,
  ChangePasswordModal,
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
      <ChangePasswordModal
        open={isModalOpen("changePassword")}
        onClose={() => closeModal("changePassword")}
      />
      {/* DashboardShippingModal will be added in Task 2.6 */}
    </>
  );
}
