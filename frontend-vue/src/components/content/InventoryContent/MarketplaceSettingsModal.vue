<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>⚙️ Marketplace Allocation Settings</h3>
        <button class="close-btn" @click="$emit('close')">&times;</button>
      </div>

      <div class="modal-body">
        <div class="settings-section">
          <h4>📊 Column Mapping</h4>
          <p class="section-desc">
            Select which columns to use for Total and Auto values
          </p>

          <div class="form-group">
            <label for="totalColumn">Total Column:</label>
            <select
              id="totalColumn"
              v-model="localSettings.totalColumn"
              class="form-select"
            >
              <option value="">-- Select Column --</option>
              <option v-for="col in numericColumns" :key="col" :value="col">
                {{ col }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label for="autoColumn">Auto Column (Boolean):</label>
            <select
              id="autoColumn"
              v-model="localSettings.autoColumn"
              class="form-select"
            >
              <option value="">-- Select Column --</option>
              <option v-for="col in booleanColumns" :key="col" :value="col">
                {{ col }}
              </option>
            </select>
          </div>
        </div>

        <div class="settings-section">
          <h4>📈 Allocation Ratios</h4>
          <p class="section-desc">
            Set the percentage ratio for each marketplace
          </p>

          <div class="form-group">
            <label for="shopeeRatio">
              Shopee Ratio: {{ (localSettings.shopeeRatio * 100).toFixed(0) }}%
            </label>
            <input
              id="shopeeRatio"
              type="range"
              v-model.number="localSettings.shopeeRatio"
              min="0"
              max="1"
              step="0.05"
              class="form-range"
            />
          </div>

          <div class="form-group">
            <label for="tiktokRatio">
              Tiktok Ratio: {{ (localSettings.tiktokRatio * 100).toFixed(0) }}%
            </label>
            <input
              id="tiktokRatio"
              type="range"
              v-model.number="localSettings.tiktokRatio"
              min="0"
              max="1"
              step="0.05"
              class="form-range"
            />
          </div>

          <div class="ratio-preview">
            <span class="preview-item shopee">
              Shopee: {{ previewShopee }}
            </span>
            <span class="preview-item tiktok">
              Tiktok: {{ previewTiktok }}
            </span>
            <span class="preview-item lazada">
              Lazada: {{ previewLazada }}
            </span>
            <span class="preview-total">
              = {{ previewTotal }} (Total: 20)
            </span>
          </div>
        </div>

        <div class="formula-info">
          <h4>📝 Formula Reference</h4>
          <ul>
            <li><strong>Shopee:</strong> MIN(CEILING(ratio × Total), Total)</li>
            <li>
              <strong>Tiktok:</strong> MIN(CEILING(ratio × Total), Total -
              Shopee)
            </li>
            <li><strong>Lazada:</strong> Total - Shopee - Tiktok</li>
            <li><strong>Auto = TRUE:</strong> All platforms = Total</li>
          </ul>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="resetDefaults">
          🔄 Reset Defaults
        </button>
        <button class="btn btn-primary" @click="saveAndClose">
          💾 Save Settings
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useMarketplaceSettings } from "./composables/useMarketplaceSettings";

interface Props {
  show: boolean;
  schemaColumns: Array<{ column_name: string; column_type?: string }>;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
  saved: [];
}>();

const { getSettings, updateSettings, resetToDefaults } =
  useMarketplaceSettings();

const localSettings = ref({
  totalColumn: "",
  autoColumn: "",
  shopeeRatio: 0.6,
  tiktokRatio: 0.3,
});

// Load settings on mount
onMounted(() => {
  const current = getSettings();
  localSettings.value = { ...current };
});

// Reload when modal opens
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      const current = getSettings();
      localSettings.value = { ...current };
    }
  }
);

// Filter columns by type - show all with suggested ones first
const numericColumns = computed(() => {
  const allCols = props.schemaColumns.map((c) => c.column_name);

  // Suggested columns (likely numeric)
  const suggested = allCols.filter((name) => {
    const lower = name.toLowerCase();
    return (
      lower.includes("total") ||
      lower.includes("qty") ||
      lower.includes("stock") ||
      lower.includes("quantity") ||
      lower.includes("jumlah")
    );
  });

  // Rest of columns
  const others = allCols.filter((name) => !suggested.includes(name));

  return [...suggested, ...others];
});

const booleanColumns = computed(() => {
  const allCols = props.schemaColumns.map((c) => c.column_name);

  // Suggested columns (likely boolean)
  const suggested = allCols.filter((name) => {
    const lower = name.toLowerCase();
    return (
      lower.includes("auto") ||
      lower.includes("active") ||
      lower.includes("enabled") ||
      lower.includes("flag") ||
      lower.includes("status") ||
      lower.includes("bool")
    );
  });

  // Rest of columns
  const others = allCols.filter((name) => !suggested.includes(name));

  return [...suggested, ...others];
});

// Preview calculation with sample Total = 20
const previewShopee = computed(() => {
  const total = 20;
  return Math.min(Math.ceil(localSettings.value.shopeeRatio * total), total);
});

const previewTiktok = computed(() => {
  const total = 20;
  const remaining = Math.max(0, total - previewShopee.value);
  return Math.min(
    Math.ceil(localSettings.value.tiktokRatio * total),
    remaining
  );
});

const previewLazada = computed(() => {
  const total = 20;
  return Math.max(0, total - previewShopee.value - previewTiktok.value);
});

const previewTotal = computed(() => {
  return previewShopee.value + previewTiktok.value + previewLazada.value;
});

function saveAndClose() {
  updateSettings(localSettings.value);
  emit("saved");
  emit("close");
}

function resetDefaults() {
  resetToDefaults();
  const current = getSettings();
  localSettings.value = { ...current };
}
</script>

<style scoped>
@import "./MarketplaceSettingsModal.styles.css";
</style>
