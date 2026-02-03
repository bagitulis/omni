<template>
  <div class="tab-content">
    <div class="info-banner">
      <Icon name="settings" size="sm" />
      <div class="info-text">
        <strong>Pengaturan Bulk Pricing</strong>
        <p class="info-note">
          Konfigurasi tier dan admin fee untuk perhitungan harga
        </p>
      </div>
    </div>

    <form @submit.prevent="handleSave" class="settings-form">
      <!-- Admin Fee -->
      <div class="form-group">
        <label>Admin Fee (Rp)</label>
        <input
          type="number"
          v-model.number="form.admin_fee"
          min="0"
          step="100"
          required
        />
        <span class="help-text">Biaya admin yang akan dibagi ke tier</span>
      </div>

      <!-- Tier 1 -->
      <div class="tier-group">
        <h4>Tier 1</h4>
        <div class="tier-inputs">
          <div class="form-group">
            <label>Min Order</label>
            <input
              type="number"
              v-model.number="form.min_order_1"
              min="2"
              required
            />
          </div>
          <div class="form-group">
            <label>Max Order</label>
            <input
              type="number"
              v-model.number="form.max_order_1"
              :min="form.min_order_1"
              required
            />
          </div>
        </div>
      </div>

      <!-- Tier 2 (auto-calculated) -->
      <div class="tier-group readonly">
        <h4>Tier 2 <span class="auto-badge">Auto</span></h4>
        <div class="tier-inputs">
          <div class="form-group">
            <label>Min Order</label>
            <input type="number" :value="tier2Min" disabled />
          </div>
          <div class="form-group">
            <label>Max Order</label>
            <input type="number" :value="tier2Max" disabled />
          </div>
        </div>
      </div>

      <!-- Tier 3 -->
      <div class="tier-group">
        <h4>Tier 3</h4>
        <div class="tier-inputs">
          <div class="form-group">
            <label>Min Order</label>
            <input type="number" :value="tier3Min" disabled />
            <span class="help-text">Auto dari Tier 2 + 1</span>
          </div>
          <div class="form-group">
            <label>Max Order</label>
            <input
              type="number"
              v-model.number="form.max_order_tier_3"
              :min="tier3Min"
              required
            />
          </div>
        </div>
      </div>

      <!-- Preview Calculation -->
      <div class="preview-calc">
        <h4>Preview Perhitungan</h4>
        <div class="calc-input">
          <label>Harga Sample:</label>
          <input type="number" v-model.number="samplePrice" min="0" />
        </div>
        <table class="calc-table" aria-label="Price calculation preview">
          <thead>
            <tr>
              <th scope="col">Tier</th>
              <th scope="col">Range</th>
              <th scope="col">Harga/Unit</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Tier 1</td>
              <td>{{ form.min_order_1 }} - {{ form.max_order_1 }}</td>
              <td class="price-cell">{{ formatPrice(calcTier1) }}</td>
            </tr>
            <tr>
              <td>Tier 2</td>
              <td>{{ tier2Min }} - {{ tier2Max }}</td>
              <td class="price-cell">{{ formatPrice(calcTier2) }}</td>
            </tr>
            <tr>
              <td>Tier 3</td>
              <td>{{ tier3Min }} - {{ form.max_order_tier_3 }}</td>
              <td class="price-cell">{{ formatPrice(calcTier3) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Save Button -->
      <div class="action-buttons">
        <button
          type="submit"
          class="btn-save"
          :disabled="saving"
          aria-label="Save settings"
        >
          <Icon name="download" size="sm" />
          {{ saving ? "Menyimpan..." : "Simpan Pengaturan" }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Icon from "@/components/ui/Icon.vue";
import type { WholesaleSettings } from "../../../../services/wholesaleService";
import wholesaleService from "../../../../services/wholesaleService";

const props = defineProps<{
  settings: WholesaleSettings;
}>();

const emit = defineEmits<{
  saved: [settings: WholesaleSettings];
}>();

const saving = ref(false);
const samplePrice = ref(100000);

const form = ref<WholesaleSettings>({ ...props.settings });

watch(
  () => props.settings,
  (newVal) => {
    form.value = { ...newVal };
  },
);

const tier2Min = computed(() => form.value.max_order_1 + 1);
const tier2Max = computed(() => tier2Min.value + 1);
const tier3Min = computed(() => tier2Max.value + 1);

// Price calculations
const calcTier1 = computed(() => {
  const base = samplePrice.value - form.value.admin_fee;
  return base + form.value.admin_fee / form.value.min_order_1;
});

const calcTier2 = computed(() => {
  const base = samplePrice.value - form.value.admin_fee;
  return base + form.value.admin_fee / tier2Min.value;
});

const calcTier3 = computed(() => {
  const base = samplePrice.value - form.value.admin_fee;
  return base + form.value.admin_fee / tier3Min.value;
});

function formatPrice(price: number): string {
  return "Rp " + new Intl.NumberFormat("id-ID").format(Math.round(price));
}

async function handleSave() {
  saving.value = true;
  try {
    await wholesaleService.updateSettings(form.value);
    emit("saved", form.value);
  } catch (err: any) {
    console.error("Failed to save settings:", err);
    alert("Gagal menyimpan: " + (err.message || "Unknown error"));
  } finally {
    saving.value = false;
  }
}
</script>

<style src="./SettingsTab.styles.css" scoped></style>
