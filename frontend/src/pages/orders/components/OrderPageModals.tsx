import { OrderDetailModal } from "@/components/modals/OrderDetailModal";
import {
  OrderShipModal,
  ShipConfirmPayload,
} from "@/components/modals/OrderShipModal";
import {
  OrderCancelModal,
  CancelFormValues,
} from "@/components/modals/OrderCancelModal";
import { OrderDetail, Order } from "@/types/order";

interface OrderPageModalsProps {
  isDetailModalOpen: boolean;
  isShipModalOpen: boolean;
  isCancelModalOpen: boolean;
  selectedOrder: OrderDetail | Order | null;
  onDetailClose: () => void;
  onShipClose: () => void;
  onCancelClose: () => void;
  onShipConfirm: (payload: ShipConfirmPayload) => Promise<void>;
  onCancelConfirm: (orderSn: string, values: CancelFormValues) => Promise<void>;
  isSingleShipping: boolean;
  isCancelling: boolean;
}

export function OrderPageModals({
  isDetailModalOpen,
  isShipModalOpen,
  isCancelModalOpen,
  selectedOrder,
  onDetailClose,
  onShipClose,
  onCancelClose,
  onShipConfirm,
  onCancelConfirm,
  isSingleShipping,
  isCancelling,
}: OrderPageModalsProps) {
  return (
    <>
      <OrderDetailModal
        open={isDetailModalOpen}
        order={selectedOrder as OrderDetail}
        onClose={onDetailClose}
      />

      <OrderShipModal
        open={isShipModalOpen}
        order={selectedOrder as Order}
        onClose={onShipClose}
        onConfirm={onShipConfirm}
        loading={isSingleShipping}
      />

      <OrderCancelModal
        open={isCancelModalOpen}
        order={selectedOrder as Order}
        onClose={onCancelClose}
        onConfirm={onCancelConfirm}
        loading={isCancelling}
      />
    </>
  );
}
