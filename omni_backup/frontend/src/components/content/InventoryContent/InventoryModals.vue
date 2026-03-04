<template>
  <div class="inventory-modals">
    <InventorySyncHistory
      :sync-history="syncHistory"
      :show="showSyncHistory"
      @close-history="$emit('close-sync-history')"
    />

    <MarketplaceSettingsModal
      :show="showMarketplaceSettings"
      :schema-columns="schemaColumns"
      @close="$emit('close-marketplace-settings')"
      @saved="$emit('marketplace-saved')"
    />

    <WholesaleBatchDeleteModal
      :show="showWholesaleModal"
      :skus="wholesaleSkus"
      @close="$emit('close-wholesale-modal')"
      @completed="$emit('wholesale-completed', $event)"
    />

    <WholesaleUpdateModal
      :show="showWholesaleUpdateModal"
      :items="wholesaleUpdateItems"
      @close="$emit('close-wholesale-update')"
      @completed="$emit('wholesale-update-completed', $event)"
    />

    <WholesaleMpqModal
      :show="showWholesaleMpqModal"
      :items="wholesaleMpqItems"
      @close="$emit('close-mpq-modal')"
      @completed="$emit('mpq-completed', $event)"
    />
  </div>
</template>

<script>
import { defineComponent } from "vue";
import InventorySyncHistory from "./InventorySyncHistory.vue";
import MarketplaceSettingsModal from "./MarketplaceSettingsModal.vue";
import WholesaleBatchDeleteModal from "./WholesaleBatchDeleteModal.vue";
import WholesaleUpdateModal from "./WholesaleUpdateModal.vue";
import WholesaleMpqModal from "./WholesaleMpqModal.vue";

export default defineComponent({
  name: "InventoryModals",
  components: {
    InventorySyncHistory,
    MarketplaceSettingsModal,
    WholesaleBatchDeleteModal,
    WholesaleUpdateModal,
    WholesaleMpqModal,
  },
  props: {
    syncHistory: { type: Array, default: () => [] },
    showSyncHistory: { type: Boolean, default: false },
    showMarketplaceSettings: { type: Boolean, default: false },
    schemaColumns: { type: Array, default: () => [] },
    showWholesaleModal: { type: Boolean, default: false },
    wholesaleSkus: { type: Array, default: () => [] },
    showWholesaleUpdateModal: { type: Boolean, default: false },
    wholesaleUpdateItems: { type: Array, default: () => [] },
    showWholesaleMpqModal: { type: Boolean, default: false },
    wholesaleMpqItems: { type: Array, default: () => [] },
  },
  emits: [
    "close-sync-history",
    "close-marketplace-settings",
    "marketplace-saved",
    "close-wholesale-modal",
    "wholesale-completed",
    "close-wholesale-update",
    "wholesale-update-completed",
    "close-mpq-modal",
    "mpq-completed",
  ],
});
</script>
