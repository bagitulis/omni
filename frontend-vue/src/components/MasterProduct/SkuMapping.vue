<template>
  <div class="sku-mapping">
    <div class="mapping-header">
      <h3>SKU Mapping Status</h3>
      <button @click="refreshMapping" class="btn-refresh" :disabled="loading">
        <span :class="{ spinning: loading }">↻</span> Refresh
      </button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading mapping...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button @click="refreshMapping">Retry</button>
    </div>

    <div v-else-if="mappingData" class="mapping-content">
      <SkuMappingSummary :platforms="mappingData.platforms" />

      <SkuMappingTable
        :skus="mappingData.skus"
        @link="openLinkDialog"
        @auto-map="handleAutoMap"
        @unlink="showUnlinkMenu"
      />
    </div>

    <SkuMappingLinkModal
      :is-open="linkDialogOpen"
      :platform="linkDialogPlatform"
      :is-submitting="linking"
      @close="closeLinkDialog"
      @submit="submitLink"
    />

    <div v-if="message" :class="['toast-message', messageType]">
      {{ message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import masterProductService from "@/services/masterProductService";
import type {
  MappingStatus,
  SkuMappingInfo,
} from "@/services/masterProductService";
import SkuMappingSummary from "./SkuMappingSummary.vue";
import SkuMappingTable from "./SkuMappingTable.vue";
import SkuMappingLinkModal from "./SkuMappingLinkModal.vue";

interface Props {
  productId: number;
}
const props = defineProps<Props>();

// State
const loading = ref(true);
const error = ref("");
const mappingData = ref<MappingStatus | null>(null);
const linkDialogOpen = ref(false);
const linkDialogPlatform = ref("");
const linkDialogSku = ref<SkuMappingInfo | null>(null);
const linking = ref(false);
const message = ref("");
const messageType = ref<"success" | "error">("success");

// Methods
const refreshMapping = async () => {
  loading.value = true;
  error.value = "";

  try {
    mappingData.value = await masterProductService.getMappingStatus(
      props.productId,
    );
  } catch (err: any) {
    error.value = err.message || "Failed to load mapping";
  } finally {
    loading.value = false;
  }
};

const openLinkDialog = ({
  sku,
  platform,
}: {
  sku: SkuMappingInfo;
  platform: string;
}) => {
  linkDialogSku.value = sku;
  linkDialogPlatform.value = platform;
  linkDialogOpen.value = true;
};

const closeLinkDialog = () => {
  linkDialogOpen.value = false;
  linkDialogSku.value = null;
};

const submitLink = async (payload: {
  platformItemId: string;
  platformSkuId: string;
}) => {
  if (!linkDialogSku.value || !payload.platformItemId) return;

  linking.value = true;
  try {
    await masterProductService.manualLink(
      linkDialogSku.value.master_sku_id,
      linkDialogPlatform.value,
      payload.platformItemId,
      payload.platformSkuId || undefined,
    );
    showMessage("success", "SKU linked successfully");
    closeLinkDialog();
    refreshMapping();
  } catch (err: any) {
    showMessage("error", err.message || "Failed to link");
  } finally {
    linking.value = false;
  }
};

const handleAutoMap = async (sku: SkuMappingInfo) => {
  try {
    const result = await masterProductService.autoMapSku(sku.seller_sku);
    if (result.matches?.length > 0) {
      showMessage("success", `Found ${result.matches.length} match(es)`);
    } else {
      showMessage("error", "No matches found");
    }
    refreshMapping();
  } catch (err: any) {
    showMessage("error", err.message || "Auto-map failed");
  }
};

const showUnlinkMenu = async (sku: SkuMappingInfo) => {
  const platform = prompt("Enter platform to unlink (shopee/tiktok/lazada):");
  if (!platform) return;

  try {
    await masterProductService.unlinkSku(sku.master_sku_id, platform);
    showMessage("success", "Unlinked successfully");
    refreshMapping();
  } catch (err: any) {
    showMessage("error", err.message || "Unlink failed");
  }
};

const showMessage = (type: "success" | "error", msg: string) => {
  messageType.value = type;
  message.value = msg;
  setTimeout(() => {
    message.value = "";
  }, 3000);
};

onMounted(refreshMapping);
</script>

<style scoped>
.sku-mapping {
  padding: 1.5rem;
}

.mapping-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.mapping-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  cursor: pointer;
}

.spinning {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem;
  gap: 1rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f4f6;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.toast-message {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 0.75rem 1.5rem;
  border-radius: 0.375rem;
  font-weight: 500;
  z-index: 1001;
}

.toast-message.success {
  background: #dcfce7;
  color: #166534;
}
.toast-message.error {
  background: #fee2e2;
  color: #991b1b;
}
</style>
