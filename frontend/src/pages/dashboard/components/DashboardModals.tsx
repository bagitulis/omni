import { useModalsStore } from "@/stores/modalsStore";
import {
  PriceModal,
  ExportOrdersModal,
  WalletModal,
  TokenModal,
  ChangePasswordModal,
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
      {/* DashboardShippingModal will be added in Task 2.6 */}
    </>
  );
}
