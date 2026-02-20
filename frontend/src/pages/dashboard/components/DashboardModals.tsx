import { useModalsStore } from "@/stores/modalsStore";
import {
  ExportOrdersModal,
  TokenModal,
  ChangePasswordModal,
} from "@/components/modals";

export function DashboardModals() {
  const { isModalOpen, closeModal } = useModalsStore();

  return (
    <>
      <ExportOrdersModal
        open={isModalOpen("exportOrders")}
        onClose={() => closeModal("exportOrders")}
      />
      <TokenModal
        open={isModalOpen("token")}
        onClose={() => closeModal("token")}
      />
      <ChangePasswordModal
        open={isModalOpen("changePassword")}
        onClose={() => closeModal("changePassword")}
      />
    </>
  );
}
