import { defineStore } from "pinia";
import { ref, Ref } from "vue";
import type { ModalsStateType } from "../types/store";

export const useModalsStore = defineStore("modals", () => {
  const modals: Ref<ModalsStateType> = ref({
    token: { visible: false, platform: null },
    wallet: { visible: false },
    shipping: { visible: false },
    price: { visible: false, platform: null },
    exportOrders: { visible: false, platform: null },
    debug: { visible: false },
  });

  function showModal(modalType: keyof ModalsStateType, platform: string | null = null): void {
    modals.value[modalType] = { visible: true, platform };
  }

  function closeModal(modalType: keyof ModalsStateType): void {
    modals.value[modalType] = { visible: false, platform: null };
  }

  return {
    modals,
    showModal,
    closeModal,
  };
});
