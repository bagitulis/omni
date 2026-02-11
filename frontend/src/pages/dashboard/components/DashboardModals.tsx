import { useModalsStore } from "@/stores/modalsStore";
import {
  PriceModal,
  ExportOrdersModal,
  WalletModal,
  TokenModal,
  ChangePasswordModal,
  DashboardShippingModal,
} from "@/components/modals";

export function DashboardModals() {
  const { isModalOpen, closeModal } = useModalsStore();

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
      />
      <TokenModal
        open={isModalOpen("token")}
        onClose={() => closeModal("token")}
      />
      <ChangePasswordModal
        open={isModalOpen("changePassword")}
        onClose={() => closeModal("changePassword")}
      />
      <DashboardShippingModal
        open={isModalOpen("shipping")}
        onClose={() => closeModal("shipping")}
      />
    </>
  );
}
