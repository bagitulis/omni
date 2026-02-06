<template>
  <Modal
    :is-open="visible"
    :title="`Export ${getPlatformLabel()} Orders`"
    @close="onModalClose"
  >
    <!-- Loading State -->
    <div
      v-if="exporting"
      class="flex flex-col items-center justify-center py-12 gap-4"
    >
      <span class="loading loading-spinner loading-lg text-primary"></span>
      <p class="text-base-content/70 font-medium">Exporting orders...</p>
    </div>

    <!-- Options Grid -->
    <div v-else class="grid grid-cols-1 gap-3">
      <button
        v-for="option in exportOptions"
        :key="option.value"
        @click="selectOption(option.value)"
        :class="{
          'btn btn-primary': selectedOption === option.value,
          'btn btn-outline': selectedOption !== option.value,
        }"
        class="h-auto py-3 px-4 flex items-center gap-4 justify-start"
      >
        <div class="text-3xl flex-shrink-0">
          {{ option.icon }}
        </div>

        <!-- Content -->
        <div class="flex-1 text-left">
          <h3 class="font-semibold text-base-content">{{ option.label }}</h3>
          <p class="text-xs text-base-content/70">{{ option.description }}</p>
        </div>

        <!-- Checkmark -->
        <div
          v-if="selectedOption === option.value"
          class="badge badge-primary flex-shrink-0"
        >
          ✓
        </div>
      </button>
    </div>

    <template #footer>
      <div class="flex gap-2 justify-end">
        <Button
          label="Cancel"
          color="ghost"
          size="sm"
          :disabled="exporting"
          @click="closeModal"
        />
        <Button
          label="Export"
          color="primary"
          size="sm"
          :loading="exporting"
          :disabled="!selectedOption || exporting"
          @click="confirmExport"
        />
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import Modal from "@/components/Modal.vue";
import Button from "@/components/Button.vue";
import { useToast } from "@/composables/useToast";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

// Get base URL for API
const API_BASE_URL = getApiBaseUrl();

interface ExportOption {
  value: string;
  icon: string;
  label: string;
  description: string;
}

interface Props {
  visible?: boolean;
  platform?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  platform: null,
});

const emit = defineEmits<{
  close: [];
  "execute-order-export": [data: any];
  "update:visible": [value: boolean];
}>();

const toast = useToast();
const selectedOption = ref<string | null>(null);
const exporting = ref(false);

const getPlatformLabel = (): string => {
  return props.platform
    ? props.platform.charAt(0).toUpperCase() + props.platform.slice(1)
    : "Orders";
};

const exportOptions = computed(() => {
  const options: Record<string, ExportOption[]> = {
    shopee: [
      {
        value: "UNPAID",
        icon: "⏳",
        label: "Unpaid Orders",
        description: "Orders waiting for customer payment",
      },
      {
        value: "READY_TO_SHIP",
        icon: "🚚",
        label: "Ready to Ship",
        description: "Orders ready to be shipped out",
      },
      {
        value: "PROCESSED",
        icon: "⚙️",
        label: "Processing",
        description: "Orders currently being processed",
      },
      {
        value: "COMPLETED",
        icon: "✅",
        label: "Completed Orders",
        description: "Finished and delivered orders",
      },
      {
        value: "ALL",
        icon: "📋",
        label: "All Orders",
        description: "Export all order statuses combined",
      },
    ],
    lazada: [
      {
        value: "unpaid",
        icon: "⏳",
        label: "Unpaid Orders",
        description: "Orders waiting for payment",
      },
      {
        value: "pending",
        icon: "⚠️",
        label: "Pending",
        description: "Orders in pending status",
      },
      {
        value: "topack",
        icon: "📦",
        label: "To Pack",
        description: "Orders waiting to be packed",
      },
      {
        value: "toship",
        icon: "🚚",
        label: "To Ship",
        description: "Orders waiting to be shipped",
      },
      {
        value: "ALL",
        icon: "📋",
        label: "All Orders",
        description: "Export all order statuses combined",
      },
    ],
    tiktok: [
      {
        value: "UNPAID",
        icon: "⏳",
        label: "Unpaid Orders",
        description: "Orders pending payment",
      },
      {
        value: "AWAITING_SHIPMENT",
        icon: "🚚",
        label: "Awaiting Shipment",
        description: "Orders ready for shipping",
      },
      {
        value: "AWAITING_COLLECTION",
        icon: "📥",
        label: "Awaiting Collection",
        description: "Orders awaiting customer collection",
      },
      {
        value: "COMPLETED",
        icon: "✅",
        label: "Completed Orders",
        description: "Successfully completed orders",
      },
      {
        value: "ALL",
        icon: "📋",
        label: "All Orders",
        description: "Export all order statuses combined",
      },
    ],
  };
  return options[props.platform || ""] || [];
});

const selectOption = (value: string): void => {
  selectedOption.value = value;
};

const confirmExport = async (): Promise<void> => {
  if (!selectedOption.value) return;
  await executeOrderExport(selectedOption.value);
};

const executeOrderExport = async (exportType: string): Promise<void> => {
  exporting.value = true;
  try {
    const response = await fetch(`${API_BASE_URL}/execute-order-export`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify({
        platform: props.platform,
        order_type: exportType,
        days: 7,
      }),
    });

    const result = await response.json();

    if (result.success) {
      toast.success(
        "Export Successful",
        `${exportType} orders exported to Google Sheets`
      );
      closeModal();
    } else {
      toast.error("Export Failed", result.message || "Failed to export orders");
    }
  } catch (error: unknown) {
    const errorMessage =
      error instanceof Error ? error.message : "An unknown error occurred";
    toast.error("Error", errorMessage);
  } finally {
    exporting.value = false;
  }
};

const closeModal = (): void => {
  selectedOption.value = null;
  emit("update:visible", false);
  emit("close");
};

const onModalClose = (): void => {
  selectedOption.value = null;
  emit("update:visible", false);
};
</script>

<style src="./ExportOrdersModal.styles.css" scoped></style>
