<template>
  <Modal
    :is-open="visible"
    title="Shipping Fee Check"
    @close="$emit('update:visible', false)"
  >
    <div class="shipping-modal-content">
      <!-- Option 1: Get from Wallet -->
      <ShippingForm
        v-if="showShippingForm"
        :params="internalShippingParams"
        @update:params="updateShippingParams"
      />

      <!-- Option 2: File List -->
      <ShippingFiles
        v-if="showShippingFileList"
        :files="shippingFiles"
        @process="$emit('process-shipping-file', $event)"
      />
    </div>

    <template #footer>
      <div class="footer-actions">
        <button class="btn-cancel" @click="$emit('update:visible', false)">
          Cancel
        </button>
        <button
          v-if="showShippingForm"
          class="btn-primary"
          @click="$emit('export-shipping-to-sheets')"
        >
          ✓ Export to Sheets
        </button>
        <button
          v-if="showShippingFileList"
          class="btn-secondary"
          @click="$emit('load-shipping-files')"
        >
          ↻ Refresh Files
        </button>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import Modal from "@/components/Modal.vue";
import ShippingForm from "./ShippingModal/ShippingForm.vue";
import ShippingFiles from "./ShippingModal/ShippingFiles.vue";
import "./ShippingModal.styles.css";

interface ShippingParams {
  [key: string]: any;
}

interface Props {
  visible?: boolean;
  shippingParams?: ShippingParams;
  shippingFiles?: any[];
  selectedFile?: string;
  showShippingForm?: boolean;
  showShippingFileList?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  shippingParams: () => ({}),
  shippingFiles: () => [],
  selectedFile: "",
  showShippingForm: false,
  showShippingFileList: false,
});

const emit = defineEmits<{
  "update:visible": [value: boolean];
  "update-shipping-params": [params: ShippingParams];
  "load-shipping-files": [];
  "select-shipping-option": [option: number];
  "process-shipping-file": [filename: string];
  "export-shipping-to-sheets": [];
  close: [];
}>();

const internalShippingParams = ref<ShippingParams>({ ...props.shippingParams });
const showShippingForm = ref<boolean>(false);
const showShippingFileList = ref<boolean>(false);

const updateShippingParams = (params: ShippingParams): void => {
  internalShippingParams.value = params;
  emit("update-shipping-params", params);
};

watch(
  () => props.shippingParams,
  (newParams: ShippingParams) => {
    internalShippingParams.value = { ...newParams };
  },
  { deep: true }
);

watch(
  () => props.showShippingForm,
  (newVal: boolean) => {
    showShippingForm.value = newVal;
  }
);

watch(
  () => props.showShippingFileList,
  (newVal: boolean) => {
    showShippingFileList.value = newVal;
  }
);
</script>

<style src="./ShippingModal.styles.css" scoped></style>
