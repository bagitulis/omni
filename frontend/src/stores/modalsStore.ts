import { create } from "zustand";

interface ModalState {
  isOpen: boolean;
  data?: unknown;
}

interface ModalsStore {
  modals: Record<string, ModalState>;
  openModal: (modalName: string, data?: unknown) => void;
  closeModal: (modalName: string) => void;
  getModalData: (modalName: string) => unknown;
  isModalOpen: (modalName: string) => boolean;
}

export const useModalsStore = create<ModalsStore>((set, get) => ({
  modals: {},
  openModal: (modalName, data) =>
    set((state) => {
      // Close all other modals first (single modal enforcement)
      const closedModals = Object.keys(state.modals).reduce(
        (acc, key) => ({
          ...acc,
          [key]: { ...state.modals[key], isOpen: false },
        }),
        {} as Record<string, ModalState>,
      );

      return {
        modals: {
          ...closedModals,
          [modalName]: { isOpen: true, data },
        },
      };
    }),
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
