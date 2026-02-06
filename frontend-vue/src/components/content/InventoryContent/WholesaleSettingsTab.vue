<template>
  <div class="tab-content">
    <div class="settings-form">
      <div class="setting-group">
        <label>Admin Fee (Rp)</label>
        <input
          type="number"
          :value="settings.admin_fee"
          @input="
            updateField(
              'admin_fee',
              parseInt(($event.target as HTMLInputElement).value),
            )
          "
          min="0"
          step="100"
        />
        <span class="setting-desc"
          >Biaya admin per transaksi (default: 1500)</span
        >
      </div>

      <div class="setting-group">
        <label>Tier 1 Min Order</label>
        <input
          type="number"
          :value="settings.min_order_1"
          @input="
            updateField(
              'min_order_1',
              parseInt(($event.target as HTMLInputElement).value),
            )
          "
          min="2"
          max="99"
        />
        <span class="setting-desc">Minimal order untuk tier 1</span>
      </div>

      <div class="setting-group">
        <label>Tier 1 Max Order</label>
        <input
          type="number"
          :value="settings.max_order_1"
          @input="
            updateField(
              'max_order_1',
              parseInt(($event.target as HTMLInputElement).value),
            )
          "
          :min="settings.min_order_1"
          max="99"
        />
        <span class="setting-desc">Maksimal order untuk tier 1</span>
      </div>

      <div class="setting-group">
        <label>Tier 3 Max Order</label>
        <input
          type="number"
          :value="settings.max_order_tier_3"
          @input="
            updateField(
              'max_order_tier_3',
              parseInt(($event.target as HTMLInputElement).value),
            )
          "
          min="10"
          max="10000"
        />
        <span class="setting-desc"
          >Maksimal order untuk tier 3 (default: 1000)</span
        >
      </div>

      <div class="tier-preview">
        <h4>Preview Tier Ranges</h4>
        <div class="tier-ranges">
          <div class="tier-item">
            <span class="tier-label">Tier 1:</span>
            <span class="tier-value"
              >{{ settings.min_order_1 }} - {{ settings.max_order_1 }}</span
            >
          </div>
          <div class="tier-item">
            <span class="tier-label">Tier 2:</span>
            <span class="tier-value">{{ tier2Min }} - {{ tier2Max }}</span>
          </div>
          <div class="tier-item">
            <span class="tier-label">Tier 3:</span>
            <span class="tier-value"
              >{{ tier3Min }} - {{ settings.max_order_tier_3 }}</span
            >
          </div>
        </div>
      </div>

      <button class="btn-save-settings" @click="$emit('save')">
        Simpan Settings
      </button>
      <p v-if="settingsSaved" class="settings-saved-msg">Settings tersimpan!</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { WholesaleSettings } from "@/services/wholesaleService";

const props = defineProps<{
  settings: WholesaleSettings;
  tier2Min: number;
  tier2Max: number;
  tier3Min: number;
  settingsSaved: boolean;
}>();

const emit = defineEmits<{
  save: [];
  "update:settings": [settings: WholesaleSettings];
}>();

function updateField(field: keyof WholesaleSettings, value: number) {
  emit("update:settings", { ...props.settings, [field]: value });
}
</script>

<style src="./WholesaleUpdateModal.styles.css" scoped></style>
