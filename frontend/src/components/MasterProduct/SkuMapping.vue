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
      <!-- Summary -->
      <div class="mapping-summary">
        <div
          class="summary-item"
          v-for="(stats, platform) in mappingData.platforms"
          :key="platform"
        >
          <span class="platform-icon">{{ getPlatformIcon(platform) }}</span>
          <span class="platform-name">{{ platform }}</span>
          <span class="linked-count">{{ stats.linked }}/{{ stats.total }}</span>
        </div>
      </div>

      <!-- SKU Table -->
      <div class="mapping-table-wrapper">
        <table class="mapping-table">
          <thead>
            <tr>
              <th>Seller SKU</th>
              <th>Variant</th>
              <th>🟠 Shopee</th>
              <th>⬛ TikTok</th>
              <th>🔵 Lazada</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sku in mappingData.skus" :key="sku.master_sku_id">
              <td class="sku-code">{{ sku.seller_sku }}</td>
              <td>{{ sku.variant_name || "-" }}</td>

              <td class="platform-cell">
                <span v-if="getPlatformLink(sku, 'shopee')" class="linked-badge"
                  >✓ Linked</span
                >
                <button
                  v-else
                  @click="openLinkDialog(sku, 'shopee')"
                  class="btn-link"
                >
                  Link
                </button>
              </td>
              <td class="platform-cell">
                <span v-if="getPlatformLink(sku, 'tiktok')" class="linked-badge"
                  >✓ Linked</span
                >
                <button
                  v-else
                  @click="openLinkDialog(sku, 'tiktok')"
                  class="btn-link"
                >
                  Link
                </button>
              </td>
              <td class="platform-cell">
                <span v-if="getPlatformLink(sku, 'lazada')" class="linked-badge"
                  >✓ Linked</span
                >
                <button
                  v-else
                  @click="openLinkDialog(sku, 'lazada')"
                  class="btn-link"
                >
                  Link
                </button>
              </td>

              <td class="actions-cell">
                <button
                  @click="handleAutoMap(sku)"
                  title="Auto-map"
                  class="btn-action"
                >
                  🔍
                </button>
                <button
                  v-if="hasAnyLink(sku)"
                  @click="showUnlinkMenu(sku)"
                  title="Unlink"
                  class="btn-action"
                >
                  🔗
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Link Dialog Modal -->
    <div
      v-if="linkDialogOpen"
      class="modal-overlay"
      @click.self="closeLinkDialog"
    >
      <div class="modal-content">
        <h4>Link SKU to {{ linkDialogPlatform }}</h4>

        <div class="form-group">
          <label>Platform Item ID</label>
          <input
            v-model="linkForm.platformItemId"
            placeholder="e.g., 12345678"
          />
        </div>

        <div class="form-group">
          <label>Platform SKU ID (optional)</label>
          <input
            v-model="linkForm.platformSkuId"
            placeholder="e.g., sku_12345"
          />
        </div>

        <div class="modal-actions">
          <button @click="closeLinkDialog" class="btn-cancel">Cancel</button>
          <button @click="submitLink" class="btn-submit" :disabled="linking">
            {{ linking ? "Linking..." : "Link" }}
          </button>
        </div>
      </div>
    </div>

    <!-- Messages -->
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
const linkForm = ref({ platformItemId: "", platformSkuId: "" });
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

const getPlatformIcon = (platform: string): string => {
  const icons: Record<string, string> = {
    shopee: "🟠",
    tiktok: "⬛",
    lazada: "🔵",
  };
  return icons[platform] || "⚪";
};

const getPlatformLink = (sku: SkuMappingInfo, platform: string) => {
  return sku.platform_links?.find((link) => link.platform === platform);
};

const hasAnyLink = (sku: SkuMappingInfo): boolean => {
  return (sku.platform_links?.length || 0) > 0;
};

const openLinkDialog = (sku: SkuMappingInfo, platform: string) => {
  linkDialogSku.value = sku;
  linkDialogPlatform.value = platform;
  linkForm.value = { platformItemId: "", platformSkuId: "" };
  linkDialogOpen.value = true;
};

const closeLinkDialog = () => {
  linkDialogOpen.value = false;
  linkDialogSku.value = null;
};

const submitLink = async () => {
  if (!linkDialogSku.value || !linkForm.value.platformItemId) return;

  linking.value = true;
  try {
    await masterProductService.manualLink(
      linkDialogSku.value.master_sku_id,
      linkDialogPlatform.value,
      linkForm.value.platformItemId,
      linkForm.value.platformSkuId || undefined,
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

.mapping-summary {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.5rem;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.platform-icon {
  font-size: 1.25rem;
}
.platform-name {
  font-weight: 600;
  text-transform: capitalize;
}
.linked-count {
  color: #6b7280;
  font-size: 0.875rem;
}

.mapping-table-wrapper {
  overflow-x: auto;
}

.mapping-table {
  width: 100%;
  border-collapse: collapse;
}

.mapping-table th,
.mapping-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.mapping-table th {
  background: #f9fafb;
  font-weight: 600;
  font-size: 0.875rem;
}

.sku-code {
  font-family: monospace;
  font-weight: 600;
}

.platform-cell {
  text-align: center;
}

.linked-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: #dcfce7;
  color: #166534;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.btn-link {
  padding: 0.25rem 0.75rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  cursor: pointer;
}

.btn-link:hover {
  background: #2563eb;
}

.actions-cell {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  width: 32px;
  height: 32px;
  border: 1px solid #e5e7eb;
  background: white;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 1rem;
}

.btn-action:hover {
  background: #f3f4f6;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  width: 100%;
  max-width: 400px;
}

.modal-content h4 {
  margin: 0 0 1.5rem;
  text-transform: capitalize;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.btn-cancel {
  padding: 0.5rem 1rem;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  cursor: pointer;
}

.btn-submit {
  padding: 0.5rem 1rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
}

.btn-submit:disabled {
  opacity: 0.6;
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
