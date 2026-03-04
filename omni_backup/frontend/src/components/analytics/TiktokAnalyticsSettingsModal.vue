<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h2>⚙️ TikTok Analytics Settings</h2>
        <button
          class="btn-close"
          type="button"
          aria-label="Close modal"
          @click="$emit('close')"
        >
          ×
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label>Price Column Name</label>
          <select v-model="localSettings.priceColumn" class="form-control">
            <option value="">-- Select Column --</option>
            <option v-for="col in priceColumns" :key="col" :value="col">
              {{ col }}
            </option>
          </select>
          <p class="help-text">
            Column name in inventory that contains marketplace price
          </p>
        </div>

        <div class="form-group">
          <label>Formula: Expected Income</label>
          <div class="formula-box">
            <code
              >(Harga Marketplace - {{ localSettings.formulaDeduction }}) ×
              {{ localSettings.formulaMultiplier }}</code
            >
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Admin Fee (Rp)</label>
            <input
              v-model.number="localSettings.formulaDeduction"
              type="number"
              min="0"
              class="form-control"
            />
            <p class="help-text">TikTok admin fee deduction</p>
          </div>
          <div class="form-group">
            <label>Multiplier</label>
            <input
              v-model.number="localSettings.formulaMultiplier"
              type="number"
              step="0.01"
              min="0"
              max="1"
              class="form-control"
            />
            <p class="help-text">After-fee percentage (0.84 = 84%)</p>
          </div>
        </div>

        <div class="example-box">
          <h4>📝 Example Calculation</h4>
          <div class="example-content">
            <p>If Harga Marketplace = <strong>Rp 100,000</strong></p>
            <p>
              Expected = (100,000 - {{ localSettings.formulaDeduction }}) ×
              {{ localSettings.formulaMultiplier }}
            </p>
            <p class="result">
              = <strong>{{ formatPrice(calculateExample()) }}</strong>
            </p>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" @click="$emit('close')">Cancel</button>
        <button class="btn-save" @click="handleSave">💾 Save Settings</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from "vue";
import type { TiktokAnalyticsSettings } from "@/composables/useTiktokAnalytics";
import { getAuthHeaders } from "@/utils/apiHeaders";

const props = defineProps<{ settings: TiktokAnalyticsSettings }>();
const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", settings: TiktokAnalyticsSettings): void;
}>();

const localSettings = ref<TiktokAnalyticsSettings>({ ...props.settings });
const inventoryColumns = ref<string[]>([]);

async function fetchInventoryColumns() {
  try {
    const host = window.location.hostname;
    const protocol = window.location.protocol;
    const isLocalhost = host === "localhost" || host === "127.0.0.1";
    const apiUrl = isLocalhost
      ? `${protocol}//${host}:3000/api/inventory/stats`
      : `${protocol}//${host}/api/inventory/stats`;
    const response = await fetch(apiUrl, {
      headers: getAuthHeaders(),
    });
    if (response.ok) {
      const data = await response.json();
      if (data.success && data.data?.columns) {
        inventoryColumns.value = data.data.columns.map(
          (c: any) => c.column_name || c.name || c
        );
      }
    }
  } catch (error) {
    console.error("Failed to fetch inventory columns:", error);
  }
}

const priceColumns = computed(() => {
  // Show suggested price columns first, then all others
  const suggested = inventoryColumns.value.filter((name) => {
    const lower = name.toLowerCase();
    return (
      lower.includes("harga") ||
      lower.includes("price") ||
      lower.includes("cost") ||
      lower.includes("nilai")
    );
  });
  const others = inventoryColumns.value.filter(
    (name) => !suggested.includes(name)
  );
  return [...suggested, ...others];
});

function calculateExample(): number {
  return (
    (100000 - localSettings.value.formulaDeduction) *
    localSettings.value.formulaMultiplier
  );
}

function formatPrice(price: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
}

function handleSave() {
  emit("save", { ...localSettings.value });
}

watch(
  () => props.settings,
  (newSettings) => {
    localSettings.value = { ...newSettings };
  },
  { immediate: true }
);

onMounted(() => {
  fetchInventoryColumns();
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-lg);
  border-bottom: 1px solid var(--color-border);
}

.modal-header h2 {
  margin: 0;
  font-size: 1.1rem;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--color-text-muted);
}

.modal-body {
  padding: var(--space-lg);
}

.form-group {
  margin-bottom: var(--space-md);
}

.form-group label {
  display: block;
  margin-bottom: var(--space-xs);
  font-weight: 500;
  color: var(--color-text-secondary);
}

.form-control {
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

.help-text {
  margin: var(--space-xs) 0 0;
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.form-row {
  display: flex;
  gap: var(--space-md);
}

.form-row .form-group {
  flex: 1;
}

.formula-box {
  padding: var(--space-md);
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
}

.formula-box code {
  font-family: monospace;
  color: var(--color-primary);
}

.example-box {
  padding: var(--space-md);
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
  margin-top: var(--space-md);
}

.example-box h4 {
  margin: 0 0 var(--space-sm);
  font-size: 0.9rem;
}

.example-content p {
  margin: var(--space-xs) 0;
  font-size: 0.85rem;
}

.example-content .result {
  color: var(--color-success);
  font-size: 1rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-sm);
  padding: var(--space-md) var(--space-lg);
  border-top: 1px solid var(--color-border);
}

.btn-cancel,
.btn-save {
  padding: var(--space-sm) var(--space-lg);
  border: none;
  border-radius: var(--radius-md);
  font-weight: 500;
  cursor: pointer;
}

.btn-cancel {
  background: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

.btn-save {
  background: var(--color-primary);
  color: var(--color-text-inverse);
}
</style>
