import { create } from "zustand";

interface ModalState {
  isOpen: boolean;
  data?: any;
}

interface ModalsStore {
  modals: Record<string, ModalState>;
  openModal: (modalName: string, data?: any) => void;
  closeModal: (modalName: string) => void;
  getModalData: (modalName: string) => any;
  isModalOpen: (modalName: string) => boolean;
}

export const useModalsStore = create<ModalsStore>((set, get) => ({
  modals: {},
  openModal: (modalName, data) =>
    set((state) => ({
      modals: {
        ...state.modals,
        [modalName]: { isOpen: true, data },
      },
    })),
  closeModal: (modalName) =>
    set((state) => ({
      modals: {
        ...state.modals,
        [modalName]: { ...state.modals[modalName], isOpen: false },
      },
    })),
  getModalData: (modalName) => get().modals[modalName]?.data,
  isModalOpen: (modalName) => !!get().modals[modalName]?.isOpen,
}));
