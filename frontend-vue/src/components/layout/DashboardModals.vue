<template>
  <TokenModal
    :visible="modals.token.visible"
    @update:visible="modals.token.visible = $event"
    :platform="modals.token.platform"
    @close="closeModal('token')"
    @handle-token-operation="handleTokenOperation"
  />
  <WalletModal
    :visible="modals.wallet.visible"
    @update:visible="modals.wallet.visible = $event"
    :wallet-params="walletParams"
    @close="closeModal('wallet')"
    @update-wallet-params="handleUpdateWalletParams"
    @export-wallet-to-sheets="exportWalletToSheets"
  />
  <ShippingModal
    :visible="modals.shipping.visible"
    @update:visible="modals.shipping.visible = $event"
    :shipping-params="shippingParams"
    :shipping-files="shippingFiles"
    :show-shipping-form="showShippingForm"
    :show-shipping-file-list="showShippingFileList"
    @close="handleCloseShippingModal"
    @select-shipping-option="selectShippingOption"
    @load-shipping-files="loadShippingFiles"
    @process-shipping-file="processShippingFile"
    @update-shipping-params="handleUpdateShippingParams"
    @export-shipping-to-sheets="exportShippingToSheets"
  />
  <PriceModal
    :visible="modals.price.visible"
    @update:visible="modals.price.visible = $event"
    :platform="modals.price.platform"
    :current-price-option="currentPriceOption"
    @close="closeModal('price')"
    @select-price-option="handleSelectPriceOption"
    @execute-price-update="executePriceUpdate"
  />
  <ExportOrdersModal
    :visible="modals.exportOrders.visible"
    @update:visible="modals.exportOrders.visible = $event"
    :platform="modals.exportOrders.platform"
    @close="closeModal('exportOrders')"
    @execute-order-export="executeExportOrder"
  />
</template>

<script setup lang="ts">
interface Props {
  modals: any;
  walletParams: any;
  shippingParams: any;
  shippingFiles: any;
  showShippingForm: any;
  showShippingFileList: any;
  currentPriceOption: any;
}

interface Emits {
  (e: "handle-token-operation", operation: string): void;
  (e: "handle-update-wallet-params", params: Record<string, any>): void;
  (e: "handle-export-wallet"): void;
  (e: "handle-close-shipping"): void;
  (e: "handle-select-shipping-option", option: number): void;
  (e: "handle-load-shipping-files"): void;
  (e: "handle-process-shipping-file", filename: string): void;
  (e: "handle-update-shipping-params", params: Record<string, any>): void;
  (e: "handle-export-shipping"): void;
  (e: "handle-select-price-option", option: string): void;
  (e: "handle-execute-price-update"): void;
  (e: "handle-execute-order-export", type: string): void;
  (e: "close-modal", modal: string): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();

const closeModal = (modal: string) => emit("close-modal", modal);
const handleTokenOperation = (op: string) => emit("handle-token-operation", op);
const handleUpdateWalletParams = (params: Record<string, any>) =>
  emit("handle-update-wallet-params", params);
const exportWalletToSheets = () => emit("handle-export-wallet");
const handleCloseShippingModal = () => emit("handle-close-shipping");
const selectShippingOption = (option: number) =>
  emit("handle-select-shipping-option", option);
const loadShippingFiles = () => emit("handle-load-shipping-files");
const processShippingFile = (filename: string) =>
  emit("handle-process-shipping-file", filename);
const handleUpdateShippingParams = (params: Record<string, any>) =>
  emit("handle-update-shipping-params", params);
const exportShippingToSheets = () => emit("handle-export-shipping");
const handleSelectPriceOption = (option: string) =>
  emit("handle-select-price-option", option);
const executePriceUpdate = () => emit("handle-execute-price-update");
const executeExportOrder = (type: string) =>
  emit("handle-execute-order-export", type);

// Import modals
import { defineAsyncComponent } from "vue";
const TokenModal = defineAsyncComponent(
  () => import("../Modals/TokenModal.vue")
);
const WalletModal = defineAsyncComponent(
  () => import("../Modals/WalletModal.vue")
);
const ShippingModal = defineAsyncComponent(
  () => import("../Modals/ShippingModal.vue")
);
const PriceModal = defineAsyncComponent(
  () => import("../Modals/PriceModal.vue")
);
const ExportOrdersModal = defineAsyncComponent(
  () => import("../Modals/ExportOrdersModal.vue")
);
</script>
