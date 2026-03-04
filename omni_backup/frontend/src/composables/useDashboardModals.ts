import { ref, computed } from "vue";
import { useAppStore } from "../store/app";

export function useDashboardModals() {
  // const modalsStore = useModalsStore();
  const appStore = useAppStore();

  const currentPriceOption = ref<any>(null);
  const showShippingForm = ref<boolean>(false);
  const showShippingFileList = ref<boolean>(false);

  const shippingFiles = computed(() => appStore.shippingFiles);

  const resetShippingModal = () => {
    showShippingForm.value = false;
    showShippingFileList.value = false;
  };

  return {
    currentPriceOption,
    showShippingForm,
    showShippingFileList,
    shippingFiles,
    resetShippingModal,
  };
}
