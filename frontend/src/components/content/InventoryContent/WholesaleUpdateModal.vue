<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="closeIfNotProcessing">
      <div class="modal-content wholesale-modal">
        <div class="modal-header">
          <h3>
            <Icon name="shopping-cart" size="sm" /> Update Wholesale - Shopee
          </h3>
          <button class="close-btn" @click="closeIfNotProcessing">✕</button>
        </div>

        <div class="modal-body">
          <div class="tab-nav">
            <button
              class="tab-btn"
              :class="{ active: activeTab === 'preview' }"
              @click="activeTab = 'preview'"
            >
              <Icon name="search" size="sm" /> Preview
            </button>
            <button
              class="tab-btn"
              :class="{ active: activeTab === 'settings' }"
              @click="activeTab = 'settings'"
            >
              <Icon name="settings" size="sm" /> Settings
            </button>
          </div>

          <WholesalePreviewTab
            v-if="activeTab === 'preview'"
            :item-count="items.length"
            :preview-items="previewItems"
            :min-order1="settings.min_order_1"
            :max-order1="settings.max_order_1"
            :tier2-min="tier2Min"
            :tier2-max="tier2Max"
            :tier3-min="tier3Min"
            :max-order-tier3="settings.max_order_tier_3"
            :format-price="formatPrice"
          />

          <WholesaleSettingsTab
            v-if="activeTab === 'settings'"
            :settings="settings"
            :tier2-min="tier2Min"
            :tier2-max="tier2Max"
            :tier3-min="tier3Min"
            :settings-saved="settingsSaved"
            @save="saveSettings"
            @update:settings="settings = $event"
          />

          <div v-if="processing" class="processing-status">
            <div class="spinner"></div>
            <span>Memproses update wholesale...</span>
          </div>

          <div v-if="result" class="result-section" :class="resultClass">
            <div class="result-header">
              <Icon
                v-if="result.success"
                name="check"
                size="sm"
                class="text-green-600"
              />
              <Icon v-else name="warning" size="sm" class="text-yellow-600" />
              {{ result.message }}
            </div>
            <div class="result-details">
              <p>Total SKU: {{ result.data.total_skus }}</p>
              <p>Item Unik: {{ result.data.unique_items }}</p>
              <p>Berhasil: {{ result.data.processed }}</p>
              <p v-if="result.data.failed > 0" class="failed">
                Gagal: {{ result.data.failed }}
              </p>
              <p v-if="result.data.skipped.length > 0" class="skipped">
                SKU tidak ditemukan: {{ result.data.skipped.join(", ") }}
              </p>
            </div>
          </div>

          <div v-if="error" class="error-message">
            <Icon name="warning" size="sm" class="text-yellow-600" />
            {{ error }}
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-cancel" @click="closeIfNotProcessing">
            {{ result ? "Tutup" : "Batal" }}
          </button>
          <button
            v-if="!result"
            class="btn-update"
            @click="handleUpdate(items)"
            :disabled="items.length === 0 || processing"
          >
            <Icon name="shopping-cart" size="sm" />
            {{ processing ? "Memproses..." : "Update Wholesale" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { toRef } from "vue";
import Icon from "@/components/ui/Icon.vue";
import WholesalePreviewTab from "./WholesalePreviewTab.vue";
import WholesaleSettingsTab from "./WholesaleSettingsTab.vue";
import {
  useWholesaleUpdate,
  type UpdateItem,
} from "./composables/useWholesaleUpdate";
import type { BatchUpdateBySkusResult } from "@/services/wholesaleService";

const props = defineProps<{
  show: boolean;
  items: UpdateItem[];
}>();

const emit = defineEmits<{
  close: [];
  completed: [result: BatchUpdateBySkusResult];
}>();

const showRef = toRef(props, "show");
const itemsRef = toRef(props, "items");

const {
  activeTab,
  processing,
  result,
  error,
  settingsSaved,
  settings,
  tier2Min,
  tier2Max,
  tier3Min,
  previewItems,
  resultClass,
  saveSettings,
  handleUpdate,
  formatPrice,
} = useWholesaleUpdate(showRef, itemsRef, (res) => emit("completed", res));

function closeIfNotProcessing() {
  if (!processing.value) {
    emit("close");
  }
}
</script>

<style src="./WholesaleUpdateModal.styles.css" scoped></style>
