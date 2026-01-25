import type { ModalsStateType } from "@/types/store";

interface ModalsStoreType {
  closeModal: (modalType: keyof ModalsStateType) => void;
}

export function useDashboardHandlers(
  walletParams: any,
  shippingParams: any,
  currentPriceOption: any,
  showShippingForm: any,
  showShippingFileList: any,
  modalsStore: ModalsStoreType
) {
  const handleCloseShippingModal = () => {
    modalsStore.closeModal("shipping");
    showShippingForm.value = false;
    showShippingFileList.value = false;
  };

  const handleSelectPriceOption = (option: string) => {
    currentPriceOption.value = option;
  };

  const handleUpdateWalletParams = (newParams: Record<string, any>) => {
    walletParams.value = { ...walletParams.value, ...newParams };
  };

  const handleUpdateShippingParams = (newParams: Record<string, any>) => {
    shippingParams.value = { ...shippingParams.value, ...newParams };
  };

  return {
    handleCloseShippingModal,
    handleSelectPriceOption,
    handleUpdateWalletParams,
    handleUpdateShippingParams,
  };
}
